package filestorage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

// Client is a client for google cloud storage
type Client struct {
	client *storage.Client
}

// New returns an instance of GCS
func New() (*Client, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao criar client do GCS, erro %q", err)
	}
	return &Client{
		client: client,
	}, nil
}

// UploadFile gets and io.Reader, like a os.File, and uploads
// its content to a bucket accoring with the given path
func (gcs *Client) UploadFile(reader io.Reader, bucket, path string) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()
	wc := gcs.client.Bucket(bucket).Object(path).NewWriter(ctx)
	if _, err := io.Copy(wc, reader); err != nil {
		return fmt.Errorf("falha ao copiar conteúdo de arquivo local para o bucket no GCS, erro %q", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("falha ao fechar storate.Writter object, erro %q", err)
	}
	return nil
}
