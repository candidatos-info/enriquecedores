package tseutils

import (
	"fmt"
	"log"
	"time"

	"github.com/candidatos-info/descritor"
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
func RemoveDuplicates(candidates []*Candidatura, fileBeingHandled string) (map[string]*descritor.Candidatura, error) {
	candidatesMap := make(map[string]*descritor.Candidatura)
	fileLines := 0
	duplicatedLines := 0
	for _, c := range candidates {
		foundCandidate := candidatesMap[c.CPF]
		if foundCandidate == nil { // candidate not present on map, add it
			fileLines++
			var candidateBirth time.Time
			if c.Candidato.Nascimento != "" {
				candidateBirth, err := time.Parse("02/01/2006", c.Candidato.Nascimento)
				if err != nil {
					return nil, fmt.Errorf("falha ao fazer parse da data de nascimento do candidato [%s] para o formato 02/01/2006, erro %q", c.Candidato.Nascimento, err)
				}
				_ = candidateBirth
			}
			newCandidate := &descritor.Candidatura{
				Legislatura:         c.Legislatura,
				Cargo:               rolesMap[c.Cargo],
				UF:                  c.UF,
				Municipio:           c.Municipio,
				NomeUrna:            c.NomeUrna,
				Aptidao:             c.Aptidao,
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
				Candidato: descritor.Candidato{
					UF:              c.Candidato.UF,
					Municipio:       c.Candidato.Municipio,
					Nascimento:      candidateBirth,
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
