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

	"github.com/candidatos-info/enriquecedores/status"
	"github.com/labstack/echo"
)

const (
	testFile = "files.zip"
)

func fakeServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := ioutil.ReadFile(testFile)
		if err != nil {
			t.Errorf("want error of opening file test nil, got %q", err)
		}
		w.Write(bytes)
	}))
}

func TestPost(t *testing.T) {
	ts := fakeServer(t)
	defer ts.Close()
	fakeURLString := ts.URL + "/%d"
	cceHandler := New(fakeURLString, ".")
	e := echo.New()
	in := postRequest{
		Year: 2016,
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
	cceHandler.post(c)
	res := rec.Result()
	defer res.Body.Close()
	expectedGeneratedFileName := fmt.Sprintf("cce_sheets_%d.zip", in.Year)
	_, err = ioutil.ReadFile(expectedGeneratedFileName)
	if err != nil {
		t.Errorf("failed to read the expected file, got err %q", err)
	}
}

func TestInvalidPostRequest(t *testing.T) {
	ts := fakeServer(t)
	defer ts.Close()
	fakeURLString := ts.URL + "/%d"
	cceHandler := New(fakeURLString, ".")
	e := echo.New()
	req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader("INVALID REQUEST BODY"))
	if err != nil {
		t.Errorf("failed to create test request, expect error nil, got %q", err)
	}
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	cceHandler.post(c)
	res := rec.Result()
	defer res.Body.Close()
	if cceHandler.status != status.Idle {
		t.Errorf("expected status to stay Idle when sending invalid request body, got %d", status.Idle)
	}
}

func TestDownload(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "string")
	}))
	defer ts.Close()
	var buf bytes.Buffer
	b, err := donwloadFile(ts.URL, &buf)
	if err != nil {
		t.Errorf("want error nil, got %q", err)
	}
	if buf.String() != "string" {
		t.Errorf("want %s, got %s", "string", buf.String())
	}
	if b == nil {
		t.Errorf("exepcted buf different of nil")
	}
}
