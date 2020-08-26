package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/candidatos-info/enriquecedores/filestorage"
	"github.com/matryer/try"
)

const (
	maxAttempts = 5 // number of times to retry
)

func main() {
	stateDir := flag.String("inDir", "", "diretório onde as fotos do estado estão")                     // fotos estão em um path local
	destinationDir := flag.String("outDir", "", "local onde ficam os arquivos de candidaturas e fotos") // OBS: arquivos de candidaturas e fotos ficam armazenados no mesmo diretório/bucket. Se for para usar o gcs usar gs://BUCKET, se for local basta passar o path
	googleDriveCredentialsFile := flag.String("credentials", "", "chave de credenciais o Goodle Drive")
	goodleDriveOAuthTokenFile := flag.String("OAuthToken", "", "arquivo com token oauth")
	flag.Parse()
	if *stateDir == "" {
		log.Fatal("informe o diretório onde as fotos do estão estão")
	}
	if *destinationDir == "" {
		log.Fatal("informe o local onde as candidaturas estão")
	}
	if err := process(*stateDir, *destinationDir, *googleDriveCredentialsFile, *goodleDriveOAuthTokenFile); err != nil {
		log.Fatalf("falha ao enriquecer fotos, erro %q", err)
	}
}

// it gets as argument the local path where pictures to be processed are placed (stateDir)
// and the storageDir which is the place where candidatures are placed
// and where the pictures will be placed too.
func process(stateDir, storageDir, googleDriveCredentialsFile, goodleDriveOAuthTokenFile string) error {
	var client filestorage.FileStorage
	if googleDriveCredentialsFile != "" && goodleDriveOAuthTokenFile != "" {
		var err error
		client, err = filestorage.NewGoogleDriveStorage(googleDriveCredentialsFile, goodleDriveOAuthTokenFile)
		if err != nil {
			return fmt.Errorf("falha ao criar cliente do Google Drive, erro %q", err)
		}
	} else {
		client = filestorage.NewLocalStorage()
	}
	err := filepath.Walk(stateDir, func(path string, info os.FileInfo, err error) error {
		if path != stateDir {
			fileName := filepath.Base(path)
			filePath := fmt.Sprintf("%s_%s", stateDir, fileName) // ${ESTADO}_${SEQUENCIAL_CANDIDATE}
			b, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("falha ao ler arquivo [%s], erro %q", path, err)
			}
			err = try.Do(func(attempt int) (bool, error) {
				return attempt < maxAttempts, client.Upload(b, storageDir, filePath)
			})
			if err != nil {
				return fmt.Errorf("falha ao salvar arquivo de candidatura [%s] no bucket [%s], erro %q", filePath, storageDir, err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("falha ao percorrer arquivos de diretótio %s, erro %q", stateDir, err)
	}
	return nil
}
