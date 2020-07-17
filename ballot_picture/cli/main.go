package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/candidatos-info/enriquecedores/filestorage"
	"github.com/matryer/try"
)

const (
	maxAttempts = 5 // number of times to retry
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
	logFileName := fmt.Sprintf("%s.txt", filepath.Base(*stateDir))
	logErrorFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("falha ao criar arquivo de fotos com falha %s, erro %q", logFileName, err)
	}
	defer logErrorFile.Close()
	err = filepath.Walk(*stateDir, func(path string, info os.FileInfo, err error) error {
		if path != *stateDir {
			fileName := filepath.Base(path)
			fileExtension := filepath.Ext(fileName)
			sequencialCandidate := strings.TrimSuffix(fileName, fileExtension)
			candidatureFilePath := fmt.Sprintf("%s.zip", sequencialCandidate)
			if strings.Contains(*candidatesDir, "gs://") { // using GCS
				bucket := strings.ReplaceAll(*candidatesDir, "gs://", "")
				if gcsClient.FileExists(bucket, candidatureFilePath) {
					picturePathOnGCS := fmt.Sprintf("%s%s", sequencialCandidate, fileExtension)
					b, err := ioutil.ReadFile(path)
					if err != nil {
						return fmt.Errorf("falha ao ler arquivo %s, erro %q", path, err)
					}
					if strings.Contains(*picturesDir, "gs://") { // save pictures on gcs
						err = try.Do(func(attempt int) (bool, error) {
							return attempt < maxAttempts, gcsClient.Upload(b, bucket, picturePathOnGCS)
						})
						if err != nil {
							return fmt.Errorf("falha ao salvar arquivo de candidatura %s no bucket %s, erro %q", picturePathOnGCS, bucket, err)
						}
					} else {
						// TODO save pictures localy
					}
				} else {
					newLine := fmt.Sprintf("%s\n", fileName)
					if _, err := logErrorFile.WriteString(newLine); err != nil {
						return fmt.Errorf("falha ao escrever que arquivo não encontrado no GCS (%s) no arquivo de log %s, erro %q", sequencialCandidate, logFileName, err)
					}
				}
			} else {
				// TODO implement for local
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("falha ao percorrer arquivos de diretótio %s, erro %q", *stateDir, err)
	}
}
