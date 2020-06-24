package cce

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/candidatos-info/enriquecedores/status"
	"github.com/labstack/echo"
)

// Handler is a struct to hold important data for this package
type Handler struct {
	SourceURL        string        `json:"source_url"`        // URL to retrieve files. It can be a path for a file or an URL
	Status           status.Status `json:"status"`            // enrich status
	Err              string        `json:"err"`               // last error message
	SourceFileHash   string        `json:"source_file_hash"`  // hash of last downloaded .zip file
	SourceLocalPath  string        `json:"source_local_path"` // the path where downloaded files should stay
	CandidaturesPath string        `json:"candidatures_path"` // the place where candidatures files will stay
}

// used on Post
type postRequest struct {
	Year int `json:"year"`
}

// New returns a new CCE handler
func New(sheetsServerString, sourceLocalPath string) *Handler {
	return &Handler{
		SourceURL:        sheetsServerString,
		CandidaturesPath: sourceLocalPath,
		Status:           status.Idle,
	}
}

// Get returns current state and last error
func (h *Handler) Get(c echo.Context) error {
	return c.JSON(http.StatusOK, h)
}

func (h *Handler) post(in *postRequest) {
	h.Status = status.Collecting
	h.SourceURL = fmt.Sprintf(h.SourceURL, in.Year)
	zipFileName := fmt.Sprintf("cce_sheets_%d.zip", in.Year)
	f, err := os.Create(zipFileName)
	if err != nil {
		handleError(fmt.Sprintf("ocorreu uma falha durante a criação dos arquivos zip com nome %s, erro: %q", zipFileName, err), h)
		return
	}
	buf, err := donwloadFile(h.SourceURL, f)
	if err != nil {
		handleError(fmt.Sprintf("ocorreu uma falha ao fazer o download dos arquivos csv da legislatura %d pelo link %s, errro: %q", in.Year, h.SourceURL, err), h)
		return
	}
	h.Status = status.Processing
	ha, err := hash(buf)
	if err != nil {
		handleError(fmt.Sprintf("falha ao gerar hash de arquivo do TCE baixado, erro: %q", err), h)
		return
	}
	if strings.HasPrefix(h.CandidaturesPath, "gc://") {
		// TODO add GCS implementation
	} else {
		if err := executeForLocal(ha, in.Year, buf); err != nil {
			handleError(fmt.Sprintf("falha executar processamento local, erro: %q", err), h)
			return
		}
	}
}

func executeForLocal(hash string, year int, buf []byte) error {
	hashFile, err := resolveHashFile(year)
	if err != nil {
		return err
	}
	hashFileBytes, err := ioutil.ReadAll(hashFile)
	if err != nil {
		return fmt.Errorf("failed to read bytes of file %s, got error %q", hashFile.Name(), err)
	}
	if hash == string(hashFileBytes) {
		log.Printf("arquivo baixado é o mesmo, possui o mesmo hash %s\n", hash)
		return nil
	}
	// TODO unzip file and iterate through files
	return nil
}

func resolveHashFile(year int) (*os.File, error) {
	hashFileName := fmt.Sprintf("cce_hash_%d", year)
	_, err := os.Stat(hashFileName)
	if err == nil {
		f, err := os.Open(hashFileName)
		if err != nil {
			return nil, fmt.Errorf("failed to open file %s, got %q", hashFileName, err)
		}
		return f, nil
	}
	hashFile, err := os.Create(hashFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create %s file for cce, got %q", hashFileName, err)
	}
	return hashFile, nil
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
	h.Status = status.Idle
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
