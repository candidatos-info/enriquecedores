package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	stateDir := flag.String("inDir", "", "diretório onde as fotos do estado estão")
	candidaturesBucket := flag.String("candidatesDir", "", "bucket onde estão as as candidaturas")
	picturesBucket := flag.String("outDir", "", "bucket onde as fotos devem ser salvas")
	flag.Parse()
	if *stateDir == "" {
		log.Fatal("informe o diretório onde as fotos do estão estão")
	}
	if *candidaturesBucket == "" {
		log.Fatal("informe o bucket onde as candidaturas estão")
	}
	if *picturesBucket == "" {
		log.Fatal("informe o bucket onde as fotos devem ser salvs")
	}
	err := filepath.Walk(*stateDir, func(path string, info os.FileInfo, err error) error {
		if path != *stateDir {
			fileName := filepath.Base(path)
			candidateSequencialCode := strings.TrimSuffix(fileName, filepath.Ext(fileName))
			possibleCandidatureFile := fmt.Sprintf("%s/%s.json", *candidaturesBucket, candidateSequencialCode)
			fmt.Println(possibleCandidatureFile)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("falha ao percorrer arquivos de diretótio %s, erro %q", *stateDir, err)
	}
}
