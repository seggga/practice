package cloremover

import (
	"sort"

	"github.com/seggga/practice/internal/domain"
	"github.com/seggga/practice/internal/ports"
	"go.uber.org/zap"
)

type service struct {
	fs      ports.FileSystem
	storage ports.Storager
	dir     string
	logger  *zap.SugaredLogger
	// mutex sync.Mutex
}

func New(fs ports.FileSystem, stor ports.Storager, logger *zap.SugaredLogger) *service {
	return &service{
		fs:      fs,
		storage: stor,
		logger:  logger,
		// mutex: sync.Mutex{},
	}
}

func (srv *service) FindFiles(path string) error {
	// obtain all the subfolders in the given path
	dirSlice, err := srv.fs.FindSubfolders(path)
	if err != nil {
		srv.logger.Errorf("error finding subfolders, %v", err)
		return err
	}
	// obtain all the files in all subfolders
	fileSlice, err := srv.fs.FindFiles(dirSlice)
	if err != nil {
		srv.logger.Errorf("error finding files, %v", err)
		return err
	}
	// store files data in the storage
	srv.storage.StoreFiles(fileSlice)
	return nil
}

// LeaveClones filters out all unique files from slice
func (srv *service) GetClones() error {
	srv.storage.GetClones()
	return nil
}

// sortFiles sorts by CloneID
func sortByID(sl []domain.File) []domain.File {
	sort.Slice(sl, func(i, j int) bool {
		return sl[i].CloneID < sl[j].CloneID
	})
	return sl
}

// sortByPath sorts by Path
func sortByPath(sl []domain.File) []domain.File {
	sort.Slice(sl, func(i, j int) bool {
		return sl[i].Path < sl[j].Path
	})
	return sl
}

// RemoveCLones calls file removal and corresponding storage data change
func (srv *service) RemoveClones() error {

	clones := srv.storage.ReadFiles()
	clones = sortByID(clones)

	var cloneID = 0
	cloneGroup := make([]domain.File, 0)
	for i, fileData := range clones {
		// the last element in slice
		if i == len(clones)-1 {
			cloneGroup = append(cloneGroup, fileData)
			cloneID = 0
		}
		if fileData.CloneID != cloneID {
			// new group of clones
			if len(cloneGroup) != 0 {
				// cloneGroup contains all identical files
				// delete all files except the first (with the smallest path)
				cloneGroup = sortByPath(cloneGroup)
				for i := 1; i < len(cloneGroup); i += 1 {
					// remove file from file system
					if err := srv.fs.RemoveFile(cloneGroup[i]); err != nil {
						// log
						return err
					}
					// remove data about deleted file from the storage
					srv.storage.RemoveFile(cloneGroup[i])
					// log
				}
				// clear cloneGroup
				cloneGroup = cloneGroup[:0]
				continue
			}
			// create a new slice of clones
			cloneID = fileData.CloneID
		}
		cloneGroup = append(cloneGroup, fileData)
	}
	return nil
}
