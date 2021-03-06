package gzip

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestGZIPFilesHaveZeroMTimes checks that every .gz file in the tree
// has a zero MTIME. This is a requirement for the Debian maintainers
// to be able to have deterministic packages.
//
// See https://golang.org/issue/14937.
func TestGZIPFilesHaveZeroMTimes(t *testing.T) {
	goroot, err := filepath.EvalSymlinks(runtime.GOROOT())
	if err != nil {
		t.Fatal("error evaluating GOROOT: ", err)
	}
	var files []string
	err = filepath.Walk(goroot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".gz") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			t.Skipf("skipping: GOROOT directory not found: %s", runtime.GOROOT())
		}
		t.Fatal("error collecting list of .gz files in GOROOT: ", err)
	}
	if len(files) == 0 {
		t.Fatal("expected to find some .gz files under GOROOT")
	}
	for _, path := range files {
		checkZeroMTime(t, path)
	}
}

func checkZeroMTime(t *testing.T, path string) {
	f, err := os.Open(path)
	if err != nil {
		t.Error(err)
		return
	}
	defer f.Close()
	gz, err := NewReader(f)
	if err != nil {
		t.Errorf("cannot read gzip file %s: %s", path, err)
		return
	}
	defer gz.Close()
	if !gz.ModTime.IsZero() {
		t.Errorf("gzip file %s has non-zero mtime (%s)", path, gz.ModTime)
	}
}
