package filestorage

// FileStorage is the package interface
type FileStorage interface {
	Upload(b []byte, bucket, path string) error

	FileExists(bucket, fileName string) bool
}
