package filestorage

// FileStorage is the package interface
type FileStorage interface {
	Upload(b []byte, bucket, fileName string) error

	FileExists(bucket, fileName string) bool
}
