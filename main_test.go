package main

import (
	"os"
	"testing"
)

func TestCreateArchiveDir(t *testing.T) {
	dst, err := createArchiveDir()
	if err != nil {
		t.Fatal(err.Error())
	}

	stat, err := os.Stat(dst)
	if os.IsNotExist(err) {
		t.Fatalf("%s does not exist", dst)
	}

	if !stat.IsDir() {
		t.Fatalf("%s is not a folder", dst)
	}

	err = os.Remove(dst)
	if err != nil {
		t.Fatalf("Failed removing %s: %s", dst, err.Error())
	}
}
