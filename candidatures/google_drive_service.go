package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

type googleDrive struct {
	service *drive.Service
}

func newGoogleDriveStorage(credentialsFile, oauthToken string) (storageService, error) {
	b, err := ioutil.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("falha ao ler arquivo de crendenciais [%s], erro %q", credentialsFile, err)
	}
	config, err := google.ConfigFromJSON(b, drive.DriveScope)
	if err != nil {
		log.Fatalf("falha ao processar configuraçōes usando o arquivo [%s], erro %q", credentialsFile, err)
	}
	f, err := os.Open(oauthToken)
	if err != nil {
		log.Fatalf("falha ao abrir arquivo de token oauth [%s], erro %q", oauthToken, err)
	}
	defer f.Close()
	tok := &oauth2.Token{}
	if err = json.NewDecoder(f).Decode(tok); err != nil {
		log.Fatalf("falha ao fazer bind do token OAuth, erro %q", err)
	}
	client := config.Client(context.Background(), tok)
	service, err := drive.New(client)
	if err != nil {
		return nil, fmt.Errorf("não foi possível criar Google Drive service, erro %q", err)
	}
	return &googleDrive{
		service: service,
	}, nil
}

// the bucket argument for Google Drive is the folder ID.
func (gd *googleDrive) Upload(b []byte, bucket, fileName string) error {
	f := &drive.File{
		MimeType: "application/octet-stream",
		Name:     fileName,
		Parents:  []string{bucket},
	}
	buffer := new(bytes.Buffer)
	if _, err := buffer.Write(b); err != nil {
		return fmt.Errorf("falha ao copiar conteúdo de arquivo para buffer temporário, erro %q", err)
	}
	if _, err := gd.service.Files.Create(f).Media(buffer).Do(); err != nil {
		return fmt.Errorf("falha ao criar quivo [%s] na pasta [%s] no Google Drive, erro %q", fileName, bucket, err)
	}
	return nil
}
