// Copyright (c) 2021 Aiden Langley
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package paths provides shortcuts the ability to [Resolve] a path, and a way
// to store [os.Stat] info.
package paths

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"time"
)

func Exists(s string) bool {
	_, err := os.Stat(s)
	return !errors.Is(err, os.ErrNotExist)
}

// Absolute turns a relative path into an absolute one via [filepath.Abs].
func Absolute(s string) (string, error) {
	return filepath.Abs(s)
}

// Resolve attempts to resolve a path by checking first if it exists,
// otherwise make it absolute and then check again. Returns [fs.ErrNotExist] on
// failure.
func Resolve(s string) (string, error) {
	if Exists(s) {
		return s, nil
	}

	abs, err := Absolute(s)
	if err != nil {
		return s, err
	}

	if Exists(abs) {
		return abs, nil
	}

	return s, fs.ErrNotExist
}

// Path holds a string, can be configured to use a [Resolver], and stores the
// file, plus metadata, after file operations.
type Path struct {
	Path string

	resolver *Resolver

	File     *os.File
	FileInfo fs.FileInfo
}

func New(s string) *Path {
	return &Path{
		Path:     s,
		resolver: &Resolver{},
	}
}

func From(s string) *Path {
	return New(s)
}

// WithResolver will add a [Resolver] to this [Path]. Builder pattern style.
func (p *Path) WithResolver(r *Resolver) *Path {
	p.resolver = r
	return p
}

func (p *Path) Exists() bool {
	return Exists(p.String())
}

func (p *Path) Absolute() (string, error) {
	return Absolute(p.Path)
}

// Resolve just calls [Resolver.Resolve], all the logic can be found there.
func (p *Path) Resolve() (string, error) {
	resolved, err := p.resolver.Resolve(p.Path)
	p.Path = resolved
	return p.Path, err
}

// Stat calls [os.File.Stat] and sets [Path.FileInfo], or returns an error.
func (p *Path) Stat() (fs.FileInfo, error) {
	fi, err := p.File.Stat()
	if err == nil {
		p.FileInfo = fi
	}

	return p.FileInfo, err
}

// GetFileInfo gets [os.FileInfo] and sets [Path.FileInfo], regardless of error.
func (p *Path) GetFileInfo() fs.FileInfo {
	_, _ = p.Stat()
	return p.FileInfo
}

func (p *Path) FileName() string {
	if p.FileInfo == nil {
		_ = p.GetFileInfo()
	}

	return p.FileInfo.Name()
}

func (p *Path) Open() (*os.File, error) {
	f, err := os.Open(p.String())
	if err != nil {
		return f, err
	}

	p.File = f
	_ = p.GetFileInfo()

	return p.File, err
}

func (p *Path) OpenFile(flag int, perm os.FileMode) (*os.File, error) {
	f, err := os.OpenFile(p.String(), flag, perm)
	if err != nil {
		return f, err
	}

	p.File = f
	_ = p.GetFileInfo()

	return f, err
}

func (p *Path) Create() (*os.File, error) {
	f, err := os.Create(p.String())
	if err != nil {
		return f, err
	}

	p.File = f
	_ = p.GetFileInfo()

	return f, err
}

func (p *Path) Write(b []byte) (int, error) {
	if p.File == nil {
		f, err := p.OpenFile(os.O_RDWR, 0o644)
		if err != nil {
			if f, err = p.Create(); err != nil {
				return 0, err
			}
		}

		p.File = f
		_ = p.GetFileInfo()
	}

	return p.File.Write(b)
}

// Equals determines if this [Path] is the same as other by comparing
// [fs.FileInfo] data such as: [fs.FileInfo.Name], [fs.FileInfo.Size],
// [fs.FileInfo.Mode] and [fs.FileInfo.ModTime].
func (p *Path) Equals(other *Path) bool {
	if p.FileInfo == nil || other.FileInfo == nil {
		return false
	}

	return p.FileInfo.Name() == other.FileInfo.Name() &&
		p.FileInfo.Size() == other.FileInfo.Size() &&
		p.FileInfo.Mode() == other.FileInfo.Mode() &&
		p.FileInfo.ModTime().Equal(other.FileInfo.ModTime()) &&
		p.FileInfo.IsDir() == other.FileInfo.IsDir()
}

// Newer determines if this [os.File] is newer than the other.
func (p *Path) Newer(other *Path) bool {
	if other.FileInfo == nil {
		return true
	}

	if p.FileInfo == nil {
		_ = p.GetFileInfo()
	}

	return p.FileInfo.ModTime().After(other.FileInfo.ModTime())
}

// TimeSinceModified returns the [time.Duration] since this [os.File] was last
// modified.
func (p *Path) TimeSinceModified() time.Duration {
	if p.FileInfo == nil {
		_ = p.GetFileInfo()
	}
	return time.Since(p.FileInfo.ModTime())
}

func (p *Path) String() string {
	return p.Path
}

// Resolver For more complex operations, this struct is provided to facilitate
// options such as considering `$HOME` when resolving.
type Resolver struct {
	ResolveToHome bool
}

// NewResolver creates a new [*Resolver].
//
//	r := NewResolver(path, ResolveToHome())
//	path, err := r.Resolve()
func NewResolver(options ...Option) *Resolver {
	resolver := &Resolver{}
	for _, option := range options {
		option(resolver)
	}
	return resolver
}

// Option is a function that accepts a [Resolver] and applies some
// configuration.
type Option func(*Resolver)

// ResolveToHome instructs [Resolver] to consider `$HOME` when resolving a path.
func ResolveToHome() Option {
	return func(resolver *Resolver) {
		resolver.ResolveToHome = true
	}
}

// Resolve calls [Resolve], but on error, it will consider its [Option]s and act
// appropriately.
func (r *Resolver) Resolve(s string) (string, error) {
	resolved, err := Resolve(s)
	if err != nil {
		// Hasn't worked, so we can try resolving to $HOME.
		if r.ResolveToHome {
			return resolveToHome(s)
		}
	}

	return resolved, err
}

// When resolving the path, it will assume it's been called from the $HOME dir,
// so we prepend $HOME to the path, and then try to resolve it again.
func resolveToHome(s string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return s, err
	}

	return Resolve(filepath.Join(home, s))
}
