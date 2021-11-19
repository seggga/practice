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
	slogger *zap.SugaredLogger
	// mutex sync.Mutex
}

func New(fs ports.FileSystem, stor ports.Storager, slogger *zap.SugaredLogger) *service {
	return &service{
		fs:      fs,
		storage: stor,
		slogger: slogger,
		// mutex: sync.Mutex{},
	}
}

func (srv *service) FindFiles(path string) error {
	// obtain all the subfolders in the given path
	dirSlice, err := srv.fs.FindSubfolders(path)
	if err != nil {
		srv.slogger.Errorf("error finding subfolders, %v", err)
		return err
	}
	// obtain all the files in all subfolders
	fileSlice, err := srv.fs.FindFiles(dirSlice)
	if err != nil {
		srv.slogger.Errorf("error finding files, %v", err)
		return err
	}
	// store files data in the storage
	srv.storage.StoreFiles(fileSlice)
	return nil
}

// LeaveClones filters out all unique files from slice
func (srv *service) GetClones() error {
	srv.storage.GetClones()
	srv.storage.GetDeletable()
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
	err := srv.storage.GetClones()
	if err != nil {
		srv.slogger.Errorf("error geting clones, %v", err)
		return err
	}
	srv.storage.GetDeletable()
	clones := srv.storage.ReadFiles()

	for _, file := range clones {
		// remove file from file system
		if err := srv.fs.RemoveFile(file); err != nil {
			return err
		}
	}
	srv.storage.StoreFiles(clones[:0])
	srv.slogger.Debug("clones has been deleted")
	return nil
}
