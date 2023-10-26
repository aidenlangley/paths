package paths

import (
	"os"
	"testing"
	"time"
)

func resolve(t *testing.T, p string) string {
	if _, err := Resolve(p); err != nil {
		t.Errorf("ERR resolve %v %v", p, err)
	}
	return p
}

func TestResolveLocal(t *testing.T) {
	_ = resolve(t, "../main.go")
}

func TestResolveHome(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("ERR home %v", err)
	}
	_ = resolve(t, home)
}

func TestResolveSystem(t *testing.T) {
	sep := os.PathSeparator
	if sep == '/' {
		_ = resolve(t, "/home")
	} else if sep == '\'' {
		_ = resolve(t, "C:\\Windows")
	}
}

func openHome(t *testing.T) FileInfo {
	var err error
	info, err := Open(Home())
	if err != nil {
		t.Errorf("ERR %s", err)
	}
	return info
}

func TestOpenHome(t *testing.T) {
	_ = openHome(t)
}

func TestFileInfoEquals(t *testing.T) {
	info := openHome(t)
	if other := info; !info.Equals(&other) {
		t.Errorf("ERR %s != %s", info, other)
	}
}

func TestFileInfoNewer(t *testing.T) {
	info := openHome(t)
	other := info

	manyYearsAgo := time.Now().AddDate(-42, 0, 0)
	other.Modified = manyYearsAgo

	if !info.Newer(&other) {
		t.Errorf("ERR %s > %s", info.Modified.String(), other.Modified.String())
	}
}
