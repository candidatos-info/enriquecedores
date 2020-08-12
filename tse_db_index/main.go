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

// TODO refact these items

// Candidato representa os dados de um candidato
type Candidato struct {
	UF              string `csv:"SG_UF_NASCIMENTO"`              // Identificador (2 caracteres) da unidade federativa de nascimento do candidato.
	Municipio       string `csv:"NM_MUNICIPIO_NASCIMENTO"`       // Município de nascimento do candidato.
	Nascimento      string `csv:"DT_NASCIMENTO"`                 // Data de nascimento do candidato.
	TituloEleitoral string `csv:"NR_TITULO_ELEITORAL_CANDIDATO"` // Titulo eleitoral do candidato.
	Genero          string `csv:"DS_GENERO"`                     // Gênero do candidato (MASCULINO ou FEMININO).
	GrauInstrucao   string `csv:"DS_GRAU_INSTRUCAO"`             // Grau de instrução do candidato.
	EstadoCivil     string `csv:"DS_ESTADO_CIVIL"`               // Estado civil do candidato.
	Raca            string `csv:"DS_COR_RACA"`                   // Raça do candidato (como BRANCA ou PARDA).
	Ocupacao        string `csv:"DS_OCUPACAO"`                   // Ocupação do candidato (como COMERCIANTE e ARTISTA por exemplo).
	CPF             string `csv:"NR_CPF_CANDIDATO"`              // CPF do candidato.
	Nome            string `csv:"NM_CANDIDATO"`                  // Nome de pessoa física do candidato.
	Email           string `csv:"NM_EMAIL"`                      // Email do candidato.
}

// Candidatura representa dados de uma candidatura
type Candidatura struct {
	Legislatura         int    `csv:"ANO_ELEICAO"`              // Ano eleitoral em que a candidatura foi homologada.
	Cargo               string `csv:"DS_CARGO"`                 // Cargo sendo pleiteado pela candidatura.
	UF                  string `csv:"SG_UF"`                    // Identificador (2 caracteres) de unidade federativa onde ocorreu a candidatura.
	Municipio           string `csv:"NM_UE"`                    // Município que ocorreu a eleição.
	NumeroUrna          int    `csv:"NR_CANDIDATO"`             // Número do candidato na urna.
	NomeUrna            string `csv:"NM_URNA_CANDIDATO"`        // Nome do candidato na urna.
	Aptidao             string `csv:"DS_SITUACAO_CANDIDATURA"`  // Aptidao da candidatura (podendo ser APTO ou INAPTO).
	Deferimento         string `csv:"DS_DETALHE_SITUACAO_CAND"` // Situação do candidato (pondendo ser DEFERIDO ou INDEFERIDO).
	TipoAgremiacao      string `csv:"TP_AGREMIACAO"`            // Indica o tipo de agremiação do candidato (podendo ser PARTIDO ISOLADO ou AGREMIAÇÃO).
	NumeroPartido       int    `csv:"NR_PARTIDO"`               // Número do partido do candidato.
	LegendaPartido      string `csv:"SG_PARTIDO"`               // Legenda do partido do candidato.
	NomePartido         string `csv:"NM_PARTIDO"`               // Nome do partido do candidato.
	NomeColigacao       string `csv:"NM_COLIGACAO"`             // Nome da coligação a qual o candidato pertence.
	PartidosColigacao   string `csv:"DS_COMPOSICAO_COLIGACAO"`  // Partidos pertencentes à coligação do candidato.
	DeclarouBens        string `csv:"ST_DECLARAR_BENS"`         // Flag que informa se o candidato declarou seus bens na eleição.s
	Situacao            string `csv:"DS_SIT_TOT_TURNO"`         // Campo que informa como o candidato terminou o primeiro turno da eleição (por exemplo como ELEITO, NÃO ELEITO, ELEITO POR MÉDIA) ou se foi para o segundo turno (ficando com situação SEGUNDO TURNO).
	Turno               int    `csv:"NR_TURNO"`                 // Campo que informa número do turno
	SequencialCandidato string `csv:"SQ_CANDIDATO"`             // ID sequencial do candidato no sistema do TSE
	Candidato
}

func main() {
	source := flag.String("coleta", "", "fonte do arquivo zip do TSE com as candidaturas") // pode ser um path usando protocolo file:// ou http://
	outDir := flag.String("outdir", "", "diretório de saída onde os arquivos descomprimidos serão colocados")
	state := flag.String("estado", "", "estado para ser enriquecido")
	projectID := flag.String("projectid", "", "id do projeto no Google Cloud")
	flag.Parse()
	ctx := context.Background()
	if *source != "" {
		if *outDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if err := collect(*source, *outDir); err != nil {
			log.Fatalf("falha ao executar coleta, erro %q", err)
		}
	} else {
		if *projectID == "" {
			log.Fatal("informe o id de projeto")
		}
		client, err := datastore.NewClient(ctx, *projectID)
		if err != nil {
			log.Fatalf("falha ao criar cliente do Datastore, erro %q", err)
		}
		if err != nil {
			log.Fatalf("falha ao criar cliente de datastore: %v", err)
		}
		if *outDir == "" {
			log.Fatal("informe diretório de saída")
		}
		if *state == "" {
			log.Fatal("informar o estado a ser enriquecido")
		}
		if err := process(*outDir, *state, client); err != nil {
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

func process(outDir, state string, client *datastore.Client) error {
	pathToOpen := ""
	err := filepath.Walk(outDir, func(path string, info os.FileInfo, err error) error {
		if strings.Contains(path, state) {
			pathToOpen = path
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("falha ao passear entre arquivos do diretório %s, erro %q", outDir, err)
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
	if err := saveCandidatures(filteredCandidatures, client); err != nil {
		return fmt.Errorf("falha ao salvar candidaturas no banco, erro %q", err)
	}
	return nil
}

func saveCandidatures(candidatures map[string]*descritor.Candidatura, client *datastore.Client) error {
	for _, candidature := range candidatures {
		candidate, err := findCandidateByLegalCode(candidature.Candidato.CPF, client)
		if err != nil {
			return err
		}
		if candidate == nil { // do not exits
			// TODO implement when exists
		} else {
			// TODO implement when do not exits
		}
	}
	return nil
}

func findCandidateByLegalCode(legalCode string, client *datastore.Client) (*tseutils.Candidate, error) {
	var c []*tseutils.Candidate
	query := datastore.NewQuery(candidaturesCollection).Filter("LegalCode =", legalCode)
	if _, err := client.GetAll(context.Background(), query, &c); err != nil {
		return nil, fmt.Errorf("falha ao buscar candidato por CPF, erro %q", err)
	}
	return c[0], nil
}
