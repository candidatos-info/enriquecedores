package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/candidatos-info/enriquecedores/filestorage"
	"github.com/matryer/try"
)

const (
	maxAttempts = 5 // number of times to retry
)

var (
	re = regexp.MustCompile(`([0-9]+)`) // regexp to extract numbers
)

func main() {
	stateDir := flag.String("inDir", "", "diretório onde as fotos do estado estão")
	destinationDir := flag.String("outDir", "", "local onde ficam os arquivos de candidaturas e fotos")
	year := flag.Int("year", -1, "ano da eleição")
	state := flag.String("state", "", "estado da eleição")
	outputFile := flag.String("outputFile", "", "path do arquivo de saída dos paths dos arquivos de candidaturas")                      // if not passed a new one will be created
	handledPicturesFile := flag.String("handledPicturesFile", "", "path para o arquivo contendo o sequencial ID das fotos processadas") // if not passed a new one will be created
	flag.Parse()
	if *stateDir == "" {
		log.Fatal("informe o diretório onde as fotos do estão estão")
	}
	if *destinationDir == "" {
		log.Fatal("informe o local onde as candidaturas estão")
	}
	if *year == -1 {
		log.Fatal("informe o ano da eleição")
	}
	if *state == "" {
		log.Fatal("informe o estado da eleição")
	}
	var picturesReferenceFile *os.File
	var err error
	if *outputFile == "" {
		protoBuffFileName := fmt.Sprintf("pictures_references-%d-%s.csv", *year, *state)
		picturesReferenceFile, err = os.Create(protoBuffFileName)
		if _, err := picturesReferenceFile.WriteString("s3_url,tse_sequencial_id\n"); err != nil {
			log.Fatalf("falha ao escrever tags no arquivo csv, erro %v\n", err)
		}
		if err != nil {
			log.Fatalf("falha ao criar arquivo com caminhos de protocol buffers, erro %v\n", err)
		}
	} else {
		picturesReferenceFile, err = os.OpenFile(*outputFile, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			log.Fatalf("falha ao abrir arquivo passado para ser usado como arquivo de saída localizado em [%s], erro %v", *outputFile, err)
		}
	}
	defer picturesReferenceFile.Close()
	var handledPictures *os.File
	if *handledPicturesFile == "" {
		handledPictures, err = os.Create(fmt.Sprintf("handled_pictures-%d-%s", *year, *state))
		if err != nil {
			log.Fatalf("falha ao criar arquivo com caminhos de protocol buffers, erro %v\n", err)
		}
	} else {
		handledPictures, err = os.OpenFile(*handledPicturesFile, os.O_APPEND|os.O_RDWR, os.ModeAppend)
		if err != nil {
			log.Fatalf("falha ao abrir arquivo passado para ser usado como cache de fotos já processadas [%s], erro %v", *handledPicturesFile, err)
		}
	}
	defer handledPictures.Close()
	client := filestorage.NewAWSClient()
	if err := process(*state, *stateDir, *destinationDir, client, picturesReferenceFile, handledPictures); err != nil {
		log.Fatalf("falha ao enriquecer fotos, erro %q", err)
	}
}

// it gets as argument the local path where pictures to be processed are placed (stateDir)
// and the storageDir which is the place where candidatures are placed
// and where the pictures will be placed too.
func process(state, stateDir, storageDir string, client filestorage.FileStorage, picturesReferenceFile, handledPictures *os.File) error {
	processedPictures := make(map[string]struct{})
	scanner := bufio.NewScanner(handledPictures)
	for scanner.Scan() {
		processedPictures[scanner.Text()] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("falha ao ler arquivo de cache de fotos processadas [%s], erro %v", handledPictures.Name(), err)
	}
	return filepath.Walk(stateDir, func(path string, info os.FileInfo, err error) error {
		if path != stateDir {
			var googleDriveID string
			fileName := filepath.Base(path)
			sequencialCandidateFromFileName := re.FindAllString(filepath.Base(fileName), -1)[0]
			if _, ok := processedPictures[fileName]; !ok { // checking if picture has already been processed
				filePath := fmt.Sprintf("%s_%s%s", state, sequencialCandidateFromFileName, filepath.Ext(fileName)) // ${ESTADO}_${SEQUENCIAL_CANDIDATE}.${EXTENSION}
				b, err := ioutil.ReadFile(path)
				if err != nil {
					return fmt.Errorf("falha ao ler arquivo [%s], erro %q", path, err)
				}
				err = try.Do(func(attempt int) (bool, error) {
					googleDriveID, err = client.Upload(b, storageDir, filePath)
					return attempt < maxAttempts, err
				})
				if err != nil {
					return fmt.Errorf("falha ao salvar arquivo de candidatura [%s] no bucket [%s], erro %q", filePath, storageDir, err)
				}
				log.Printf("sent file [%s]\n", filePath)
				if _, err := picturesReferenceFile.WriteString(fmt.Sprintf("%s,%s\n", googleDriveID, sequencialCandidateFromFileName)); err != nil {
					log.Fatalf("falha ao escrever tags no arquivo csv, erro %v\n", err)
				}
				if _, err := handledPictures.WriteString(fmt.Sprintf("%s\n", fileName)); err != nil {
					log.Fatalf("falha ao escrever arquivo processado [%s], erro %v\n", sequencialCandidateFromFileName, err)
				}
			} else {
				log.Printf("file [%s] already processed\n", sequencialCandidateFromFileName)
			}
		}
		return nil
	})
}
