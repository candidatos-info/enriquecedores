package tseutils

import (
	"fmt"
	"log"
	"time"

	"github.com/candidatos-info/descritor"
	"github.com/golang/protobuf/ptypes/timestamp"
)

// RemoveDuplicates iterates through the candidates list and returns a map of
// struct *descritor.Candidatura where the key is the candidate CPF.
// To handle the duplicated canidate data lines is used the candidate
// CPF as search key.
// This is necessary due to the fact that TSE CSV duplicate candidate's
// data if it goes to the election second round, changing only two columns:
// the election round (NR_TURNO) and the candidature situation (DS_SIT_TOT_TURNO).
// This function takes care of it by collecting the candidate only once and
// registering if it has gone or not to election second round.
func RemoveDuplicates(candidates []*RegistroTSE, fileBeingHandled string) (map[string]*descritor.Candidatura, error) {
	candidatesMap := make(map[string]*descritor.Candidatura)
	fileLines := 0
	duplicatedLines := 0
	for _, c := range candidates {
		foundCandidate := candidatesMap[c.CPF]
		if foundCandidate == nil { // candidate not present on map, add it
			fileLines++
			fmt.Println("NASCIMENTO ", c.CandidatoTSE.Nascimento)
			nascimentoAsTime, err := time.Parse("02/01/2006", c.CandidatoTSE.Nascimento)
			if err != nil {
				return nil, fmt.Errorf("falha ao fazer parse de data de nascimento de candidato %s para time.Time, erro %q", c.CandidatoTSE.Nascimento, err)
			}
			s := int64(nascimentoAsTime.Second())
			n := int32(nascimentoAsTime.Nanosecond())
			nascimentoAsTimestamp := &timestamp.Timestamp{Seconds: s, Nanos: n}
			newCandidate := &descritor.Candidatura{
				Aptdao:              c.Aptidao,
				Legislatura:         c.Legislatura,
				Cargo:               rolesMap[c.Cargo],
				UF:                  c.UF,
				Municipio:           c.Municipio,
				NomeUrna:            c.NomeUrna,
				Deferimento:         c.Deferimento,
				TipoAgremiacao:      c.TipoAgremiacao,
				NumeroPartido:       c.NumeroPartido,
				NumeroUrna:          c.NumeroUrna,
				LegendaPartido:      c.LegendaPartido,
				NomePartido:         c.NomePartido,
				NomeColigacao:       c.NomeColigacao,
				PartidosColigacao:   c.PartidosColigacao,
				DeclarouBens:        declaredPossessions[c.DeclarouBens],
				SequencialCandidato: c.SequencialCandidato,
				Candidato: &descritor.Candidato{
					UF:              c.CandidatoTSE.UF,
					Municipio:       c.CandidatoTSE.Municipio,
					Nascimento:      nascimentoAsTimestamp,
					TituloEleitoral: c.CandidatoTSE.TituloEleitoral,
					Genero:          c.CandidatoTSE.Genero,
					GrauInstrucao:   c.CandidatoTSE.GrauInstrucao,
					EstadoCivil:     c.CandidatoTSE.EstadoCivil,
					Raca:            c.CandidatoTSE.Raca,
					Ocupacao:        c.CandidatoTSE.Ocupacao,
					CPF:             c.CandidatoTSE.CPF,
					Nome:            c.CandidatoTSE.Nome,
					Email:           c.CandidatoTSE.Email,
				},
			}
			if c.Turno == 1 {
				newCandidate.SituacaoPrimeiroTurno = c.Situacao
			} else {
				newCandidate.SituacaoSegundoTurno = c.Situacao
			}
			candidatesMap[c.CPF] = newCandidate
		} else { // candidate already on map (maybe election second round)
			duplicatedLines++
			if c.Turno == 1 {
				foundCandidate.SituacaoPrimeiroTurno = c.Situacao
			} else {
				foundCandidate.SituacaoSegundoTurno = c.Situacao
			}
		}
	}
	log.Printf("file [%s], lines [%d], duplicated lines [%d]\n", fileBeingHandled, fileLines, duplicatedLines)
	return candidatesMap, nil
}
