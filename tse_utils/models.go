package tseutils

import "github.com/candidatos-info/descritor"

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

// Candidate is the struct for a candidate on db
type Candidate struct {
	Email        string
	State        string
	City         string
	Name         string
	Party        string
	BallotNumber int
	LegalCode    string
	Password     string
}

// CandidatureData is the data related to one election
type CandidatureData struct {
	CandidateLegalCode string
	SquencialCandidate string
	SiteURL            string
	Facebook           string
	Twitter            string
	Year               int
	Proposal           string
}

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
