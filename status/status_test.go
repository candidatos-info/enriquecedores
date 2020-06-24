package status

import "testing"

func TestText(t *testing.T) {
	testCases := []struct {
		name   string
		status Status
		out    string
	}{
		{"Testing idle status", 0, "System is idle"},
		{"Testing collecting status", 1, "System is collecting data"},
		{"Testing hashing status", 2, "Getting hash of downloaded file"},
		{"Testing processing status", 3, "System is processing data"},
		{"Testing unknown status", 505, ""},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			res := Text(tt.status)
			if res != tt.out {
				t.Errorf("want %s, got %s", tt.out, res)
			}
		})
	}
}
