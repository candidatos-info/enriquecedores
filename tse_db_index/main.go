package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/csv"
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

	"cloud.google.com/go/datastore"
	"github.com/candidatos-info/descritor"
	tseutils "github.com/candidatos-info/enriquecedores/tse_utils"
	"github.com/gocarina/gocsv"
	"golang.org/x/text/encoding/charmap"
)

const (
	candidaturesCollection = "candidatures"
)

// db schema
type votingCity struct {
	City       string
	State      string
	Candidates []*descritor.Candidatura
}

func main() {
	source := flag.String("collect", "", "fonte do arquivo zip do TSE com as candidaturas") // pode ser um path usando protocolo file:// ou http://
	localDir := flag.String("localDir", "", "diretório de saída onde os arquivos descomprimidos serão colocados")
	state := flag.String("state", "", "estado para ser enriquecido")
	projectID := flag.String("projectID", "", "id do projeto no Google Cloud")
	flag.Parse()
	if *source != "" {
		if *localDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if err := collect(*source, *localDir); err != nil {
			log.Fatalf("falha ao executar coleta, erro %q", err)
		}
	} else {
		if *projectID == "" {
			log.Fatal("informe o id de projeto")
		}
		client, err := datastore.NewClient(context.Background(), *projectID)
		if err != nil {
			log.Fatalf("falha ao criar cliente de datastore: %v", err)
		}
		if *localDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if *state == "" {
			log.Fatal("informar o estado a ser enriquecido")
		}
		if err := process(*localDir, *state, client); err != nil {
			log.Fatalf("falha ao processar dados para enriquecimento do banco, erro %q", err)
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

func process(localDir, state string, client *datastore.Client) error {
	pathToOpen := ""
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, state) {
			pathToOpen = path
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("falha ao passear entre arquivos do diretório %s, erro %q", localDir, err)
	}
	if pathToOpen == "" {
		return fmt.Errorf("falha ao encontrar path para estado %s", state)
	}
	file, err := os.Open(pathToOpen)
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
	filteredCandidatures, err := tseutils.RemoveDuplicates(c, path.Base(pathToOpen))
	if err != nil {
		return fmt.Errorf("falha ao remover candidaturas duplicadas, erro %q", err)
	}
	if err := saveCandidatures(state, filteredCandidatures, client); err != nil {
		return fmt.Errorf("falha ao salvar candidaturas no banco, erro %q", err)
	}
	return nil
}

func saveCandidatures(state string, candidatures map[string]*descritor.Candidatura, client *datastore.Client) error {
	groupedCandidatures := groupCandidaturesByCity(candidatures)
	for city, candidaturesList := range groupedCandidatures {
		votingCity := votingCity{
			City:       city,
			State:      state,
			Candidates: candidaturesList,
		}
		votinLocationID := datastore.NameKey(candidaturesCollection, fmt.Sprintf("%s_%s", state, city), nil)
		if _, err := client.Put(context.Background(), votinLocationID, &votingCity); err != nil {
			return fmt.Errorf("falha ao salvar local de votação para estado [%s] e cidade [%s] no banco, erro %q", state, city, err)
		}
	}
	return nil
}

func groupCandidaturesByCity(candidatures map[string]*descritor.Candidatura) map[string][]*descritor.Candidatura {
	cities := make(map[string][]*descritor.Candidatura)
	for _, candidature := range candidatures {
		foundCity := cities[candidature.Municipio]
		if foundCity == nil {
			cities[candidature.Municipio] = []*descritor.Candidatura{candidature}
		} else {
			cities[candidature.Municipio] = append(cities[candidature.Municipio], candidature)
		}
	}
	return cities
}
