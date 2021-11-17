package ports

import (
	"github.com/seggga/practice/internal/domain"
)

type FileSystem interface {
	FindSubfolders(path string) ([]string, error)
	FindFiles(dirSlice []string) ([]domain.File, error)
	ReadFiles()
	GetHash()
	RemoveFile(domain.File) error
	RemoveDirectory()
}

type RemoverService interface {
	FindFiles(path string) error
	GetClones(files []domain.File) error
	RemoveClones(uniqFilePath string)
}

type Storager interface {
	StoreFiles([]domain.File)
	GetClones()
	ReadFiles() []domain.File
	RemoveFile(domain.File)
}
