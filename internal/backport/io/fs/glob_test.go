// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs_test

import (
	"testing"

	. "golang.org/x/website/internal/backport/io/fs"
	"golang.org/x/website/internal/backport/osfs"
	"golang.org/x/website/internal/backport/path"
)

var globTests = []struct {
	fs              FS
	pattern, result string
}{
	{osfs.DirFS("."), "glob.go", "glob.go"},
	{osfs.DirFS("."), "gl?b.go", "glob.go"},
	{osfs.DirFS("."), `gl\ob.go`, "glob.go"},
	{osfs.DirFS("."), "*", "glob.go"},
	{osfs.DirFS(".."), "*/glob.go", "fs/glob.go"},
}

func TestGlob(t *testing.T) {
	for _, tt := range globTests {
		matches, err := Glob(tt.fs, tt.pattern)
		if err != nil {
			t.Errorf("Glob error for %q: %s", tt.pattern, err)
			continue
		}
		if !contains(matches, tt.result) {
			t.Errorf("Glob(%#q) = %#v want %v", tt.pattern, matches, tt.result)
		}
	}
	for _, pattern := range []string{"no_match", "../*/no_match", `\*`} {
		matches, err := Glob(osfs.DirFS("."), pattern)
		if err != nil {
			t.Errorf("Glob error for %q: %s", pattern, err)
			continue
		}
		if len(matches) != 0 {
			t.Errorf("Glob(%#q) = %#v want []", pattern, matches)
		}
	}
}

func TestGlobError(t *testing.T) {
	bad := []string{`[]`, `nonexist/[]`}
	for _, pattern := range bad {
		_, err := Glob(osfs.DirFS("."), pattern)
		if err != path.ErrBadPattern {
			t.Errorf("Glob(fs, %#q) returned err=%v, want path.ErrBadPattern", pattern, err)
		}
	}
}

// contains reports whether vector contains the string s.
func contains(vector []string, s string) bool {
	for _, elem := range vector {
		if elem == s {
			return true
		}
	}
	return false
}

type globOnly struct{ GlobFS }

func (globOnly) Open(name string) (File, error) { return nil, ErrNotExist }

func TestGlobMethod(t *testing.T) {
	check := func(desc string, names []string, err error) {
		t.Helper()
		if err != nil || len(names) != 1 || names[0] != "hello.txt" {
			t.Errorf("Glob(%s) = %v, %v, want %v, nil", desc, names, err, []string{"hello.txt"})
		}
	}

	// Test that ReadDir uses the method when present.
	names, err := Glob(globOnly{testFsys}, "*.txt")
	check("readDirOnly", names, err)

	// Test that ReadDir uses Open when the method is not present.
	names, err = Glob(openOnly{testFsys}, "*.txt")
	check("openOnly", names, err)
}
