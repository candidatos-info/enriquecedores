package status

// Status is a custom type to represent the possible status
type Status int

const (
	// Idle means the the service is avaiable to new processments
	Idle Status = 0

	// Collecting means that data is being collected
	Collecting Status = 1

	// Hashing means the downloaded file to preocess is being hashed
	Hashing Status = 2

	// Processing means that data is being processed
	Processing Status = 3
)

var (
	statusText = map[Status]string{
		Idle:       "System is idle",
		Collecting: "System is collecting data",
		Hashing:    "Getting hash of downloaded file",
		Processing: "System is processing data",
	}
)

// Text returns a text for a status. It returns the empty
// string if the status is unknown.
func Text(status Status) string {
	return statusText[status]
}
