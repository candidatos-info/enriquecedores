package cce

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
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

// func TestPost(t *testing.T) {
// 	path, err := os.Getwd()
// 	if err != nil {
// 		t.Errorf("expected to have err nil when getting current directory, got %q", err)
// 	}
// 	year := 2016
// 	sourceURL := fmt.Sprintf("file://%s/files_%d.zip", path, year)
// 	cceHandler := New(sourceURL, "baseDir")
// 	cceHandler.post(&postRequest{})
// 	expectedGeneratedFileName := fmt.Sprintf("cce_files_%d.zip", year)
// 	_, err = ioutil.ReadFile(expectedGeneratedFileName)
// 	if err != nil {
// 		t.Errorf("failed to read the expected file, got err %q", err)
// 	}
// 	err = os.Remove(expectedGeneratedFileName)
// 	if err != nil {
// 		t.Errorf("expected err nil when removing generated file, got %q", err)
// 	}
// 	err = os.Remove(fmt.Sprintf("cce_hash_files_%d.zip", year))
// 	if err != nil {
// 		t.Errorf("expected err nil when removing created hash file, got %q", err)
// 	}
// }

// func TestZipFile(t *testing.T) {
// 	testCases := []struct {
// 		fileName string
// 		zipName  string
// 		content  string
// 	}{
// 		{"file1.txt", "file1.zip", "Why'd you only call me when you're high?"},
// 		{"file2.txt", "file2.zip", "Do I wanna know"},
// 		{"file3.txt", "file3.zip", "Yesterday"},
// 	}
// 	for _, tt := range testCases {
// 		err := zipFile([]byte(tt.content), tt.zipName, tt.fileName)
// 		if err != nil {
// 			t.Errorf("expected to have err nil when compressing test files, got %q", err)
// 		}
// 	}
// 	for _, tt := range testCases {
// 		_, err := os.Stat(tt.zipName)
// 		if err != nil {
// 			t.Errorf("expected err nil when checking file %s, got %q", tt.zipName, err)
// 		}
// 		err = os.Remove(tt.zipName)
// 		if err != nil {
// 			t.Errorf("expected err nil when removing test file, got %q", err)
// 		}
// 	}
// }

// func TestRemoveDuplicates(t *testing.T) {
// 	candidates := []*Candidatura{
// 		&Candidatura{
// 			Legislatura:       2016,
// 			Cargo:             "PREFEITO",
// 			UF:                "AL",
// 			Municipio:         "Maceio",
// 			NumeroUrna:        505,
// 			NomeUrna:          "Lelinho",
// 			Aptidao:           "APTO",
// 			Deferimento:       "DEFERIDO",
// 			TipoAgremiacao:    "AGREMIAÇÃO",
// 			NumeroPartido:     5,
// 			LegendaPartido:    "AM",
// 			NomePartido:       "Arctic Monkeys",
// 			NomeColigacao:     "Arabella",
// 			PartidosColigacao: "Beatles / AM / Zeppelin",
// 			DeclarouBens:      "S",
// 			Situacao:          "SEGUNDO-TURNO",
// 			Turno:             1,
// 			Candidato: Candidato{
// 				CPF:        "07496470430",
// 				Nascimento: "29/11/1997",
// 			},
// 		},
// 		&Candidatura{
// 			Legislatura:       2016,
// 			Cargo:             "PREFEITO",
// 			UF:                "AL",
// 			Municipio:         "Maceio",
// 			NumeroUrna:        505,
// 			NomeUrna:          "Lelinho",
// 			Aptidao:           "APTO",
// 			Deferimento:       "DEFERIDO",
// 			TipoAgremiacao:    "AGREMIAÇÃO",
// 			NumeroPartido:     5,
// 			LegendaPartido:    "AM",
// 			NomePartido:       "Arctic Monkeys",
// 			NomeColigacao:     "Arabella",
// 			PartidosColigacao: "Beatles / AM / Zeppelin",
// 			DeclarouBens:      "S",
// 			Situacao:          "ELEITO",
// 			Turno:             2,
// 			Candidato: Candidato{
// 				CPF:        "07496470430",
// 				Nascimento: "29/11/1997",
// 			},
// 		},
// 		&Candidatura{
// 			Legislatura:       2016,
// 			Cargo:             "PREFEITO",
// 			UF:                "AL",
// 			Municipio:         "Capela",
// 			NumeroUrna:        45666,
// 			NomeUrna:          "Marcelinho",
// 			Aptidao:           "APTO",
// 			Deferimento:       "DEFERIDO",
// 			TipoAgremiacao:    "AGREMIAÇÃO",
// 			NumeroPartido:     45,
// 			LegendaPartido:    "AM",
// 			NomePartido:       "Arctic Monkeys",
// 			NomeColigacao:     "Arabella",
// 			PartidosColigacao: "Beatles / AM / Zeppelin",
// 			DeclarouBens:      "S",
// 			Situacao:          "ELEITO",
// 			Turno:             1,
// 			Candidato: Candidato{
// 				CPF:        "07496470431",
// 				Nascimento: "29/11/1998",
// 			},
// 		},
// 	}
// 	c, err := removeDuplicates(candidates, "file")
// 	if err != nil {
// 		t.Errorf("expected to have err nil when removing duplicated, got %q", err)
// 	}
// 	if len(c) != 2 {
// 		t.Errorf("expected to have only two candidates after removing duplicates, got %d", len(c))
// 	}
// }

