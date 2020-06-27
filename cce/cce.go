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
	"strconv"
	"strings"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/candidatos-info/enriquecedores/status"
	"github.com/labstack/echo"
)

const (
	electionYear              = 2  // election year on the csv file is at column 2
	role                      = 14 // candidate role on the csv file is at column 14
	state                     = 10 // state on csv is at column 10
	city                      = 12 // city on csv is at column 12
	ballotNumber              = 16 // ballor number on the csv file is at column 16
	ballotName                = 18 // ballot name on the csv file is at column 18
	ability                   = 23 // candidature ability on csv file is at column 23
	deffering                 = 25 // candidate deffering on csv is at column 25
	candidatureGroup          = 26 // candidature group on csv is at column 26
	partyNumber               = 27 // party number on csv is at column 27
	partyLabel                = 28 // party label on csv is at column 28
	partyName                 = 29 // party name on csv is at column 29
	partyGroupName            = 31 // party group name on csv is at column 31
	groupParties              = 32 // group parties on csv is at column 32
	didDeclaredPossessions    = 55 // that column that indicates if candidate has declared possesions is 55
	finalSituation            = 53 // the candidate final situation at end of election round
	candidateState            = 35 // candidate state on csv is at column 35
	candidateCity             = 37 // candidate city on csv is at column 36
	candidateBirth            = 38 // candidate birth on csv is at column 38
	candidateVotingID         = 40 // candidate voting id on csv is at column 40
	candidateGender           = 42 // candidate gender is at column 42
	candidateInstructionLevel = 44 // candidate instruction level is at column 44
	candidateCivilState       = 46 // candidate civil state is at column 46
	candidateRace             = 48 // candidate race is at column 48
	candidateOccupation       = 50 // candidate occupation is at column 50
	candidateName             = 17 // candidate name on the csv file is at column 17
	candidateCPF              = 20 // canidate cpf on csv is at column 20
	candidateEmail            = 21 // candidate email on csv is at column 21
)

var (
	rolesMap = map[string]descritor.Cargo{
		"VEREADOR":      "LM",
		"VICE-PREFEITO": "EM",
		"PREFEITO":      "EM",
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
	if strings.HasPrefix(h.CandidaturesPath, "gc://") {
		// TODO add GCS implementation
	} else {
		if err := executeForLocal(ha, buf, h); err != nil {
			handleError(fmt.Sprintf("falha executar processamento local, erro: %q", err), h)
			return
		}
	}
	if err = os.RemoveAll(unzipDestination); err != nil {
		handleError(fmt.Sprintf("falha ao remover diretorio temporario criado, erro %q", err), h)
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
	downloadedFiles, err := unzipDownloadedFiles(buf, h.UnzippedFilesDir)
	if err != nil {
		return fmt.Errorf("falha ao descomprimir arquivos baixados, erro %q", err)
	}
	for _, filePath := range downloadedFiles {
		// TODO parallelize it using goroutines
		file, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("falha ao abrir arquivo .csv descomprimido %s, erro %q", file.Name(), err)
		}
		defer file.Close()
		csvReader := csv.NewReader(file)
		csvReader.Comma = ';'
		csvReader.LazyQuotes = true
		candidates, err := getCandidates(csvReader, file.Name())
		if err != nil {
			return fmt.Errorf("falha ao criar lista de candidaturas, erro %q", err)
		}
		fmt.Println(candidates)
		// TODO save candidates into a local file
	}
	return nil
}

// it iterates through csv lines and returns a map of
// struct *descritor.Candidatura where the key is the canidate CPF.
// To handle the duplicated canidate data lines is used the canidadate
// CPF as search key
func getCandidates(reader *csv.Reader, fileName string) (map[string]*descritor.Candidatura, error) {
	candidatesMap := make(map[string]*descritor.Candidatura)
	currentLine := 0
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("falha ao ler arquivo csv %s, erro %q", fileName, err)
		}
		if currentLine > 0 {
			currentCPF := line[candidateCPF]
			foundCandidate := candidatesMap[currentCPF]
			// if candidate was not found, create it and add it to candidates map
			if foundCandidate == nil {
				candidate, err := createCandidatureFromCSVLine(line)
				if err != nil {
					return nil, fmt.Errorf("falha ao criar struct de candidatura para arquivo %s na linha %d, erro %q", fileName, currentLine, err)
				}
				candidatesMap[currentCPF] = candidate
			} else { // if candidates is alredy in canidates map, update the first or second round situation
				legislatura, err := strconv.Atoi(line[electionYear])
				if err != nil {
					return nil, fmt.Errorf("falha ao extrair o ano de legislatura, erro %q", err)
				}
				if legislatura == 1 {
					foundCandidate.SituacaoPrimeiroTurno = line[finalSituation]
				} else {
					foundCandidate.SituacaoSegundoTurno = line[finalSituation]
				}
			}
		}
		currentLine++
	}
	return candidatesMap, nil
}

// extracts data from csv columns, treat them and
// add to strurct descritor.Candidatura. For simplicity reasons
// is being used portuguese language for the attribute names
func createCandidatureFromCSVLine(line []string) (*descritor.Candidatura, error) {
	legislatura, err := strconv.Atoi(line[electionYear])
	if err != nil {
		return nil, fmt.Errorf("falha ao extrair o ano de legislatura, erro %q", err)
	}
	numeroUrna, err := strconv.Atoi(line[ballotNumber])
	if err != nil {
		return nil, fmt.Errorf("falha ao extrair o número de urna, erro %q", err)
	}
	numeroPartido, err := strconv.Atoi(line[partyNumber])
	if err != nil {
		return nil, fmt.Errorf("falha ao extrair número do partido, erro %q", err)
	}
	nascimentoCandidato, err := time.Parse("02/01/2006", line[candidateBirth])
	if err != nil {
		return nil, fmt.Errorf("falha ao fazer parse da data de nascimento do candidato, erro %q", err)
	}
	candidatura := &descritor.Candidatura{
		Legislatura:       legislatura,
		Cargo:             rolesMap[line[role]],
		UF:                line[state],
		Municipio:         line[city],
		NumeroUrna:        numeroUrna,
		NomeUrna:          line[ballotName],
		Aptidao:           line[ability],
		Deferimento:       line[deffering],
		TipoAgremiacao:    line[candidatureGroup],
		NumeroPartido:     numeroPartido,
		LegendaPartido:    line[partyLabel],
		NomePartido:       line[partyName],
		NomeColigacao:     line[partyGroupName],
		PartidosColigacao: line[groupParties],
		DeclarouBens:      declaredPossessions[line[didDeclaredPossessions]],
		Candidato: descritor.Candidato{
			UF:              line[candidateState],
			Municipio:       line[candidateCity],
			Nascimento:      nascimentoCandidato,
			TituloEleitoral: line[candidateVotingID],
			Genero:          line[candidateGender],
			GrauInstrucao:   line[candidateInstructionLevel],
			EstadoCivil:     line[candidateCivilState],
			Raca:            line[candidateRace],
			Ocupacao:        line[candidateOccupation],
			CPF:             line[candidateCPF],
			Nome:            line[candidateName],
			Email:           line[candidateEmail],
		},
	}
	if candidatura.Legislatura == 1 {
		candidatura.SituacaoPrimeiroTurno = line[finalSituation]
	} else {
		candidatura.SituacaoSegundoTurno = line[finalSituation]
	}
	return candidatura, nil
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
