package filestorage

import (
	"bytes"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const (
	region = "sa-east-1"
	acl    = "public-read"
)

// S3Client is a client for AWS S3 service
type S3Client struct {
	uploaded *s3manager.Uploader
}

// NewAWSClient returns a client with implementation for S3.
func NewAWSClient() FileStorage {
	accessKeyID := os.Getenv("ACCESS_KEY_ID")
	if accessKeyID == "" {
		log.Fatal("missing ACCESS_KEY_ID environment variable")
	}
	secretAccessKey := os.Getenv("SECRET_ACCESS_KEY")
	if secretAccessKey == "" {
		log.Fatal("missing SECRET_ACCESS_KEY environment variable")
	}
	config := aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
	}
	sess := session.New(&config)
	return &S3Client{
		uploaded: s3manager.NewUploader(sess),
	}
}

// Upload gets and io.Reader, like a os.File, and uploads
// its content to a bucket accoring with the given path
func (awsClient *S3Client) Upload(b []byte, bucket, fileName string) (string, error) {
	file := new(bytes.Buffer)
	if _, err := file.Write(b); err != nil {
		return "", fmt.Errorf("failed to write bytes of file [%s] on temporary buffer, error %v", fileName, err)
	}
	up, err := awsClient.uploaded.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		ACL:    aws.String(acl),
		Key:    aws.String(fileName),
		Body:   file,
	})
	if err != nil {
		return "", fmt.Errorf("failed to send file [%s] to bucket [%s], error %v", fileName, bucket, err)
	}
	return up.Location, nil
}
