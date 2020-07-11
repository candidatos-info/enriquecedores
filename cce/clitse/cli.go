package main

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo"
)

const (
	port      = 8080
	timeLimit = time.Second * 2
)

// struct used to pass year and source URL to CCE on post request
type postRequest struct {
	Year      int    `json:"year"`
	SourceURL string `json:"source_url"`
}

func main() {
	source := flag.String("coleta", "", "fonte do arquivo zip")
	outDir := flag.String("outdir", "", "diretório de arquivo zip a ser usado pelo CCE")
	year := flag.Int("ano", 0, "ano da eleição")
	state := flag.String("estado", "", "estado a ser processado")
	httpAddress := flag.String("remoteadd", "", "endereço web do servidor") // em produção passar o endereço ngrok, caso contrário passar http://localhost:8080
	cceAddress := flag.String("cceadd", "", "endereço web do cce")
	userName := flag.String("username", "", "user name para basic auth")
	password := flag.String("password", "", "senha para basic auth")
	flag.Parse()
	if *source != "" {
		if *outDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if err := collect(*source, *outDir); err != nil {
			log.Fatal("falha ao executar coleta, erro %q", err)
		}
	} else {
		if *state == "" {
			log.Fatal("informe estado a ser processado")
		}
		if *year == 0 {
			log.Fatal("informe ano a ser processado")
		}
		if *outDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if *httpAddress == "" {
			log.Fatal("informe o endereço fornecido pelo NGROK")
		}
		if *cceAddress == "" {
			log.Fatal("informe o endereço do CCE")
		}
		if *userName == "" {
			log.Fatal("informe o login de basic auth")
		}
		if *password == "" {
			log.Fatal("informe a senha de basic auth")
		}
		if err := process(*state, *outDir, *httpAddress, *cceAddress, *userName, *password, *year); err != nil {
			log.Fatal("falha ao executar enriquecimento, erro %v", err)
		}
	}
}

func collect(source, outDir string) error {
	tempFile, err := ioutil.TempFile("", "temporaryFile")
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo temporário para arquivo .zip", err)
	}
	bytes, err := donwloadFile(source, tempFile)
	if err != nil {
		return fmt.Errorf("falha ao fazer buscar arquivo com URL %s, erro %q", source, err)
	}
	if _, err := unzipDownloadedFiles(bytes, outDir); err != nil {
		return fmt.Errorf("falha ao descomprimir arquivos baixados, erro %q", err)
	}
	if err = os.RemoveAll(tempFile.Name()); err != nil {
		return fmt.Errorf("falha ao deletar arquivo temporário criado, erro %q", err)
	}
	return nil
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

// It unzips downloaded .zip on a temporary directory
// and returns the path of unziped files with suffix .csv
func unzipDownloadedFiles(buf []byte, unzipDestination string) ([]string, error) {
	zipReader, err := zip.NewReader(bytes.NewReader(buf), int64(len(buf)))
	if err != nil {
		return nil, err
	}
	var paths []string
	for _, f := range zipReader.File {
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("falha ao abrir arquivo %s, erro %q", f.Name, err)
		}
		path := filepath.Join(unzipDestination, f.Name)
		if strings.HasSuffix(path, ".csv") {
			paths = append(paths, path)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(path, f.Mode()); err != nil {
				return nil, fmt.Errorf("falha ao criar diretório com nome %s, erro %q", path, err)
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(path), f.Mode()); err != nil {
				return nil, fmt.Errorf("falha ao criar diretório com nome %s, erro %q", path, err)
			}
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return nil, fmt.Errorf("falha ao abrir arquivo %s, erro %q", path, err)
			}
			if _, err = io.Copy(f, rc); err != nil {
				return nil, fmt.Errorf("falha ao copiar conteúdo para arquivo temporário %s", path)
			}
			if err := f.Close(); err != nil {
				return nil, fmt.Errorf("falha ao fechar arquivo criado em diretorio temporario, erro %q", err)
			}
		}
		if err := rc.Close(); err != nil {
			return nil, fmt.Errorf("falha ao fechar leitor de arquivo dentro do zip, erro %q", err)
		}
	}
	return paths, nil
}

func process(state, outDir, ngrokAddress, cceAddress, userName, password string, year int) error {
	pathToHandle := ""
	err := filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, state) {
			pathToHandle = path
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("falha ao varrer arquivos no diretório %s, erro %q", outDir, err)
	}
	if pathToHandle == "" {
		return fmt.Errorf("falha ao encontrar arquivo para estado %s", state)
	}
	fileBytes, err := ioutil.ReadFile(pathToHandle)
	if err != nil {
		return fmt.Errorf("falha ao ler bytes de arquivo de estado %s, erro %q", pathToHandle, err)
	}
	zipName := fmt.Sprintf("%s/ARQUIVO_%s_%d.zip", outDir, state, year)
	fileName := path.Base(pathToHandle)
	if err = zipFile(fileBytes, zipName, fileName); err != nil {
		return fmt.Errorf("falha ao comprimir arquivo %s, erro %q", pathToHandle, err)
	}
	go func() {
		e := echo.New()
		e.Static("/static", outDir)
		e.Start(fmt.Sprintf(":%d", port))
	}()
	fileURL := fmt.Sprintf("%s/static/%s", ngrokAddress, path.Base(zipName))
	pr := postRequest{
		Year:      year,
		SourceURL: fileURL,
	}
	requestBodyBytes, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("falha ao pegar bytes do corpo da requisição, erro %q", err)
	}
	client := &http.Client{}
	req, err := http.NewRequest("POST", "http://localhost:8877/cce", bytes.NewBuffer(requestBodyBytes))
	req.Header.Set("Content-type", "application/json")
	req.SetBasicAuth(userName, password)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("falha na requisição ao CCE, erro %q", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return fmt.Errorf("código de resposta esperado era 200, tivemos %d", res.StatusCode)
	}
	time.Sleep(timeLimit) // sleep é importante para garantir um tempo de "sobra"
	if err = os.Remove(zipName); err != nil {
		return fmt.Errorf("falha ao deletar arquivo zip criado, erro %q", err)
	}
	return nil
}

// it gets an array of bytes to write into a file called fileName that
// will be compressed into a zip called zipName
func zipFile(bytesToWrite []byte, zipName, fileName string) error {
	outFile, err := os.Create(zipName)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo zip %s, erro %q", zipName, err)
	}
	defer outFile.Close()
	w := zip.NewWriter(outFile)
	defer w.Close()
	f, err := w.Create(fileName)
	if err != nil {
		return fmt.Errorf("falha ao criar o zip, err %q", err)
	}
	if _, err = f.Write(bytesToWrite); err != nil {
		return fmt.Errorf("falha ao escrever o zip, err %q", err)
	}
	return nil
}
