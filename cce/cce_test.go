package cce

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const (
	testFile = "files.zip"
)

func TestUnzip(t *testing.T) {
	testCases := []struct {
		in string
	}{
		{"files.zip"},
	}
	for _, tt := range testCases {
		output := strings.Split(tt.in, ".zip")[0]
		unzip(tt.in, output)
		fi, err := os.Stat(output)
		if err != nil {
			fmt.Println(err)
			return
		}
		if !fi.Mode().IsDir() {
			t.Errorf("expected to have a file after unzip")
		}
	}
}

func TestDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "string")
	}))
	defer ts.Close()
	var buf bytes.Buffer
	err := donwloadFile(ts.URL, &buf)
	if err != nil {
		t.Errorf("want error nil, got %q", err)
	}
	if buf.String() != "string" {
		t.Errorf("want %s, got %s", "string", buf.String())
	}
}
