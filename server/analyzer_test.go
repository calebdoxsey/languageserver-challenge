package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestAnalyzer(t *testing.T) {
	// basic test but shows how we could flesh it out better

	srcdir := filepath.Join(os.TempDir(), fmt.Sprintf("languageserver-challenge-tests-%d", time.Now().Unix()), "src")
	defer os.RemoveAll(srcdir)

	os.MkdirAll(filepath.Join(srcdir, "example"), 0755)

	name := filepath.Join(srcdir, "example", "example.go")
	err := ioutil.WriteFile(name, []byte(`package main

import "fmt"

func main() {
	fmt.Println("Hello World")
}
`), 0755)
	if err != nil {
		t.Error("failed to write file: ", err)
	}

	a := newAnalyzer()
	pos, err := a.findDefinition(name, nil, 6, 12)
	if err != nil {
		t.Error("expected error to be nil but got", err)
	}

	if !strings.HasSuffix(pos.String(), "print.go:263:6") {
		t.Error("expected to find print.go:263:6, got", pos.String())
	}
}
