package cce

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
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
	in := postRequest{
		Year: 2016,
	}
	cceHandler := New(ts.URL+"/%d", ".")
	cceHandler.post(&in)
	expectedGeneratedFileName := fmt.Sprintf("cce_sheets_%d.zip", in.Year)
	_, err := ioutil.ReadFile(expectedGeneratedFileName)
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
	cceHandler.Post(c)
	res := rec.Result()
	defer res.Body.Close()
	if cceHandler.Status != status.Idle {
		t.Errorf("expected status to stay Idle when sending invalid request body, got %d", status.Idle)
	}
}

func TestHash(t *testing.T) {
	f, err := os.Open("files.zip")
	if err != nil {
		t.Errorf("exepcted err to be nil when opening test file")
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("exepcted err to be nil when reading file")
	}
	h, err := hash(bytes)
	if err != nil {
		t.Errorf("exepcted err to be nil")
	}
	if h != "ed90a292c8a264622d726c2d17650d27" {
		t.Errorf("want hash ed90a292c8a264622d726c2d17650d27, got %s", h)
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
