package filestorage

import (
	"fmt"
	"io/ioutil"
	"os"
)

// GSCClient is a client for google cloud storage
type localStorage struct {
}

// NewLocalStorage returns a new local storage instance
func NewLocalStorage() FileStorage {
	return &localStorage{}
}

// Upload gets and io.Reader, like a os.File, and uploads
// its content to a bucket accoring with the given path
func (gcs *localStorage) Upload(b []byte, bucket, fileName string) (string, error) {
	_, err := os.Stat(bucket) // checking if bucket exists
	if os.IsNotExist(err) {
		err := os.MkdirAll(bucket, 0755)
		if err != nil {
			return "", fmt.Errorf("falha ao criar diretório %s, erro %q", bucket, err)
		}
	}
	name := fmt.Sprintf("%s/%s", bucket, fileName)
	if err := ioutil.WriteFile(name, b, 0644); err != nil {
		return "", fmt.Errorf("falha ao salver arquivo %s no caminho %s, erro %q", fileName, name, err)
	}
	return fmt.Sprintf("%s/%s", bucket, fileName), nil
}

// FileExists checks if file exists. If file exists
// it returns true, else false
func (gcs *localStorage) FileExists(bucket, fileName string) bool {
	_, err := os.Stat(fileName)
	return err == nil
}
