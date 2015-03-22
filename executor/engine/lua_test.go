package engine_test

import (
	"bytes"
	"strings"
	"testing"

	"git.loadtests.me/loadtests/loadtests/executor/engine"
	"golang.org/x/net/context"
)

func TestLuaEngine(t *testing.T) {
	buf := bytes.NewBuffer(nil)

	want := `{"lvl":"info","step":"first_step","msg":"hello world"}
{"lvl":"fatal","step":"second_step","msg":"oh you're still there"}
`

	script := strings.NewReader(`
step.first_step = function()
    info("hello world")
end

step.second_step = function()
    fatal("oh you're still there")
end

step.first_step = function()
    info("hello world")
end
`)
	prgm, err := engine.Lua(script, buf)
	if err != nil {
		t.Fatal(err)
	}
	err = prgm.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	got := buf.String()
	if want != got {
		t.Logf("want=%q", want)
		t.Logf(" got=%q", got)
		t.Fatalf("different output")
	}
}
