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
	UF              string    `json:"uf_origem" csv:"uf_origem"`               // Identificador (2 caracteres) da unidade federativa de nascimento do candidato.
	Municipio       string    `json:"municipio_origem" csv:"municipio_origem"` // Município de nascimento do candidato.
	Nascimento      time.Time `json:"nascimento" csv:"nascimento"`             // Data de nascimento do candidato.
	TituloEleitoral string    `json:"titulo_eleitoral" csv:"titulo_eleitoral"` // Titulo eleitoral do candidato.
	Genero          string    `json:"genero" csv:"genero"`                     // Gênero do candidato (MASCULINO ou FEMININO).
	GrauInstrucao   string    `json:"grau_instrucao" csv:"grau_instrucao"`     // Grau de instrução do candidato.
	EstadoCivil     string    `json:"estado_civil" csv:"estado_civil"`         // Estado civil do candidato.
	Raca            string    `json:"raca" csv:"raca"`                         // Raça do candidato (como BRANCA ou PARDA).
	Ocupacao        string    `json:"ocupacao" csv:"ocupacao"`                 // Ocupação do candidato (como COMERCIANTE e ARTISTA por exemplo).
	CPF             string    `json:"cpf" csv:"cpf"`                           // CPF do candidato.
	Nome            string    `json:"nome" csv:"nome"`                         // Nome de pessoa física do candidato.
	Email           string    `json:"email" csv:"email"`                       // Email do candidato.
}

// Candidatura representa uma
type Candidatura struct {
	Legislatura           int    `json:"leg" csv:"ANO_ELEICAO"`                       // Ano eleitoral em que a candidatura foi homologada.
	Cargo                 string `json:"cargo" csv:"DS_CARGO"`                        // Cargo sendo pleiteado pela candidatura.
	UF                    string `json:"uf" csv:"SG_UF"`                              // Identificador (2 caracteres) de unidade federativa onde ocorreu a candidatura.
	Municipio             string `json:"municipio" csv:"municipio"`                   // Município que ocorreu a eleição.
	NumeroUrna            int    `json:"numero_urna" csv:"numero_urna"`               // Número do candidato na urna.
	NomeUrna              string `json:"nome_urna" csv:"nome_urna"`                   // Nome do candidato na urna.
	Aptidao               string `json:"aptidao" csv:"aptidao"`                       // Aptidao da candidatura (podendo ser APTO ou INAPTO).
	Deferimento           string `json:"deferimento" csv:"deferimento"`               // Situação do candidato (pondendo ser DEFERIDO ou INDEFERIDO).
	TipoAgremiacao        string `json:"tipo_agremiacao" csv:"tipo_agremiacao"`       // Indica o tipo de agremiação do candidato (podendo ser PARTIDO ISOLADO ou AGREMIAÇÃO).
	NumeroPartido         int    `json:"numero_partio" csv:"numero_partido"`          // Número do partido do candidato.
	LegendaPartido        string `json:"legenda_partido" csv:"legenda_partido"`       // Legenda do partido do candidato.
	NomePartido           string `json:"nome_partido" csv:"nome_partido"`             // Nome do partido do candidato.
	NomeColigacao         string `json:"nome_coligacao" csv:"nome_coligacao"`         // Nome da coligação a qual o candidato pertence.
	PartidosColigacao     string `json:"partidos_coligacao" csv:"partidos_coligacao"` // Partidos pertencentes à coligação do candidato.
	DeclarouBens          bool   `json:"declarou_bens" csv:"declarou_bens"`           // Flag que informa se o candidato declarou seus bens na eleição.s
	SituacaoPrimeiroTurno string `json:"situacao_1turno" csv:"situacao_1turno"`       // Campo que informa como o candidato terminou o primeiro turno da eleição (por exemplo como ELEITO, NÃO ELEITO, ELEITO POR MÉDIA) ou se foi para o segundo turno (ficando com situação SEGUNDO TURNO).
	SituacaoSegundoTurno  string `json:"situacao_2turno" csv:"situacao_2turno"`       // Campo que informa como o candidato terminou o segundo turno da eleição, nesse caso só pondendo ficar como ELEITO ou NÃO ELEITO.
	Candidato             `json:"candidato"`
}

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
	fmt.Println("passou")
	h.Status = status.Processing
	if strings.HasPrefix(h.CandidaturesPath, "gc://") {
		// TODO add GCS implementation
	} else {
		if err := executeForLocal(buf, h); err != nil {
			handleError(fmt.Sprintf("falha executar processamento local, erro: %q", err), h)
			return
		}
	}
	if err = os.RemoveAll(unzipDestination); err != nil {
		handleError(fmt.Sprintf("falha ao remover diretorio temporario criado, erro %q", err), h)
	}
}

func executeForLocal(buf []byte, h *Handler) error {
	downloadedFiles, err := unzipDownloadedFiles(buf, h.UnzippedFilesDir)
	fmt.Println(downloadedFiles)
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
		gocsv.SetCSVReader(func(in io.Reader) gocsv.CSVReader {
			r := csv.NewReader(in)
			r.LazyQuotes = true
			r.Comma = ';'
			return r
		})
		var candidates []descritor.Candidatura
		if err := gocsv.UnmarshalFile(file, &candidates); err != nil {
			return fmt.Errorf("falha ao inflar slice de candidaturas usando arquivo csv %s, erro %q", file.Name(), err)
		}
		// _, err = filterCandidates(candidates)
		// if err != nil {
		// 	return fmt.Errorf("falha ao criar lista de candidaturas, erro %q", err)
		// }
		for _, c := range candidates {
			fmt.Println("LEG ", c.Legislatura)
			fmt.Println("Cargo ", c.Cargo)
		}
		// TODO save candidates into a local file
	}
	return nil
}

// it iterates through csv lines and returns a map of
// struct *descritor.Candidatura where the key is the canidate CPF.
// To handle the duplicated canidate data lines is used the canidadate
// CPF as search key
func filterCandidates(candidates []*descritor.Candidatura) (map[string]*descritor.Candidatura, error) {
	candidatesMap := make(map[string]*descritor.Candidatura)
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
