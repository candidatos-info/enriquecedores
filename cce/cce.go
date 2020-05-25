package cce

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo"
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

// Post should be called to dispatch the process
func (h *Handler) Post(c echo.Context) error {
	status = collecting
	in := dispatchRequest{}
	err := c.Bind(&in)
	payload := make(map[string]string)
	if err != nil {
		log.Println(fmt.Sprintf("failed to bind request input: %q", err))
		payload["message"] = "Invalid request body"
		return c.JSON(http.StatusUnprocessableEntity, payload)
	}
	downloadURL := fmt.Sprintf(hostURL, in.Year)
	zipFile := fmt.Sprintf("sheets_%d.zip", in.Year)
	f, err := os.Create(zipFile)
	if err != nil {
		log.Println(fmt.Sprintf("failed to create sheets zip file, got %q", err))
		payload["message"] = "failed to sheet files"
		return c.JSON(http.StatusInternalServerError, payload)
	}
	err = donwloadFile(downloadURL, f)
	if err != nil {
		log.Println(fmt.Sprintf("failed to download sheets, got %q", err))
		payload["message"] = "failed download sheets"
		return c.JSON(http.StatusInternalServerError, payload)
	}
	status = processing
	zipDestination := strings.Split(zipFile, ".zip")[0]
	err = unzip(zipFile, zipDestination)
	if err != nil {
		log.Println(fmt.Sprintf("failed to unzip files, %q", err))
		payload["message"] = "failed to unzip files"
		return c.JSON(http.StatusInternalServerError, payload)
	}
	err = processFiles(zipDestination)
	if err != nil {
		log.Println(fmt.Sprintf("failed on processing files, got %q", err))
		payload["message"] = "failed process files"
		return c.JSON(http.StatusInternalServerError, payload)
	}
	payload["message"] = "ok"
	return c.JSON(http.StatusOK, payload)
}

func processFiles(filesToProcess string) error {
	return nil
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
