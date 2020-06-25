package cce

import (
	"archive/zip"
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
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

func (h *Handler) post() {
	h.Status = status.Collecting
	h.SourceLocalPath = fmt.Sprintf("cce_%s", path.Base(h.SourceURL))
	f, err := os.Create(h.SourceLocalPath)
	if err != nil {
		handleError(fmt.Sprintf("ocorreu uma falha durante a criação dos arquivos zip com nome %s, erro: %q", h.SourceLocalPath, err), h)
		return
	}
	buf, err := donwloadFile(h.SourceURL, f)
	if err != nil {
		handleError(fmt.Sprintf("ocorreu uma falha ao fazer o download dos arquivos csv pelo link %s, errro: %q", h.SourceURL, err), h)
		return
	}
	h.Status = status.Hashing
	ha, err := hash(buf)
	if err != nil {
		handleError(fmt.Sprintf("falha ao gerar hash de arquivo do TSE baixado, erro: %q", err), h)
		return
	}
	if strings.HasPrefix(h.CandidaturesPath, "gc://") {
		// TODO add GCS implementation
	} else {
		if err := executeForLocal(ha, buf, h); err != nil {
			handleError(fmt.Sprintf("falha executar processamento local, erro: %q", err), h)
			return
		}
	}
}

func executeForLocal(hash string, buf []byte, h *Handler) error {
	hashFile, err := resolveHashFile(h.SourceURL)
	if err != nil {
		return err
	}
	hashFileBytes, err := ioutil.ReadAll(hashFile)
	if err != nil {
		return fmt.Errorf("falha ao ler os bytes do arquivo %s, erro: %q", hashFile.Name(), err)
	}
	if hash == string(hashFileBytes) {
		log.Printf("arquivo baixado é o mesmo (possui o mesmo hash %s)\n", hash)
		return nil
	}
	h.Status = status.Processing
	downloadedFiles, err := unzipDownloadedFiles(buf)
	if err != nil {
		return fmt.Errorf("falha ao descomprimir arquivos baixados, erro %q", err)
	}
	for _, file := range downloadedFiles {
		csvReader := csv.NewReader(bufio.NewReader(file))
		csvReader.Comma = ';'
		csvReader.LazyQuotes = true
		// TODO read lines and mount struct Descritor
	}
	return nil
}

// It unzips downloaded .zip and returns only the .csv files
func unzipDownloadedFiles(buf []byte) ([]*os.File, error) {
	unzipDestination := "unzipped"
	os.MkdirAll(unzipDestination, 0755)
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Fatal(err)
			}
		}()
		path := filepath.Join(unzipDestination, f.Name)
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
					log.Fatal(err)
				}
			}()
			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}
	zipReader, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return nil, err
	}
	for _, f := range zipReader.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return nil, err
		}
	}
	var files []*os.File
	err = filepath.Walk(unzipDestination, func(path string, info os.FileInfo, err error) error {
		pathToFile := fmt.Sprintf("%s", path)
		if path != unzipDestination && strings.HasSuffix(path, ".csv") {
			file, err := os.Open(pathToFile)
			if err != nil {
				return fmt.Errorf("falha ao abrir arquivo %s, erro: %q", pathToFile, err)
			}
			files = append(files, file)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

func resolveHashFile(sourceURL string) (*os.File, error) {
	hashFileName := fmt.Sprintf("cce_hash_%s", path.Base(sourceURL))
	_, err := os.Stat(hashFileName)
	if err == nil {
		f, err := os.Open(hashFileName)
		if err != nil {
			return nil, fmt.Errorf("falha ao abrir o arquivo %s, erro: %q", hashFileName, err)
		}
		return f, nil
	}
	hashFile, err := os.Create(hashFileName)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar arquivo %s para cce, erro: %q", hashFileName, err)
	}
	return hashFile, nil
}

// Post implements a post request for this handler
func (h *Handler) Post(c echo.Context) error {
	if h.Status != status.Idle {
		return c.String(http.StatusServiceUnavailable, "sistema está processando dados")
	}
	go h.post()
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
	var res *http.Response
	var err error
	t := &http.Transport{}
	c := &http.Client{Transport: t}
	if strings.HasPrefix(url, "http") {
		res, err = c.Get(url)
		if err != nil {
			return nil, fmt.Errorf("problema ao baixar os arquivos da url %s, erro: %q", url, err)
		}
	} else if strings.HasPrefix(url, "file") {
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		res, err = c.Get(url)
		if err != nil {
			return nil, fmt.Errorf("falha ao buscar arquivos do sistema com caminho %s, erro: %q", url, err)
		}
	} else {
		return nil, fmt.Errorf("protocolo %s não suportado", url[0:5])
	}
	defer res.Body.Close()
	bodyAsBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler os bytes da resposta da requisição, erro: %q", err)
	}
	_, err = w.Write(bodyAsBytes)
	if err != nil {
		return nil, fmt.Errorf("falha ao escrever bytes no arquivo, erro: %q", err)
	}
	return bodyAsBytes, nil
}
