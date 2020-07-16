package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/candidatos-info/enriquecedores/filestorage"
)

func main() {
	stateDir := flag.String("inDir", "", "diretório onde as fotos do estado estão")
	candidatesDir := flag.String("candidatesDir", "", "local onde estão as as candidaturas") // Se for para usar o gcs usar gs://BUCKET/ANO, se for local basta passar o path
	picturesDir := flag.String("outDir", "", "local onde as fotos devem ser salvas")         // Se for para usar o gcs usar gs://BUCKET/ANO, se for local basta passar o path
	flag.Parse()
	if *stateDir == "" {
		log.Fatal("informe o diretório onde as fotos do estão estão")
	}
	if *candidatesDir == "" {
		log.Fatal("informe o local onde as candidaturas estão")
	}
	if *picturesDir == "" {
		log.Fatal("informe o local onde as fotos devem ser salvs")
	}
	gcsClient, err := filestorage.NewGCSClient()
	if err != nil {
		log.Fatalf("falha ao criar cliente do Google Cloud Storage, erro %q", err)
	}
	err = filepath.Walk(*stateDir, func(path string, info os.FileInfo, err error) error {
		//candidateSequencialCode := strings.TrimSuffix(fileName, filepath.Ext(fileName))
		if path != *stateDir {
			if strings.Contains(*candidatesDir, "gs://") { // candidatures on GCS
				bucket, filePathOnGCS := getBucketAndFilePath(*candidatesDir, path)
				fmt.Println(gcsClient)
				fmt.Println(bucket, filePathOnGCS)
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("falha ao percorrer arquivos de diretótio %s, erro %q", *stateDir, err)
	}
}

func getBucketAndFilePath(candidatesDir, path string) (string, string) {
	elements := strings.Split(candidatesDir, "/")
	electionYear := elements[3]
	bucket := elements[2]
	fileName := filepath.Base(path)
	sequencialCandidate := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return bucket, fmt.Sprintf("%s/%s.zip", electionYear, sequencialCandidate)
}
