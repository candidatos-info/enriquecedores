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
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/labstack/echo"
)

const (
	port             = 9999 // port user to up this local server
	statusCollecting = 1    // integer to represent status collecting
	statusIdle       = 0    // integer to represent status idle
	statusHashing    = 2    // integer to represent status hashing
)

var (
	// http client
	client = &http.Client{
		Timeout: time.Second * 40,
		Transport: &http.Transport{
			Dial: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).Dial,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

// struct used to pass year and source URL to CCE on post request
type postRequest struct {
	Year      int    `json:"year"`
	SourceURL string `json:"source_url"`
}

// response about cce state
type cceStatusResponse struct {
	Status int    `json:"status"`
	Err    string `json:"err"`
}

func main() {
	source := flag.String("coleta", "", "fonte do arquivo zip")
	outDir := flag.String("outdir", "", "diretório de arquivo zip a ser usado pelo CCE")
	year := flag.Int("ano", 0, "ano da eleição")
	state := flag.String("estado", "", "estado a ser processado")
	httpAddress := flag.String("remoteadd", "", "endereço web do servidor") // em produção passar o endereço ngrok, caso contrário passar o endereço local como http://localhost:9999
	cceAddress := flag.String("cceadd", "", "endereço web do cce")
	userName := flag.String("username", "", "user name para basic auth")
	password := flag.String("password", "", "senha para basic auth")
	flag.Parse()
	if *source != "" {
		if *outDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if err := collect(*source, *outDir); err != nil {
			log.Fatalf("falha ao executar coleta, erro %q", err)
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
			log.Fatal("informe o endereço privisionado para este provedor de arquivos")
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
			log.Fatalf("falha ao executar enriquecimento, erro %v", err)
		}
	}
}

func collect(source, outDir string) error {
	tempFile := new(bytes.Buffer)
	bytes, err := donwloadFile(source, tempFile)
	if err != nil {
		return fmt.Errorf("falha ao fazer buscar arquivo com URL %s, erro %q", source, err)
	}
	if _, err := unzipDownloadedFiles(bytes, outDir); err != nil {
		return fmt.Errorf("falha ao descomprimir arquivos baixados, erro %q", err)
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

func process(state, outDir, thisServerAddress, cceAddress, userName, password string, year int) error {
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
	fileURL := fmt.Sprintf("%s/static/%s", thisServerAddress, path.Base(zipName))
	fmt.Println("URL ", fileURL)
	pr := postRequest{
		Year:      year,
		SourceURL: fileURL,
	}
	requestBodyBytes, err := json.Marshal(pr)
	if err != nil {
		return fmt.Errorf("falha ao pegar bytes do corpo da requisição, erro %q", err)
	}
	req, err := http.NewRequest("POST", cceAddress, bytes.NewBuffer(requestBodyBytes))
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
	req, err = http.NewRequest("GET", cceAddress, nil)
	req.Header.Set("Content-type", "application/json")
	req.SetBasicAuth(userName, password)
	cceResponse := cceStatusResponse{
		Status: statusCollecting,
	}
	for {
		if cceResponse.Err != "" {
			log.Printf("enriquecedor de candidaturas foi para status IDLE com erro %s\n", cceResponse.Err)
			break
		}
		if cceResponse.Status >= statusHashing { // passou do status da coleta (status >= 2)
			log.Println("enriquecedor de candidaturas iniciou processamento do arquivo")
			break
		}
		res, err = client.Do(req)
		if err != nil {
			return fmt.Errorf("falha na requisição ao CCE, erro %q", err)
		}
		defer res.Body.Close()
		bodyBytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return fmt.Errorf("falha ao ler bytes do corpo da resposta do CCE, erro %q", err)
		}
		if err = json.Unmarshal(bodyBytes, &cceResponse); err != nil {
			return fmt.Errorf("falha ao fazer unmarshal de resposta do CCE, erro %q", err)
		}
	}
	time.Sleep(time.Second * 20)
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
