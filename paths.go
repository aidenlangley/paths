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

// Provides shortcuts to [os.UserHomeDir], [os.UserConfigDir]
// & [os.UserCacheDir], the ability to [Resolve] a path, and a way to store
// [os.Stat] info as [FileInfo] via [Open].
package paths

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

var home, config, cache string

// Given a pointer to a string, check if it's empty, if not then populate it
// with the return value of the given function (unless it errors, in which case
// [log.Fatal]).
func checkAndGetString(s *string, f func() (string, error)) {
	if *s != "" {
		return
	}

	var err error
	*s, err = f()

	if err != nil {
		log.Fatalf("paths %s", err)
	}
}

// Returns [os.UserHomeDir], but the value is cached locally to avoid
// unnecessary repeat calls.
func Home() string {
	checkAndGetString(&home, os.UserHomeDir)
	return home
}

// Returns [os.UserConfigDir], but the value is cached locally to avoid
// unnecessary repeat calls.
func Config() string {
	checkAndGetString(&config, os.UserConfigDir)
	return config
}

// Returns [os.UserCacheDir], but the value is cached locally to avoid
// unnecessary repeat calls.
func Cache() string {
	checkAndGetString(&cache, os.UserCacheDir)
	return cache
}

// Check if a file or directory exists at path.
func Exists(path string) bool {
	_, err := os.Stat(path)
	return !errors.Is(err, os.ErrNotExist)
}

// Attempt to resolve a path by checking first if it exists, otherwise make it
// absolute and then check again. Returns [fs.ErrNotExist] on failure.
func Resolve(path string) (string, error) {
	if Exists(path) {
		return path, nil
	}

	if abs, err := filepath.Abs(path); err != nil {
		return path, err
	} else if Exists(abs) {
		return abs, nil
	}

	return path, fs.ErrNotExist
}

// For more complex operations, this struct is provided to facilitate options
// such as considering `$HOME` when resolving.
type PathResolver struct {
	Path          string
	ResolveToHome bool
}

// Create a new [*PathResolver].
//
//	r := paths.NewPathResolver(path, paths.ResolveToHome())
//	path, err := r.Resolve()
func NewPathResolver(path string, options ...ResolverOption) *PathResolver {
	resolver := &PathResolver{Path: path}
	for _, option := range options {
		option(resolver)
	}
	return resolver
}

// ResolverOption is a function that accepts a [*PathResolver] and applies some
// configuration.
type ResolverOption func(*PathResolver)

// Pass this to [NewPathResolver] to instruct [PathResolver] to consider `$HOME`
// when resolving a path.
func ResolveToHome() ResolverOption {
	return func(resolver *PathResolver) {
		resolver.ResolveToHome = true
	}
}

// Calls [Resolve], but on error, it will consider its [ResolverOption]s and
// act appropriately.
func (resolver *PathResolver) Resolve() (string, error) {
	// Resolve without consider `$HOME`.
	path, err := Resolve(resolver.Path)

	// If there are no errors, we can just return, we've got our path.
	if err == nil {
		return path, err
	}

	// Otherwise, the file didn't exist, and we want to ResolveToHome, so go
	// again.
	if resolver.ResolveToHome {
		path, err = Resolve(filepath.Join(Home(), resolver.Path))
	}

	return path, err
}

// Error returned by [Open], contains errors returned by [os.Open] & [os.Stat]
// so you may utilise [errors.Is] like so...
//
//	 fi, err := paths.Open(path)
//		 if errors.Is(err, os.PathError) {
//		   ...
//	 }
type FileInfoError struct {
	path string
	err  error
}

func (e *FileInfoError) Error() string {
	return fmt.Sprintf("paths %s %s", e.path, e.err)
}

func (e *FileInfoError) Unwrap() error {
	return e.err
}

// Stores information on a file.
type FileInfo struct {
	Path      string    `json:"path"`
	IsRegular bool      `json:"is_regular"`
	IsDir     bool      `json:"is_dir"`
	Name      string    `json:"filename"`
	Size      int64     `json:"size"`
	Modified  time.Time `json:"modified"`
}

// Determines if this [*FileInfo] is the same as other by comparing
// name, size & modified [time.Time].
func (fi *FileInfo) Equals(other *FileInfo) bool {
	return fi.Name == other.Name &&
		fi.Size == other.Size &&
		fi.Modified == other.Modified
}

// Determines if this [*FileInfo] is newer than other.
func (fi *FileInfo) Newer(other *FileInfo) bool {
	return fi.Modified.After(other.Modified)
}

// Return terminal friendly string.
func (fi FileInfo) String() string {
	return fi.Path
}

// Calls [os.Open] and then [os.File.Stat] to populate [FileInfo].
func Open(p string) (FileInfo, error) {
	file, err := os.Open(p)
	if err != nil {
		return FileInfo{}, &FileInfoError{path: p, err: err}
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return FileInfo{}, &FileInfoError{path: p, err: err}
	}

	isRegular := info.Mode().IsRegular()
	isDir := info.IsDir()
	name := info.Name()
	if isDir {
		name = name + "/"
	}

	return FileInfo{
		Path:      p,
		Name:      name,
		IsRegular: isRegular,
		IsDir:     isDir,
		Size:      info.Size(),
		Modified:  info.ModTime(),
	}, nil
}

// A [Path] is just a string that can be operated on.
//
// All operations are platform agnostic.
type Path string

// Check if this [*Path] exists on the system.
func (p *Path) Exists() bool {
	return Exists(p.String())
}

// Calls [Resolve] on this [*Path].
func (p *Path) Resolve() (string, error) {
	return Resolve(p.String())
}

// Return [Path] as a string.
func (p Path) String() string {
	return string(p)
}
