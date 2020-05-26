package cce

import (
	"archive/zip"
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/labstack/echo"
)

const (
	idle       = "idle"
	collecting = "collecting"
	processing = "processing"

	ballotNumber  = 16 //ballor number on the csv file is at column 16
	candidateName = 17 //candidate name on the csv file is at column 17
	ballotName    = 18 //ballor name on the csv file is at column 18
	cpf           = 20 //cpf on csv is at column 20
	email         = 21 //email on csv is at column 21
	state         = 10 //state on csv is at column 10
	year          = 2  //year on csv in at column 2
)

type payload struct {
	BallotName    string `json:"ballotName"`
	BallotNumber  string `json:"ballotNumber"`
	CPF           string `json:"cpf"`
	CandidateName string `json:"name"`
	Email         string `json:"email"`
	State         string `json:"state"`
	Year          string `json:"year"`
}

type message struct {
	Year         int64  `json:"year"`
	BallotNumber int64  `json:"ballotNumber"`
	State        string `json:"state"`
	Payload      []byte `json:"payload"`
	Origin       string `json:"origin"`
}

var (
	// hostURLFormatString can be changed by TestPost function test for tests purposes
	hostURLFormatString = "http://agencia.tse.jus.br/estatistica/sead/odsele/consulta_cand/consulta_cand_%d.zip"
)

// Handler is struct for the methods
type Handler struct {
	status string
}

// NewHandler does
func NewHandler() *Handler {
	return &Handler{
		status: "idle",
	}
}

type postRequest struct {
	Year int64 `json:"year"`
}

// Post should be called to dispatch the process
func (h *Handler) Post(c echo.Context) error {
	h.status = collecting
	in := postRequest{}
	if err := c.Bind(&in); err != nil {
		log.Println(fmt.Sprintf("failed to bind request input: %q", err))
		bodyBytes, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			log.Println(fmt.Sprintf("failed to get request body as bytes, got %q", err))
			message := fmt.Sprintf("houve falha ao pegar os bytes do corpo da requisição com erro %q", err)
			return c.JSON(http.StatusInternalServerError, message)
		}
		message := fmt.Sprintf("o corpo da requisicão enviado é inválido: %q", string(bodyBytes))
		return c.JSON(http.StatusBadRequest, message)
	}
	downloadURL := fmt.Sprintf(hostURLFormatString, in.Year)
	zipFileName := fmt.Sprintf("cce_sheets_%d.zip", in.Year)
	zipFile := fmt.Sprintf(zipFileName, in.Year)
	f, err := os.Create(zipFile)
	if err != nil {
		log.Println(fmt.Sprintf("failed to create sheets zip file with name %s, got %q", zipFileName, err))
		message := fmt.Sprintf("ocorreu uma falha durante a criação dos arquivos zip com nome %s, erro: %q", zipFileName, err)
		return c.JSON(http.StatusInternalServerError, message)
	}
	err = donwloadFile(downloadURL, f)
	if err != nil {
		log.Println(fmt.Sprintf("failed to download sheets from url %s, got %q", downloadURL, err))
		message := fmt.Sprintf("ocorreu uma falha ao fazer o download dos arquivos csv da legislatura %d pelo link %s, errro: %q", in.Year, downloadURL, err)
		return c.JSON(http.StatusInternalServerError, message)
	}
	h.status = processing
	zipDestination := strings.Split(zipFile, ".zip")[0]
	err = unzip(zipFile, zipDestination)
	if err != nil {
		log.Println(fmt.Sprintf("failed to unzip file %s,error: %q", zipFile, err))
		message := fmt.Sprintf("ocorreu uma falha ao descomprimir o arquivo %s, dando erro: %q", zipFile, err)
		return c.JSON(http.StatusInternalServerError, message)
	}
	err = processFiles(zipDestination)
	if err != nil {
		log.Println(fmt.Sprintf("failed on processing files, got %q", err))
		message := fmt.Sprintf("ocorreu uma falha processando os arquivos no diretório %s, erro: %q", zipDestination, err)
		return c.JSON(http.StatusInternalServerError, message)
	}
	message := "processo ocorreu sem falhas!"
	return c.JSON(http.StatusOK, message)
}

func processFiles(filesToProcess string) error {
	files, err := ioutil.ReadDir(filesToProcess)
	if err != nil {
		return fmt.Errorf("failed to read files dir %s, got error: %q", filesToProcess, err)
	}
	for _, f := range files {
		fileName := f.Name()
		extension := strings.Split(fileName, ".")[1]
		if extension == "csv" {
			pathToOpen := fmt.Sprintf("./%s/%s", filesToProcess, fileName)
			f, err := os.Open(pathToOpen)
			defer f.Close()
			if err != nil {
				return fmt.Errorf("failed to open sheet file %s, got %q", pathToOpen, err)
			}
			csvReader := csv.NewReader(bufio.NewReader(f))
			csvReader.Comma = ';'
			csvReader.LazyQuotes = true
			currentLine := 0
			for {
				line, err := csvReader.Read()
				if err == io.EOF {
					break
				} else if err != nil {
					return fmt.Errorf("failed to read csv file %s, got %q", pathToOpen, err)
				}
				if currentLine > 0 {
					payload := &payload{
						BallotName:    line[ballotName],
						BallotNumber:  line[ballotNumber],
						CPF:           line[cpf],
						CandidateName: line[candidateName],
						Email:         line[email],
					}
					payloadBytes, err := json.Marshal(payload)
					if err != nil {
						return fmt.Errorf("failed to get message bytes, got %q", err)
					}
					year, err := strconv.ParseInt(line[year], 10, 64)
					if err != nil {
						return fmt.Errorf("failed to parse year from string to int, got %q", err)
					}
					ballotNumber, err := strconv.ParseInt(line[ballotNumber], 10, 64)
					if err != nil {
						return fmt.Errorf("failed to parse ballot number from string to int, got %q", err)
					}
					message := message{
						Payload:      payloadBytes,
						Year:         year,
						BallotNumber: ballotNumber,
						State:        line[state],
						Origin:       "cce",
					}
					fmt.Println(message)
				}
				currentLine++
			}
			if err != nil {
				return fmt.Errorf("unable to parse csv file %s, got %q", pathToOpen, err)
			}
		}
	}
	return nil
}

// unzip files
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
		return fmt.Errorf("error downloading file from url %s, got error :%q", url, err)
	}
	defer resp.Body.Close()
	if _, err := io.Copy(w, resp.Body); err != nil {
		return fmt.Errorf("error copying response content:%q", err)
	}
	return nil
}
