package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/candidatos-info/descritor"
	tseutils "github.com/candidatos-info/enriquecedores/tse_utils"

	"github.com/gocarina/gocsv"
	"golang.org/x/text/encoding/charmap"
)

func main() {
	prevElectionCandidaturesSheet := flag.String("prevCandidaturesSheet", "", "path para arquivo de candidaturas da eleição anterior")
	currentElectionCandidaturesSheet := flag.String("currentCandidaturesSheet", "", "path para arquivo de candidaturas atual")
	dbName := flag.String("dbName", "", "nome do banco de dados")
	dbURL := flag.String("dbURL", "", "URL de conexão com banco de dados")
	flag.Parse()
	if *prevElectionCandidaturesSheet == "" {
		log.Fatal("informe caminho do arquivo de candidaturas da eleição passada")
	}
	if *currentElectionCandidaturesSheet == "" {
		log.Fatal("informe o caminho do arquivo de candidaturas da eleição atual")
	}
	if *dbName == "" {
		log.Fatal("informe o nome do banco")
	}
	if *dbURL == "" {
		log.Fatal("informe a URL de conexão com o banco")
	}
	dbClient, err := New(*dbURL, *dbName)
	if err != nil {
		log.Fatalf("falha ao se conectar com banco, error %v", err)
	}
	if err := process(*prevElectionCandidaturesSheet, *currentElectionCandidaturesSheet, dbClient); err != nil {
		log.Fatalf("falha ao processar verificação de recorrência, error %v", err)
	}
}

func process(prevElectionCandidaturesSheet, currentElectionCandidaturesSheet string, dbClient *Client) error {
	prevCandidatures, err := getCandidates(prevElectionCandidaturesSheet)
	if err != nil {
		return err
	}
	currentCandidatures, err := getCandidates(currentElectionCandidaturesSheet)
	if err != nil {
		return err
	}
	for candidateLegalCode, candidature := range prevCandidatures {
		if _, ok := currentCandidatures[candidateLegalCode]; ok {
			if err := dbClient.UpdateCandidate(int(candidature.Legislatura), candidateLegalCode, true); err != nil {
				return err
			}
			log.Printf("candidate with legal code [%s] is recurrent\n", candidateLegalCode)
		} else {
			log.Printf("candidate with legal code [%s] is NOT recurrent\n", candidateLegalCode)
		}
	}
	return nil
}

// it reads the candidatures file, remove the duplicated candidatures
// and returns them grouped on a map where the key is the candidate legal code
func getCandidates(candidaturesFile string) (map[string]*descritor.Candidatura, error) {
	file, err := os.Open(candidaturesFile)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir arquivo de candidaturas da eleição [%s], erro %v", candidaturesFile, err)
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
		return nil, fmt.Errorf("falha ao inflar slice de candidaturas usando arquivo [%s], erro %v", candidaturesFile, err)
	}
	filteredCandidatures, err := tseutils.RemoveDuplicates(c, candidaturesFile)
	if err != nil {
		return nil, fmt.Errorf("falha ao remover candidaturas duplicadas do arquivo [%s], erro %v", candidaturesFile, err)
	}
	return filteredCandidatures, nil
}
