package main

import "testing"

func TestGetBucketAndFilePath(t *testing.T) {
	testCases := []struct {
		candidatesDir    string
		path             string
		expectedBuckect  string
		expectedFilePath string
	}{
		{"gs://profiles/2016", "file/to/pic.png", "profiles", "2016/pic.zip"},
		{"gs://candidates/2020", "path/para/thumb.jpg", "candidates", "2020/thumb.zip"},
	}
	for _, tt := range testCases {
		bucket, filePathOnGCS := getBucketAndFilePath(tt.candidatesDir, tt.path)
		if bucket != tt.expectedBuckect {
			t.Errorf("expected bucket to be [ %s ], got [ %s ]", tt.expectedBuckect, bucket)
		}
		if filePathOnGCS != tt.expectedFilePath {
			t.Errorf("expected file on gcs to be [ %s ], got [ %s ]", tt.expectedFilePath, filePathOnGCS)
		}
	}
}
