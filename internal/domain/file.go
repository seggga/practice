package domain

// File represents data about given file
type File struct {
	Name        string // file name
	SizeInBytes int    // file size
	Path        string // path from parent directory to file name
	Dir         string // directory path from from parent
	CloneID     int    // a field to obtain clones
	// Hash    string
}
