package cce

import (
	"archive/zip"
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/candidatos-info/enriquecedores/status"
	"github.com/labstack/echo"
)

// Handler is a struct to hold important data for this package
type Handler struct {
	SheetsServerString string        `json:"sheets_server_string"` // URL to retrieve files from TCE
	BaseDir            string        `json:"base_dir"`             // path where .hash and candidates file will be placed. If you want to use GCS as storage put gc://${BUCKET}, but if you want to use as local just use .
	Status             status.Status `json:"status"`               // enrich status
	Err                string        `json:"err"`                  // last error message
	FileHash           string        `json:"file_hash"`            // hash of last downloaded .zip file
}

// used on Post
type postRequest struct {
	Year int `json:"year"`
}

// New returns a new CCE handler
func New(sheetsServerString, baseDir string) *Handler {
	return &Handler{
		SheetsServerString: sheetsServerString,
		BaseDir:            baseDir,
		Status:             status.Idle,
	}
}

// Get returns current state and last error
func (h *Handler) Get(c echo.Context) error {
	return c.JSON(http.StatusOK, h)
}

func (h *Handler) post(in *postRequest) {
	h.Status = status.Collecting
	log.Println("starting to collect")
	downloadURL := fmt.Sprintf(h.SheetsServerString, in.Year)
	zipFileName := fmt.Sprintf("cce_sheets_%d.zip", in.Year)
	f, err := os.Create(zipFileName)
	if err != nil {
		handleError(fmt.Sprintf("ocorreu uma falha durante a criação dos arquivos zip com nome %s, erro: %q", zipFileName, err), h)
		return
	}
	buf, err := donwloadFile(downloadURL, f)
	if err != nil {
		handleError(fmt.Sprintf("ocorreu uma falha ao fazer o download dos arquivos csv da legislatura %d pelo link %s, errro: %q", in.Year, downloadURL, err), h)
		return
	}
	h.Status = status.Processing
	log.Println("starting processment")
	hash, err := hash(buf)
	if err != nil {
		handleError(fmt.Sprintf("falha ao gerar hash de arquivo do TCE baixado, erro: %q", err), h)
		return
	}
	fmt.Println(hash)
	if strings.Contains(h.BaseDir, "gc://") {
		executeForGCP()
	} else {
		executeForLocal(hash, buf)
	}
}

func executeForLocal(hash string, buf []byte) {
	file, err := os.Open(".hash")
	// checking if .hash file already exists
	if err != nil || file == nil {
		// TODO execute action
		file, err := os.Create(".hash")
		if err != nil {
			return fmt.Errorf("failed to create .hash file, got %q", err)
		}
		_, err := file.Write([]byte(hash))
		if err != nil {
			return fmt.Errorf("failed to write hash on .hash file, got %q", err)
		}
	}
	hashFileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read bytes from .hash file, got %q", err)
	}
	if hash == string(hashFileBytes) {
		return nil
	}
	// TODO execute action
}

func unzipDownloadedFiles(bytes []byte) ([]os.FileInfo, error) {
	unzipDesitination := "unziped"
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

func executeForGCP() {

}

// Post implements a post request for this handler
func (h *Handler) Post(c echo.Context) error {
	if h.Status != status.Idle {
		return c.String(http.StatusServiceUnavailable, "sistema está processando dados")
	}
	in := &postRequest{}
	if err := c.Bind(&in); err != nil {
		return c.String(http.StatusInternalServerError, fmt.Sprintf("o corpo da requisicão enviado é inválido: %q", err))
	}
	go h.post(in)
	return c.String(http.StatusOK, "Requisição em processamento")
}

func hash(b []byte) (string, error) {
	hash := md5.New()
	if _, err := io.Copy(hash, bytes.NewReader(b)); err != nil {
		return "", err
	}
	sum := hash.Sum(nil)
	return fmt.Sprintf("%x", sum), nil
}

func handleError(message string, h *Handler) {
	log.Println(message)
	h.Err = message
}

// download a file and writes on the given writer
func donwloadFile(url string, w io.Writer) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading file from url %s, got error :%q", url, err)
	}
	defer resp.Body.Close()
	bodyAsBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body, got %q", err)
	}
	_, err = w.Write(bodyAsBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to write bytes on file, got %q", err)
	}
	return bodyAsBytes, nil
}
