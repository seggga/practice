package memrepo

import (
	"fmt"

	"github.com/seggga/practice/internal/domain"
	"go.uber.org/zap"
)

type MemRepo struct {
	fileSlice []domain.File
	slogger   *zap.SugaredLogger
}

// New creates a MemRepo instance
func New(slogger *zap.SugaredLogger) *MemRepo {
	return &MemRepo{
		fileSlice: make([]domain.File, 0),
		slogger:   slogger,
	}
}

// StoreFiles writes file-data to MemRepo.Slices
func (mr *MemRepo) StoreFiles(files []domain.File) {
	mr.fileSlice = files
	mr.slogger.Debugf("stored %d file-elements", len(files))
}

// GetClones changes storage content: it deletes file-data with unique files and leaves data about clones only
func (mr *MemRepo) GetClones() error {
	mr.slogger.Debug("started")
	cloneID := 1
	idMap := make(map[string]int)
	clonesCounter := make(map[int]int)

	// marking unique and clone files
	for i := 0; i < len(mr.fileSlice); i += 1 {
		cloneHash := mr.fileSlice[i].Name + fmt.Sprint(mr.fileSlice[i].SizeInBytes)
		id, ok := idMap[cloneHash]
		if !ok {
			// this file is unique, create a new ID
			idMap[cloneHash] = cloneID
			mr.fileSlice[i].CloneID = cloneID
			clonesCounter[cloneID] += 1
			cloneID += 1
			continue
		}
		// such a file has already been marked
		mr.fileSlice[i].CloneID = id
		clonesCounter[id] += 1
	}
	// calculate a capacity of slice to store only clone's data
	var capacity int
	for _, num := range clonesCounter {
		if num > 1 {
			capacity += num
		}
	}
	// construct a slice with only clones
	onlyClones := make([]domain.File, capacity)
	i := 0
	for _, someFile := range mr.fileSlice {
		if clonesCounter[someFile.CloneID] > 1 {
			onlyClones[i] = someFile
			i += 1
		}
	}
	// write slice back to the storage
	mr.fileSlice = onlyClones
	mr.slogger.Debugf("fount %d clones", len(onlyClones))
	return nil
}

// ReadFiles reads file-data from MemRepo.Slices
func (mr *MemRepo) ReadFiles() []domain.File {
	mr.slogger.Debugf("obtained %d file-elements", len(mr.fileSlice))
	return mr.fileSlice
}

// RemoveFile removes data about a file from the storage
func (mr *MemRepo) RemoveFile(file domain.File) {
	var ind int
	for i := 1; i < len(mr.fileSlice); i += 1 {
		if mr.fileSlice[i].Path == file.Path {
			ind = i
			break
		}
	}
	mr.slogger.Debugf("removing [ %s ] file from storage", file.Path)
	mr.fileSlice = append(mr.fileSlice[:ind], mr.fileSlice[ind+1:]...)
}
