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
	testFile = "file_2016.txt"
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
	path, err := os.Getwd()
	if err != nil {
		t.Errorf("expected to have err nil when getting current directory, got %q", err)
	}
	year := 2016
	sourceURL := fmt.Sprintf("file://%s/files_%d.zip", path, year)
	cceHandler := New(sourceURL, ".")
	cceHandler.post()
	expectedGeneratedFileName := fmt.Sprintf("cce_files_%d.zip", year)
	_, err = ioutil.ReadFile(expectedGeneratedFileName)
	if err != nil {
		t.Errorf("failed to read the expected file, got err %q", err)
	}
}

func TestRemoveDuplicates(t *testing.T) {
	candidates := []*Candidatura{
		&Candidatura{
			Legislatura:       2016,
			Cargo:             "PREFEITO",
			UF:                "AL",
			Municipio:         "Maceio",
			NumeroUrna:        505,
			NomeUrna:          "Lelinho",
			Aptidao:           "APTO",
			Deferimento:       "DEFERIDO",
			TipoAgremiacao:    "AGREMIAÇÃO",
			NumeroPartido:     5,
			LegendaPartido:    "AM",
			NomePartido:       "Arctic Monkeys",
			NomeColigacao:     "Arabella",
			PartidosColigacao: "Beatles / AM / Zeppelin",
			DeclarouBens:      "S",
			Situacao:          "SEGUNDO-TURNO",
			Turno:             1,
			Candidato: Candidato{
				CPF:        "07496470430",
				Nascimento: "29/11/1997",
			},
		},
		&Candidatura{
			Legislatura:       2016,
			Cargo:             "PREFEITO",
			UF:                "AL",
			Municipio:         "Maceio",
			NumeroUrna:        505,
			NomeUrna:          "Lelinho",
			Aptidao:           "APTO",
			Deferimento:       "DEFERIDO",
			TipoAgremiacao:    "AGREMIAÇÃO",
			NumeroPartido:     5,
			LegendaPartido:    "AM",
			NomePartido:       "Arctic Monkeys",
			NomeColigacao:     "Arabella",
			PartidosColigacao: "Beatles / AM / Zeppelin",
			DeclarouBens:      "S",
			Situacao:          "ELEITO",
			Turno:             2,
			Candidato: Candidato{
				CPF:        "07496470430",
				Nascimento: "29/11/1997",
			},
		},
		&Candidatura{
			Legislatura:       2016,
			Cargo:             "PREFEITO",
			UF:                "AL",
			Municipio:         "Capela",
			NumeroUrna:        45666,
			NomeUrna:          "Marcelinho",
			Aptidao:           "APTO",
			Deferimento:       "DEFERIDO",
			TipoAgremiacao:    "AGREMIAÇÃO",
			NumeroPartido:     45,
			LegendaPartido:    "AM",
			NomePartido:       "Arctic Monkeys",
			NomeColigacao:     "Arabella",
			PartidosColigacao: "Beatles / AM / Zeppelin",
			DeclarouBens:      "S",
			Situacao:          "ELEITO",
			Turno:             1,
			Candidato: Candidato{
				CPF:        "07496470431",
				Nascimento: "29/11/1998",
			},
		},
	}
	c, err := removeDuplicates(candidates)
	if err != nil {
		t.Errorf("expected to have err nil when removing duplicated, got %q", err)
	}
	if len(c) != 2 {
		t.Errorf("expected to have only two candidates after removing duplicates, got %d", len(c))
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
	f, err := os.Open("files_2016.zip")
	if err != nil {
		t.Errorf("expected err to be nil when opening test file")
	}
	bytes, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("expected err to be nil when reading file")
	}
	h, err := hash(bytes)
	if err != nil {
		t.Errorf("expected err to be nil")
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
		t.Errorf("expected buf different of nil")
	}
}

func TestUnzipDownloadedFiles(t *testing.T) {
	fileToUnzip := "files_2016.zip"
	file, err := os.Open(fileToUnzip)
	if err != nil {
		t.Errorf("expected to have err nil when opening sample zip files, got %q", err)
	}
	defer file.Close()
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		t.Errorf("expected to have err nil when reading bytes of sample file, got %q", err)
	}
	unzipDestination, err := ioutil.TempDir("", "unzipped")
	if err != nil {
		t.Errorf("expected err nil when creating temporary dir, got %q", err)
	}
	files, err := unzipDownloadedFiles(fileBytes, unzipDestination)
	if err != nil {
		t.Errorf("failed to unzip sample files, got %q", err)
	}
	if len(files) != 6 {
		t.Errorf("expected to have 6 csv files decompressed files, got %d", len(files))
	}
	if err = os.RemoveAll(unzipDestination); err != nil {
		t.Errorf("expected to have err nil when removing temporary dir, got %q", err)
	}
}
