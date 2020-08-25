package main

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
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
	"strconv"
	"strings"
	"time"

	"github.com/candidatos-info/enriquecedores/filestorage"
	tseutils "github.com/candidatos-info/enriquecedores/tse_utils"
	"github.com/cheggaaa/pb"
	"github.com/gocarina/gocsv"
	"github.com/golang/protobuf/proto"
	"github.com/matryer/try"
	"golang.org/x/text/encoding/charmap"
)

const (
	port             = 9999 // port user to up this local server
	statusCollecting = 1    // integer to represent status collecting
	statusIdle       = 0    // integer to represent status idle
	statusHashing    = 2    // integer to represent status hashing
	maxAttempts      = 5
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
	source := flag.String("sourceFile", "", "fonte do arquivo zip")
	localDir := flag.String("localDir", "", "diretório para colocar os arquivos .csv de candidaturas")
	state := flag.String("state", "", "estado a ser processado")
	candidaturesDir := flag.String("outDir", "", "local de armazenamento de candidaturas") // if for GCS pass gs://${BUCKET}, if for local pass the local path
	googleDriveCredentialsFile := flag.String("credentials", "", "chave de credenciais o Goodle Drive")
	goodleDriveOAuthTokenFile := flag.String("OAuthToken", "", "arquivo com token oauth")
	flag.Parse()
	if *source != "" {
		if *localDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if err := collect(*source, *localDir); err != nil {
			log.Fatalf("falha ao executar coleta, erro %q", err)
		}
	} else {
		if *candidaturesDir == "" {
			log.Fatal("informe local de armazenamento de candidaturas")
		}
		if *state == "" {
			log.Fatal("informe estado a ser processado")
		}
		if *localDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if err := process(*state, *localDir, *candidaturesDir, *googleDriveCredentialsFile, *goodleDriveOAuthTokenFile); err != nil {
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
	var length int
	tempByffer := new(bytes.Buffer)
	if strings.HasPrefix(url, "http") {
		res, err = c.Get(url)
		if err != nil {
			return nil, fmt.Errorf("problema ao baixar os arquivos da url %s, erro: %q", url, err)
		}
		contentLength := res.Header.Get("content-length")
		length, err = strconv.Atoi(contentLength)
		if err != nil {
			return nil, fmt.Errorf("falha ao pegar tamanho do arquivo a ser baixado, erro %q", err)
		}
	} else if strings.HasPrefix(url, "file") {
		t.RegisterProtocol("file", http.NewFileTransport(http.Dir("/")))
		res, err = c.Get(url)
		if err != nil {
			return nil, fmt.Errorf("falha ao buscar arquivos do sistema com caminho %s, erro: %q", url, err)
		}
		if _, err := io.Copy(tempByffer, res.Body); err != nil {
			return nil, fmt.Errorf("falha copiar bytes da requisição, erro %q", err)
		}
	} else {
		return nil, fmt.Errorf("protocolo %s não suportado", url[0:5])
	}
	defer res.Body.Close()
	if strings.HasPrefix(url, "http") {
		reader := io.LimitReader(res.Body, int64(length))
		bar := pb.Full.Start64(int64(length))
		barReader := bar.NewProxyReader(reader)
		if _, err := io.Copy(tempByffer, barReader); err != nil {
			return nil, fmt.Errorf("falha ao copiar bytes do bar reader, erro %q", err)
		}
		bar.Finish()
	}
	bodyAsBytes, err := ioutil.ReadAll(tempByffer)
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

func process(state, outDir, candidaturesDir, googleDriveCredentialsFile, goodleDriveOAuthTokenFile string) error {
	var client filestorage.FileStorage
	if googleDriveCredentialsFile != "" && goodleDriveOAuthTokenFile != "" {
		var err error
		client, err = filestorage.NewGoogleDriveStorage(googleDriveCredentialsFile, goodleDriveOAuthTokenFile)
		if err != nil {
			return fmt.Errorf("falha ao criar cliente do Google Drive, erro %q", err)
		}
	} else {
		client = filestorage.NewLocalStorage()
	}
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
	file, err := os.Open(pathToHandle)
	if err != nil {
		return fmt.Errorf("falha ao abrir arquivo .csv descomprimido %s, erro %q", file.Name(), err)
	}
	defer file.Close()
	gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
		// Enforcing reading the TSE zip file as ISO 8859-1 (latin 1)
		r := csv.NewReader(charmap.ISO8859_1.NewDecoder().Reader(in))
		r.LazyQuotes = true
		r.Comma = ';'
		return r
	})
	var c []*tseutils.RegistroTSE
	if err := gocsv.UnmarshalFile(file, &c); err != nil {
		return fmt.Errorf("falha ao inflar slice de candidaturas usando arquivo csv %s, erro %q", file.Name(), err)
	}
	filteredCandidatures, err := tseutils.RemoveDuplicates(c, path.Base(pathToHandle))
	if err != nil {
		return fmt.Errorf("falha ao remover candidaturas duplicadas, erro %q", err)
	}
	for _, candidature := range filteredCandidatures {
		b, err := proto.Marshal(candidature)
		if err != nil {
			return fmt.Errorf("falha ao serializar grupo de cidades, erro %q", err)
		}
		fileName := fmt.Sprintf("%s.pb", candidature.SequencialCandidato)
		err = try.Do(func(attempt int) (bool, error) {
			return attempt < maxAttempts, client.Upload(b, candidaturesDir, fileName)
		})
		if err != nil {
			return fmt.Errorf("falha ao salvar arquivo de candidatura [%s] no bucket [%s], erro %s", fileName, candidaturesDir, handleDriveError(err.Error()))
		}
		log.Printf("sent candidate [ %s ]\n", fileName)
		time.Sleep(time.Second * 1) // esse delay é colocado para evitar atingir o limite de requests por segundo. Preste atenção ao tamanho do arquivo que irá enviar.
	}
	return nil
}

func handleDriveError(errorMessage string) string {
	stringsWithErrorCode := strings.Split(strings.Split(errorMessage, "googleapi: ")[1], ":")
	if strings.Contains(stringsWithErrorCode[0], "403") { // o código de erro retornado em caso de estourar o rate limite é 403
		return fmt.Sprintf("recebemos 403 do Google Drive, provavelmente o limite do uso foi atingido. mensagem: %s", stringsWithErrorCode[1])
	}
	return fmt.Sprintf("recebemos erro do Google Drive com messagem %s", stringsWithErrorCode[1])
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
