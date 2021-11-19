package ports

import (
	"github.com/seggga/practice/internal/domain"
)

type FileSystem interface {
	FindSubfolders(path string) ([]string, error)
	FindFiles(dirSlice []string) ([]domain.File, error)
	RemoveFile(domain.File) error
}

type RemoverService interface {
	FindFiles(path string) error
	GetClones(files []domain.File) error
	RemoveClones(uniqFilePath string)
}

type Storager interface {
	StoreFiles([]domain.File)
	GetClones() error
	GetDeletable()
	ReadFiles() []domain.File
	RemoveFile(domain.File)
}
