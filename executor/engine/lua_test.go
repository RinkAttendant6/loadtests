package engine_test

import (
	"bytes"
	"testing"

	"git.loadtests.me/loadtests/loadtests/executor/engine"
	"golang.org/x/net/context"
)

func TestLuaEngine(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	script := `info("hello world")`
	prgm, err := engine.Lua(script, buf)
	if err != nil {
		t.Fatal(err)
	}
	err := prgm.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want, got := "hello world", buf.String(); want != got {
		t.Fatalf("want output %q, got %q", want, got)
	}
}
