package engine_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lgpeterson/loadtests/executor/engine"
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
	prgm, err := engine.Lua(script, engine.SetLogger(buf))
	if err != nil {
		t.Log(buf.String())
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

func TestLuaHTTPBinding(t *testing.T) {

	want := `{"hello":"world"}`
	wantHeaderValue := "lolololol"
	var got string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			w.Header().Set("Lol-Lol-Lol-Lol", wantHeaderValue)
			w.Write([]byte(want))
		case "POST":
			p, err := ioutil.ReadAll(r.Body)
			if err != nil {
				t.Fatal(err)
			} else {
				got = string(p)
			}
		}
	}))
	defer srv.Close()

	script := strings.NewReader(fmt.Sprintf(`
step.first_step = function()
	resp = get(%q)
	info(resp.header['Lol-Lol-Lol-Lol'])
	post(%q, "application/json", resp.body)
end
`, srv.URL, srv.URL))

	buf := bytes.NewBuffer(nil)
	prgm, err := engine.Lua(script, engine.SetLogger(buf))
	if err != nil {
		t.Fatal(err)
	}
	err = prgm.Execute(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if want != got {
		t.Logf("want=%q", want)
		t.Logf(" got=%q", got)
		t.Fatalf("different output")
	}

	wantLog := `{"lvl":"info","step":"first_step","msg":"` + wantHeaderValue + `"}` + "\n"
	gotLog := buf.String()
	if wantLog != gotLog {
		t.Logf("want=%q", wantLog)
		t.Logf(" got=%q", gotLog)
		t.Fatalf("different output")
	}
}
