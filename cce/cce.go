package cce

import (
	"archive/zip"
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
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/enriquecedores/status"
	"github.com/gocarina/gocsv"
	"github.com/labstack/echo"
)

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
	Legislatura       int    `csv:"ANO_ELEICAO"`              // Ano eleitoral em que a candidatura foi homologada.
	Cargo             string `csv:"DS_CARGO"`                 // Cargo sendo pleiteado pela candidatura.
	UF                string `csv:"SG_UF"`                    // Identificador (2 caracteres) de unidade federativa onde ocorreu a candidatura.
	Municipio         string `csv:"NM_UE"`                    // Município que ocorreu a eleição.
	NumeroUrna        int    `csv:"NR_CANDIDATO"`             // Número do candidato na urna.
	NomeUrna          string `csv:"NM_URNA_CANDIDATO"`        // Nome do candidato na urna.
	Aptidao           string `csv:"DS_SITUACAO_CANDIDATURA"`  // Aptidao da candidatura (podendo ser APTO ou INAPTO).
	Deferimento       string `csv:"DS_DETALHE_SITUACAO_CAND"` // Situação do candidato (pondendo ser DEFERIDO ou INDEFERIDO).
	TipoAgremiacao    string `csv:"TP_AGREMIACAO"`            // Indica o tipo de agremiação do candidato (podendo ser PARTIDO ISOLADO ou AGREMIAÇÃO).
	NumeroPartido     int    `csv:"NR_PARTIDO"`               // Número do partido do candidato.
	LegendaPartido    string `csv:"SG_PARTIDO"`               // Legenda do partido do candidato.
	NomePartido       string `csv:"NM_PARTIDO"`               // Nome do partido do candidato.
	NomeColigacao     string `csv:"NM_COLIGACAO"`             // Nome da coligação a qual o candidato pertence.
	PartidosColigacao string `csv:"DS_COMPOSICAO_COLIGACAO"`  // Partidos pertencentes à coligação do candidato.
	DeclarouBens      string `csv:"ST_DECLARAR_BENS"`         // Flag que informa se o candidato declarou seus bens na eleição.s
	Situacao          string `csv:"DS_SIT_TOT_TURNO"`         // Campo que informa como o candidato terminou o primeiro turno da eleição (por exemplo como ELEITO, NÃO ELEITO, ELEITO POR MÉDIA) ou se foi para o segundo turno (ficando com situação SEGUNDO TURNO).
	Turno             int    `csv:"NR_TURNO"`                 // Campo que informa número do turno
	Candidato
}

var (
	rolesMap = map[string]descritor.Cargo{
		"VEREADOR":      "LM",  // Legislativo Municipal
		"VICE-PREFEITO": "VEM", // Vice Executivo Municipal
		"PREFEITO":      "EM",  // Executivo Municipal
	}

	declaredPossessions = map[string]bool{
		"S": true,
		"N": false,
	}
)

