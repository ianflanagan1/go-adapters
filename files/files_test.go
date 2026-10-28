package files_test

import (
	"testing"

	"github.com/ianflanagan1/go-adapters/files"
)

func openDir(t *testing.T) *files.Dir {
	t.Helper()
	dir, err := files.Open(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := dir.Close(); err != nil {
			t.Error(err)
		}
	})
	return dir
}

func TestAWrittenFileReadsBack(t *testing.T) {
	dir := openDir(t)
	if err := dir.WriteFile("a.txt", []byte("hello")); err != nil {
		t.Fatal(err)
	}
	if got, err := dir.ReadFile("a.txt"); err != nil || string(got) != "hello" {
		t.Fatalf("got %q, %v; want \"hello\"", got, err)
	}
}

func TestPathsCantEscapeTheDirectory(t *testing.T) {
	dir := openDir(t)
	if err := dir.WriteFile("../escaped.txt", []byte("x")); err == nil {
		t.Fatal("wrote outside the directory")
	}
	if _, err := dir.ReadFile("../escaped.txt"); err == nil {
		t.Fatal("read outside the directory")
	}
}

func TestAMissingDirectoryIsAnError(t *testing.T) {
	if _, err := files.Open(t.TempDir() + "/missing"); err == nil {
		t.Fatal("got no error")
	}
}
