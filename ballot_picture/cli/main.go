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
	candidatesDir := flag.String("candidatesDir", "", "local onde estão as as candidaturas") // Se for para usar o gcs usar gs://BUCKET, se for local basta passar o path
	picturesDir := flag.String("outDir", "", "local onde as fotos devem ser salvas")         // Se for para usar o gcs usar gs://BUCKET, se for local basta passar o path
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
	logErrorFile, err := os.Create(logFileName)
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
			if strings.HasPrefix(*candidatesDir, "gs://") { // using GCS
				bucket := strings.ReplaceAll(*candidatesDir, "gs://", "")
				if gcsClient.FileExists(bucket, candidatureFilePath) {
					b, err := ioutil.ReadFile(path)
					if err != nil {
						return fmt.Errorf("falha ao ler arquivo %s, erro %q", path, err)
					}
					if strings.Contains(*picturesDir, "gs://") { // save pictures on gcs
						if err := uploadToGCS(b, bucket, fileName, gcsClient); err != nil {
							return err
						}
						log.Printf("saved file [ %s ]\n", fileName)
					} else {
						if err := saveFileLocally(*picturesDir, fileName, b); err != nil {
							return err
						}
					}
				} else {
					if err := handlePictureNotRelated(sequencialCandidate, logErrorFile); err != nil {
						return err
					}
				}
			} else {
				filePath := fmt.Sprintf("%s/%s", *candidatesDir, fileName)
				if _, err := os.Stat(filePath); err != nil {
					b, err := ioutil.ReadFile(filePath)
					if err != nil {
						return fmt.Errorf("falha ao ler arquivo %s, erro %q", path, err)
					}
					if strings.Contains(*picturesDir, "gs://") { // save pictures on gcs
						bucket := strings.ReplaceAll(*picturesDir, "gs://", "")
						if err := uploadToGCS(b, bucket, fileName, gcsClient); err != nil {
							return err
						}
						log.Printf("saved file [ %s ]\n", fileName)
					} else {
						if err := saveFileLocally(*picturesDir, fileName, b); err != nil {
							return err
						}
					}
				} else {
					if err := handlePictureNotRelated(sequencialCandidate, logErrorFile); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		log.Fatalf("falha ao percorrer arquivos de diretótio %s, erro %q", *stateDir, err)
	}
}

func uploadToGCS(bytes []byte, bucket, filePath string, gcsClient *filestorage.GSCClient) error {
	err := try.Do(func(attempt int) (bool, error) {
		return attempt < maxAttempts, gcsClient.Upload(bytes, bucket, filePath)
	})
	if err != nil {
		return fmt.Errorf("falha ao salvar arquivo de candidatura %s no bucket %s, erro %q", filePath, bucket, err)
	}
	return nil
}

func saveFileLocally(dir, fileName string, bytesToWrite []byte) error {
	path := fmt.Sprintf("%s/%s", dir, fileName)
	filePicture, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("falha ao criar arquivo de foto local %s, erro %q", path, err)
	}
	defer filePicture.Close()
	if _, err := filePicture.Write(bytesToWrite); err != nil {
		return fmt.Errorf("falha ao salvar foto no diretório de saída, erro %q", err)
	}
	return nil
}

func handlePictureNotRelated(sequencialCandidate string, logErrorFile *os.File) error {
	log.Printf("código %s não encontrado no GCS\n", sequencialCandidate)
	newLine := fmt.Sprintf("%s\n", sequencialCandidate)
	if _, err := logErrorFile.WriteString(newLine); err != nil {
		return fmt.Errorf("falha ao escrever que arquivo não encontrado no GCS (%s) no arquivo de log %s, erro %q", sequencialCandidate, logErrorFile.Name(), err)
	}
	return nil
}
