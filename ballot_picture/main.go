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
	stateDir := flag.String("inDir", "", "diretório onde as fotos do estado estão")                             // fotos estão em um path local
	destinationDir := flag.String("destinationDir", "", "local onde ficam os arquivos de candidaturas e fotos") // OBS: arquivos de candidaturas e fotos ficam armazenados no mesmo diretório/bucket. Se for para usar o gcs usar gs://BUCKET, se for local basta passar o path
	production := flag.Bool("prod", false, "informe se deve salvar os arquivos localmente ou na nuvem")
	flag.Parse()
	if *stateDir == "" {
		log.Fatal("informe o diretório onde as fotos do estão estão")
	}
	if *destinationDir == "" {
		log.Fatal("informe o local onde as candidaturas estão")
	}
	logFileName := fmt.Sprintf("%s.txt", filepath.Base(*stateDir))
	logErrorFile, err := os.Create(logFileName)
	if err != nil {
		log.Fatalf("falha ao criar arquivo de fotos com falha %s, erro %q", logFileName, err)
	}
	if err := process(*stateDir, *destinationDir, *production, logErrorFile); err != nil {
		log.Fatalf("falha ao enriquecer fotos, erro %q", err)
	}
	defer logErrorFile.Close()
}

// it gets as argument the local path where pictures to be processed are placed (stateDir)
// and the storageDir which is the place where candidatures are placed
// and where the pictures will be placed too.
func process(stateDir, storageDir string, production bool, logErrorFile *os.File) error {
	var filestorageClient filestorage.FileStorage
	if production {
		filestorageClient = filestorage.NewGCSClient()
	} else {
		filestorageClient = filestorage.NewLocalStorage()
	}
	err := filepath.Walk(stateDir, func(path string, info os.FileInfo, err error) error {
		if path != stateDir {
			fileName := filepath.Base(path)
			fileExtension := filepath.Ext(fileName)
			sequencialCandidate := strings.TrimSuffix(fileName, fileExtension)
			candidatureFilePath := fmt.Sprintf("%s.zip", sequencialCandidate)
			if strings.HasPrefix(storageDir, "gs://") { // using GCS (flag PROD true)
				bucket := strings.ReplaceAll(storageDir, "gs://", "")
				if filestorageClient.FileExists(bucket, candidatureFilePath) {
					b, err := ioutil.ReadFile(path)
					if err != nil {
						return fmt.Errorf("falha ao ler arquivo %s, erro %q", path, err)
					}
					if err := saveFiles(b, bucket, fileName, filestorageClient); err != nil {
						return err
					}
				} else {
					if err := handlePictureNotRelated(sequencialCandidate, logErrorFile); err != nil {
						return err
					}
				}
			} else { // (flag PROD false)
				filePath := fmt.Sprintf("%s/%s", storageDir, fileName)
				if _, err := os.Stat(filePath); err != nil {
					b, err := ioutil.ReadFile(filePath)
					if err != nil {
						return fmt.Errorf("falha ao ler arquivo %s, erro %q", path, err)
					}
					if err := saveFiles(b, storageDir, fileName, filestorageClient); err != nil {
						return err
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
		return fmt.Errorf("falha ao percorrer arquivos de diretótio %s, erro %q", stateDir, err)
	}
	return nil
}

func saveFiles(bytes []byte, bucket, filePath string, filestorageClient filestorage.FileStorage) error {
	err := try.Do(func(attempt int) (bool, error) {
		return attempt < maxAttempts, filestorageClient.Upload(bytes, bucket, filePath)
	})
	if err != nil {
		return fmt.Errorf("falha ao salvar arquivo de candidatura [%s] no bucket [%s], erro %q", filePath, bucket, err)
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