// Handler is a struct to hold important data for this package
type Handler struct {
	SourceURL        string        `json:"source_url"`         // URL to retrieve files. It can be a path for a file or an URL
	Status           status.Status `json:"status"`             // enrich status
	Err              string        `json:"err"`                // last error message
	SourceFileHash   string        `json:"source_file_hash"`   // hash of last downloaded .zip file
	SourceLocalPath  string        `json:"source_local_path"`  // the path where downloaded files should stay
	CandidaturesPath string        `json:"candidatures_path"`  // the place where candidatures files will stay
	UnzippedFilesDir string        `json:"unzipped_files_dir"` // temporary directory where unzipped files ares placed
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
	h.SourceFileHash = ha
	if err != nil {
		handleError(fmt.Sprintf("falha ao gerar hash de arquivo do TSE baixado, erro: %q", err), h)
		return
	}
	unzipDestination, err := ioutil.TempDir("", "unzipped")
	if err != nil {
		handleError(fmt.Sprintf("falha ao criar diretório temporário unzipped, erro: %q", err), h)
		return
	}
	h.UnzippedFilesDir = unzipDestination
	hashFile, err := resolveHashFile(h.SourceURL)
	if err != nil {
		handleError(fmt.Sprintf("falha ao resolver arquivo de hash, erro %q", err), h)
		return
	}
	hashFileBytes, err := ioutil.ReadAll(hashFile)
	if err != nil {
		handleError(fmt.Sprintf("falha ao ler os bytes do arquivo %s, erro: %q", hashFile.Name(), err), h)
		return
	}
	if ha == string(hashFileBytes) {
		log.Printf("arquivo baixado é o mesmo (possui o mesmo hash %s)\n", ha)
		h.Status = status.Idle
		return
	}
	h.Status = status.Processing
	downloadedFiles, err := unzipDownloadedFiles(buf, h.UnzippedFilesDir)
	if err != nil {
		handleError(fmt.Sprintf("falha ao descomprimir arquivos baixados, erro %q", err), h)
		return
	}
	var candidates []*descritor.Candidatura
	for _, filePath := range downloadedFiles {
		// TODO parallelize it using goroutines
		file, err := os.Open(filePath)
		if err != nil {
			handleError(fmt.Sprintf("falha ao abrir arquivo .csv descomprimido %s, erro %q", file.Name(), err), h)
			return
		}
		defer file.Close()
		gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
			r := csv.NewReader(in)
			r.LazyQuotes = true
			r.Comma = ';'
			return r
		})
		var c []*Candidatura
		if err := gocsv.UnmarshalFile(file, &c); err != nil {
			handleError(fmt.Sprintf("falha ao inflar slice de candidaturas usando arquivo csv %s, erro %q", file.Name(), err), h)
			return
		}
		filteredCandidatures, err := removeDuplicates(c)
		if err != nil {
			handleError(fmt.Sprintf("falha ao remover candidaturas duplicadas, erro %q", err), h)
			return
		}
		for _, fc := range filteredCandidatures {
			candidates = append(candidates, fc)
		}
	}
	if strings.HasPrefix(h.CandidaturesPath, "gc://") {
		// TODO add GCS implementation
	} else {
		if err := saveCandidatesLocal(candidates, h.CandidaturesPath); err != nil {
			handleError(fmt.Sprintf("falha ao salvar arquivos de candidaturas localmente, erro: %q", err), h)
			return
		}
	}
	if err = os.RemoveAll(unzipDestination); err != nil {
		handleError(fmt.Sprintf("falha ao remover diretorio temporario criado %s, erro %q", unzipDestination, err), h)
	}
}

// save candidatures localy on the given path
func saveCandidatesLocal(candidates []*descritor.Candidatura, pathToSave string) error {
	return nil
}

// it iterates through the candidates list and returns a map of
// struct *descritor.Candidatura where the key is the candidate CPF.
// To handle the duplicated canidate data lines is used the candidate
// CPF as search key.
// This is necessary due to the fact that TSE CSV duplicate candidate's
// data if it goes to the election second round, changing only two columns:
// the election round (NR_TURNO) and the candidature situation (DS_SIT_TOT_TURNO).
// This function takes care of it by collecting the candidate only once and
// registering if it has gone or not to election second round.
func removeDuplicates(candidates []*Candidatura) (map[string]*descritor.Candidatura, error) {
	candidatesMap := make(map[string]*descritor.Candidatura)
	for _, c := range candidates {
		foundCandidate := candidatesMap[c.CPF]
		if foundCandidate == nil { // candidate not present on map, add it
			nascimentoCandidato, err := time.Parse("02/01/2006", c.Candidato.Nascimento)
			if err != nil {
				return nil, fmt.Errorf("falha ao fazer parse da data de nascimento do candidato %s para o formato 02/01/2006, erro %q", c.Candidato.Nascimento, err)
			}
			newCandidate := &descritor.Candidatura{
				Legislatura:       c.Legislatura,
				Cargo:             rolesMap[c.Cargo],
				UF:                c.UF,
				Municipio:         c.Municipio,
				NomeUrna:          c.NomeUrna,
				Aptidao:           c.Aptidao,
				Deferimento:       c.Deferimento,
				TipoAgremiacao:    c.TipoAgremiacao,
				NumeroPartido:     c.NumeroUrna,
				LegendaPartido:    c.LegendaPartido,
				NomePartido:       c.NomePartido,
				NomeColigacao:     c.NomeColigacao,
				PartidosColigacao: c.PartidosColigacao,
				DeclarouBens:      declaredPossessions[c.DeclarouBens],
				Candidato: descritor.Candidato{
					UF:              c.Candidato.UF,
					Municipio:       c.Candidato.Municipio,
					Nascimento:      nascimentoCandidato,
					TituloEleitoral: c.Candidato.TituloEleitoral,
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
			if c.Turno == 1 {
				newCandidate.SituacaoPrimeiroTurno = c.Situacao
			} else {
				newCandidate.SituacaoSegundoTurno = c.Situacao
			}
			candidatesMap[c.CPF] = newCandidate
		} else { // candidate already on map (maybe election second round)
			if c.Turno == 1 {
				foundCandidate.SituacaoPrimeiroTurno = c.Situacao
			} else {
				foundCandidate.SituacaoSegundoTurno = c.Situacao
			}
		}
	}
	return candidatesMap, nil
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
