package cce

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/candidatos-info/enriquecedores/status"
	"github.com/labstack/echo"
)

// Handler is a struct to hold important data for this package
type Handler struct {
	sheetsServerString string        // URL to retrieve files from TCE
	baseDir            string        // files path
	status             status.Status // enrich status
	err                string        // last error message
}

// used on Post
type postRequest struct {
	Year int `json:"year"`
}

// New returns a new CCE handler
func New(sheetsServerString, baseDir string) *Handler {
	return &Handler{
		sheetsServerString: sheetsServerString,
		baseDir:            baseDir,
		status:             status.Idle,
	}
}

// Get returns current state and last error
func (h *Handler) Get(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"errorMessage": h.err,
		"status":       h.status,
	})
}

func (h *Handler) post(c echo.Context) {
	in := postRequest{}
	if err := c.Bind(&in); err != nil {
		handleError(fmt.Sprintf("o corpo da requisicão enviado é inválido: %q", err), h)
		return
	}
	h.status = status.Collecting
	downloadURL := fmt.Sprintf(h.sheetsServerString, in.Year)
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
	h.status = status.Processing
	hash := md5.New()
	if _, err := io.Copy(hash, buf); err != nil {
		handleError(fmt.Sprintf("falha ao gerar hash de arquivo do TCE baixado, erro: %q", err), h)
		return
	}
	sum := hash.Sum(nil)
	//TODO compare hash with .hash
	fmt.Printf("%x\n", sum)
}

// Post implements a post request for this handler
func (h *Handler) Post(c echo.Context) error {
	if h.status != status.Idle {
		return c.String(http.StatusServiceUnavailable, "sistema está processando dados")
	}
	go h.post(c)
	return c.String(http.StatusOK, "Requisição em processamento")
}

func handleError(message string, h *Handler) {
	log.Println(message)
	h.err = message
}

// download a file and writes on the given writer
func donwloadFile(url string, w io.Writer) (io.Reader, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error downloading file from url %s, got error :%q", url, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return nil, fmt.Errorf("error copying response content:%q", err)
	}
	return resp.Body, nil
}
