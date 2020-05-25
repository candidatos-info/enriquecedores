package cce

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo"
)

const (
	testFile = "files.zip"
)

func TestPost(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadFile(testFile)
		if err != nil {
			t.Errorf("want error of opening file test nil, got %q", err)
		}
		w.Write(bytes)
	}))
	defer ts.Close()
	e := echo.New()
	// point the hostURL to the server test host
	hostURL = fmt.Sprintf("%s/%s", ts.URL, "%d")
	testCases := []struct {
		year int64
	}{
		{2016},
	}
	for _, tt := range testCases {
		in := dispatchRequest{
			Year: tt.year,
		}
		inAsBytes, err := json.Marshal(in)
		if err != nil {
			t.Errorf("failed to marshal in into bytes, want error nil, got: %q", err)
		}
		req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(string(inAsBytes)))
		if err != nil {
			t.Errorf("failed to create test request, expect error nil, got %q", err)
		}
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		cceHandler := NewHandler()
		cceHandler.Post(c)
		res := rec.Result()
		defer res.Body.Close()
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