// func TestHash(t *testing.T) {
// 	f, err := os.Open("files_2016.zip")
// 	if err != nil {
// 		t.Errorf("expected err to be nil when opening test file")
// 	}
// 	bytes, err := ioutil.ReadAll(f)
// 	if err != nil {
// 		t.Errorf("expected err to be nil when reading file")
// 	}
// 	h, err := hash(bytes)
// 	if err != nil {
// 		t.Errorf("expected err to be nil")
// 	}
// 	if h != "ed90a292c8a264622d726c2d17650d27" {
// 		t.Errorf("want hash ed90a292c8a264622d726c2d17650d27, got %s", h)
// 	}
// }

// func TestDownload(t *testing.T) {
// 	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		fmt.Fprint(w, "string")
// 	}))
// 	defer ts.Close()
// 	var buf bytes.Buffer
// 	b, err := donwloadFile(ts.URL, &buf)
// 	if err != nil {
// 		t.Errorf("want error nil, got %q", err)
// 	}
// 	if buf.String() != "string" {
// 		t.Errorf("want %s, got %s", "string", buf.String())
// 	}
// 	if b == nil {
// 		t.Errorf("expected buf different of nil")
// 	}
// }

// func TestUnzipDownloadedFiles(t *testing.T) {
// 	fileToUnzip := "files_2016.zip"
// 	file, err := os.Open(fileToUnzip)
// 	if err != nil {
// 		t.Errorf("expected to have err nil when opening sample zip files, got %q", err)
// 	}
// 	defer file.Close()
// 	fileBytes, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		t.Errorf("expected to have err nil when reading bytes of sample file, got %q", err)
// 	}
// 	unzipDestination, err := ioutil.TempDir("", "unzipped")
// 	if err != nil {
// 		t.Errorf("expected err nil when creating temporary dir, got %q", err)
// 	}
// 	files, err := unzipDownloadedFiles(fileBytes, unzipDestination)
// 	if err != nil {
// 		t.Errorf("failed to unzip sample files, got %q", err)
// 	}
// 	if len(files) != 6 {
// 		t.Errorf("expected to have 6 csv files decompressed files, got %d", len(files))
// 	}
// 	if err = os.RemoveAll(unzipDestination); err != nil {
// 		t.Errorf("expected to have err nil when removing temporary dir, got %q", err)
// 	}
// }
