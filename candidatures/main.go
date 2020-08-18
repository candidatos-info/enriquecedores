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
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/enriquecedores/filestorage"
	tseutils "github.com/candidatos-info/enriquecedores/tse_utils"
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
	flag.Parse()
	if *source != "" {
		if *localDir == "" {
			log.Fatal("informe diretório de saída")
		}
		s := spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		s.Prefix = fmt.Sprintf("downloading from %s ", *source)
		s.Start()
		if err := collect(*source, *localDir); err != nil {
			log.Fatalf("falha ao executar coleta, erro %q", err)
		}
		s.Stop()
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
		if err := process(*state, *localDir, *candidaturesDir); err != nil {
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

func process(state, outDir, candidaturesDir string) error {
	var client filestorage.FileStorage
	if strings.HasPrefix(candidaturesDir, "gs://") {
		client = filestorage.NewGCSClient()
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
	candidaturesByCity := grouCandidaturesByCity(filteredCandidatures)
	cities := len(candidaturesByCity)
	log.Printf("cities to process [ %d ]\n", cities)
	for city, group := range candidaturesByCity {
		b, err := proto.Marshal(group)
		if err != nil {
			return fmt.Errorf("falha ao serializar grupo de cidades, erro %q", err)
		}
		fileName := fmt.Sprintf("%s_%s", state, city)
		bucket := strings.ReplaceAll(candidaturesDir, "gs://", "")
		err = try.Do(func(attempt int) (bool, error) {
			return attempt < maxAttempts, client.Upload(b, bucket, fileName)
		})
		if err != nil {
			return fmt.Errorf("falha ao salvar arquivo de candidatura %s no bucket %s, erro %q", fileName, bucket, err)
		}
		log.Printf("sent city [ %s ]\n", city)
	}
	return nil
}

// it group candidatures by cities
func grouCandidaturesByCity(cands map[string]*descritor.Candidatura) map[string]*descritor.CandidaturasDeCidade {
	groups := make(map[string]*descritor.CandidaturasDeCidade)
	for _, c := range cands {
		ca := &descritor.Candidatura{
			Legislatura:           c.Legislatura,
			Cargo:                 c.Cargo,
			UF:                    c.UF,
			Municipio:             c.Municipio,
			NumeroUrna:            c.NumeroUrna,
			NomeUrna:              c.NomeUrna,
			Aptdao:                c.Aptdao,
			Deferimento:           c.Deferimento,
			TipoAgremiacao:        c.TipoAgremiacao,
			NumeroPartido:         c.NumeroPartido,
			LegendaPartido:        c.LegendaPartido,
			NomePartido:           c.NomePartido,
			NomeColigacao:         c.NomeColigacao,
			PartidosColigacao:     c.PartidosColigacao,
			DeclarouBens:          c.DeclarouBens,
			SituacaoPrimeiroTurno: c.SituacaoPrimeiroTurno,
			SituacaoSegundoTurno:  c.SituacaoSegundoTurno,
			SequencialCandidato:   c.SequencialCandidato,
			Candidato: &descritor.Candidato{
				UF:              c.Candidato.UF,
				Municipio:       c.Candidato.Municipio,
				TituloEleitoral: c.Candidato.TituloEleitoral,
				Nascimento:      c.Candidato.Nascimento,
				Genero:          c.Candidato.Genero,
				GrauInstrucao:   c.Candidato.GrauInstrucao,
				EstadoCivil:     c.Candidato.EstadoCivil,
				Raca:            c.Candidato.Raca,
				Ocupacao:        c.Candidato.Ocupacao,
				CPF:             c.Candidato.CPF,
				Nome:            c.Candidato.Nome,
				Email:           c.Candidato.Email,
			},
		}
		city := groups[c.Municipio]
		if city == nil {
			group := make(map[string]*descritor.Candidatura)
			group[c.SequencialCandidato] = ca
			groups[c.Municipio] = &descritor.CandidaturasDeCidade{
				Group: group,
			}
		} else {
			groups[c.Municipio].Group[c.SequencialCandidato] = ca
		}
	}
	return groups
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
