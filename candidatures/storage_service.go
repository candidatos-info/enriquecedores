package main

type storageService interface {
	Upload(b []byte, bucket, fileName string) error
}
