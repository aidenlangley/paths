package paths_test

import (
	"log"
	"os"
	"path"
	"runtime"
	"testing"

	"git.sr.ht/~nedia/paths"
)

type Platform int

const (
	Linux Platform = iota
	MacOS
	Windows
)

var (
	homeDir  string
	tmpDir   = os.TempDir()
	platform Platform
)

func init() {
	var err error
	if homeDir, err = os.UserHomeDir(); err != nil {
		log.Fatalf("ERR homeDir err=%v", err)
	}

	if runtime.GOOS == "windows" {
		platform = Windows
	} else if os.PathSeparator == '/' {
		platform = Linux | MacOS
	}
}

func home(s string) string {
	return path.Join(homeDir, s)
}

func tmp(s string) string {
	return path.Join(tmpDir, s)
}

func simpleResolve(t *testing.T, s string) {
	resolved, err := paths.Resolve(s)
	if err != nil {
		t.Errorf("ERR simpleResolve err=%v", err)
	}
	t.Logf("LOG simpleResolve s=%v resolved=%v", s, resolved)
}

func TestSimpleResolve(t *testing.T) {
	// Resolve local to this code.
	simpleResolve(t, "paths.go")
	simpleResolve(t, "paths_test.go")

	// Resolve path relative to home.
	simpleResolve(t, home("Downloads/"))
	simpleResolve(t, home(".config/nvim"))

	// Resolve system paths.
	switch platform {
	case Windows:
		simpleResolve(t, "C:\\Windows")
	case Linux | MacOS:
		simpleResolve(t, tmp(""))
		simpleResolve(t, "/etc/")
	}
}

func newPath(t *testing.T, s string, r *paths.Resolver) *paths.Path {
	p := paths.New(s).WithResolver(r)
	t.Logf("LOG newPath s=%v r=%v", s, r)
	return p
}

func pathResolve(t *testing.T, s string, r *paths.Resolver) {
	p := newPath(t, s, r)
	resolved, err := p.Resolve()
	if err != nil {
		t.Errorf("ERR pathResolve err=%v", err)
	}
	t.Logf("LOG pathResolve p=%v resolved=%v", p, resolved)
}

func TestPathResolve(t *testing.T) {
	// Resolve local to this code.
	pathResolve(t, "paths.go", nil)
	pathResolve(t, "paths_test.go", nil)

	// Resolve path relative to home.
	pathResolve(t, home("Downloads"), nil)
	pathResolve(t, home(".config/nvim"), nil)

	// Resolve path relative to home with [paths.ResolveToHome].
	r := paths.NewResolver(paths.ResolveToHome())
	pathResolve(t, "Downloads", r)
	pathResolve(t, ".config/nvim", r)

	// Resolve system paths.
	switch platform {
	case Windows:
		pathResolve(t, "C:\\Windows", nil)
	case Linux | MacOS:
		pathResolve(t, tmp(""), nil)
		pathResolve(t, "/etc", nil)
	}
}

func deleteFile(t *testing.T, p *paths.Path) {
	if p.Exists() {
		if err := p.Delete(); err != nil {
			t.Errorf("ERR deleteFile err=%v", err)
		}
		t.Logf("LOG deleteFile p=%v", p)
	}
}

func createFile(t *testing.T, p *paths.Path) {
	if _, err := p.Create(); err != nil {
		t.Errorf("ERR createFile err=%v", err)
	}
	t.Logf("LOG createFile p=%v modified=%v", p, p.Modified())
}

func TestTimeSinceModified(t *testing.T) {
	p := newPath(t, tmp("paths_tests.go__TestTimeSinceModified__p.tmp"), nil)
	createFile(t, p)

	if p.TimeSinceModified() < 0 {
		t.Fail()
	}

	deleteFile(t, p)
}

func TestNewer(t *testing.T) {
	p := newPath(t, tmp("paths_test.go__TestNewer__p.tmp"), nil)
	other := newPath(t, tmp("paths_test.go__TestNewer__other.tmp"), nil)

	createFile(t, p)
	createFile(t, other)

	// p is not newer than other since we create p first, and other 2nd, so this
	// passes. They pretty much get made at the same time though.
	if p.Newer(other) {
		t.Fail()
	}

	deleteFile(t, p)
	deleteFile(t, other)
}
