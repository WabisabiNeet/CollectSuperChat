package server_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/server"
)

func Test(t *testing.T) {
	s := server.NewServer(func(channel string) {
	})

	s.Serve()
	t.Fatal("")
}

func Test2(t *testing.T) {
	dir, err := ioutil.TempDir("", "target")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		os.RemoveAll(dir)
	}()
	lockfile := path.Join(dir, "aaa")
	f, err := os.Create(lockfile)
	defer f.Close()
	if err != nil {
		t.Fatal("unexpected error.")
	}

	_, err = os.Create(lockfile)
	if !os.IsExist(err) {
		t.Fatal(err)
	}
}
