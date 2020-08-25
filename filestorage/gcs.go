package filestorage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"cloud.google.com/go/storage"
)

const (
	timeout = time.Second * 50
)

// GSCClient is a client for google cloud storage
type GSCClient struct {
	client *storage.Client
}

// NewGCSClient returns an instance of GCS. It
// will panic if occured failure on create a new client
func NewGCSClient() FileStorage {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatalf("falha ao criar client do GCS, erro %q", err)
	}
	return &GSCClient{
		client: client,
	}
}

// Upload gets and io.Reader, like a os.File, and uploads
// its content to a bucket accoring with the given path
func (gcs *GSCClient) Upload(b []byte, bucket, fileName string) error {
	r := bytes.NewReader(b)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	wc := gcs.client.Bucket(bucket).Object(fileName).NewWriter(ctx)
	if _, err := io.Copy(wc, r); err != nil {
		return fmt.Errorf("falha ao copiar conteúdo de arquivo local para o bucket no GCS (%s/%s), erro %q", bucket, fileName, err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("falha ao fechar storate.Writter object (%s/%s), erro %q", bucket, fileName, err)
	}
	return nil
}
