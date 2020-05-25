package cce

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const (
	idle       = "idle"
	collecting = "collecting"
	processing = "processing"
)

var (
	// hostURL can be changed by TestPost function test for tests purposes
	hostURL = "http://agencia.tse.jus.br/estatistica/sead/odsele/consulta_cand/consulta_cand_%d.zip"
	status  = idle
)

// Handler is struct for the methods
type Handler struct {
}

// NewHandler does
func NewHandler() *Handler {
	return &Handler{}
}

type dispatchRequest struct {
	Year int64 `json:"year"`
}

func unzip(fileUnzip, unzipDesitination string) error {
	r, err := zip.OpenReader(fileUnzip)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()
	os.MkdirAll(unzipDesitination, 0755)
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()
		path := filepath.Join(unzipDesitination, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}
	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}
	return nil
}

// download a file and writes on the given writer
func donwloadFile(url string, w io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading file:%q", err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("error copying response content:%q", err)
	}
	return nil
}
