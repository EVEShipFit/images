package res

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	raw := `res:/UI/Texture/Icons/bpo.png,3e/3e9a564a6a3edb96_45518159d7ae1ce0cb4e80242e22b44b,45518159d7ae1ce0cb4e80242e22b44b,12109,12015

res:/intromovie.txt,a9/a9d1721dd5cc6d54_e6bbb2df307e5a9527159a4c971034b5,e6bbb2df307e5a9527159a4c971034b5,9719,3312
`
	index, err := Parse(strings.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	if len(index) != 2 {
		t.Errorf("entries = %d, want 2", len(index))
	}
	entry, ok := index["res:/ui/texture/icons/bpo.png"]
	if !ok {
		t.Fatal("resource paths are not lowercased")
	}
	want := Entry{Path: "3e/3e9a564a6a3edb96_45518159d7ae1ce0cb4e80242e22b44b", MD5: "45518159d7ae1ce0cb4e80242e22b44b"}
	if entry != want {
		t.Errorf("entry = %+v, want %+v", entry, want)
	}
}

func TestParseMalformed(t *testing.T) {
	if _, err := Parse(strings.NewReader("res:/a.png,only-a-path\n")); err == nil {
		t.Error("expected an error")
	}
}

func TestGetDoesNotRetryNotFound(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		http.NotFound(w, r)
	}))
	defer server.Close()

	if _, err := get(server.URL); err == nil {
		t.Fatal("expected an error")
	}
	if requests != 1 {
		t.Errorf("requests = %d, want 1", requests)
	}
}
