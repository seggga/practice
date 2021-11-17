package domain

// File represents data about given file
type File struct {
	Name        string
	SizeInBytes int
	Path        string
	Dir         string
	CloneID     int
	// Hash    string

}
