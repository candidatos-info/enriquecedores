package cce

import (
	"fmt"
	"io"
	"io/ioutil"
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
		err:                "",
	}
}

// Get returns current state and last error
func (h *Handler) Get(c echo.Context) error {
	payload := make(map[string]interface{})
	payload["errorMessage"] = h.err
	payload["status"] = h.status
	return c.JSON(http.StatusOK, payload)
}

// Post implements a post request for this handler
func (h *Handler) Post(c echo.Context) error {
	if h.status != status.Idle {
		return c.JSON(http.StatusServiceUnavailable, "sistema está processando dados")
	}
	h.status = status.Collecting
	go func() {
		in := postRequest{}
		if err := c.Bind(&in); err != nil {
			log.Println(fmt.Sprintf("failed to bind request input: %q", err))
			bodyBytes, err := ioutil.ReadAll(c.Request().Body)
			if err != nil {
				log.Println(fmt.Sprintf("failed to get request body as bytes, got %q", err))
				h.err = fmt.Sprintf("houve falha ao pegar os bytes do corpo da requisição com erro %q", err)
				return
			}
			h.err = fmt.Sprintf("o corpo da requisicão enviado é inválido: %q", string(bodyBytes))
			return
		}
		downloadURL := fmt.Sprintf(h.sheetsServerString, in.Year)
		zipFileName := fmt.Sprintf("cce_sheets_%d.zip", in.Year)
		f, err := os.Create(zipFileName)
		if err != nil {
			log.Println(fmt.Sprintf("failed to create sheets zip file with name %s, got %q", zipFileName, err))
			h.err = fmt.Sprintf("ocorreu uma falha durante a criação dos arquivos zip com nome %s, erro: %q", zipFileName, err)
			return
		}
		err = donwloadFile(downloadURL, f)
		if err != nil {
			log.Println(fmt.Sprintf("failed to download sheets from url %s, got %q", downloadURL, err))
			h.err = fmt.Sprintf("ocorreu uma falha ao fazer o download dos arquivos csv da legislatura %d pelo link %s, errro: %q", in.Year, downloadURL, err)
			return
		}
		h.status = status.Processing
		// TODO implement processing
	}()
	return c.JSON(http.StatusOK, "Requisição em processamento")
}

// download a file and writes on the given writer
func donwloadFile(url string, w io.Writer) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading file from url %s, got error :%q", url, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("error copying response content:%q", err)
	}
	return nil
}
