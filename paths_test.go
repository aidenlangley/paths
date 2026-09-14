package paths_test

import (
	"log"
	"os"
	"path"
	"runtime"
	"testing"

	"git.sr.ht/~nedia/paths"
)

var (
	err      error
	homeDir  string
	resolver *paths.Resolver
)

func init() {
	homeDir, err = os.UserHomeDir()
	if err != nil {
		log.Fatalf("ERR init %v", err)
	}

	resolver = paths.NewResolver()
}

func TestSimpleResolve(t *testing.T) {
	resolve := func(s string) {
		_, err := paths.Resolve(s)
		if err != nil {
			t.Errorf("ERR resolve %v %v", s, err)
		}
	}

	// Resolve local to this code.
	resolve("paths.go")

	// Resolve path relative to home.
	resolve(path.Join(homeDir, "Work/go_paths"))

	// Resolve system paths.
	if runtime.GOOS == "windows" {
		resolve("C:\\Windows")
	} else if os.PathSeparator == '/' {
		resolve("/tmp/")
		resolve("/etc/")
	}
}

func TestPathResolve(t *testing.T) {
	resolve := func(p *paths.Path) {
		_, err := p.Resolve()
		if err != nil {
			t.Errorf("ERR resolve %v %v", p.FileName(), err)
		}
	}

	// Resolve local to this code.
	resolve(paths.New("paths.go"))

	// Resolve path relative to home.
	resolve(paths.New(path.Join(homeDir, "Work/go_paths")))

	// Resolve path relative to home with [paths.ResolveToHome].
	r := paths.NewResolver(paths.ResolveToHome())
	resolve(paths.New("go").WithResolver(r))
	resolve(paths.New(".config/nvim").WithResolver(r))

	// Resolve system paths.
	if runtime.GOOS == "windows" {
		resolve(paths.New("C:\\Windows"))
	} else if os.PathSeparator == '/' {
		resolve(paths.New("/tmp/"))
		resolve(paths.New("/etc/"))
	}
}
