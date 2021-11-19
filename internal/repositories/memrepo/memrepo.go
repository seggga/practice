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

func New(slogger *zap.SugaredLogger) *MemRepo {
	return &MemRepo{
		fileSlice: make([]domain.File, 0),
		slogger:   slogger,
	}
}

func (mr *MemRepo) StoreFiles(files []domain.File) {
	mr.fileSlice = files
}

func (mr *MemRepo) GetClones() error {

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
	return nil
}

func (mr *MemRepo) ReadFiles() []domain.File {
	return mr.fileSlice
}

func (mr *MemRepo) RemoveFile(file domain.File) {
	var ind int
	for i := 1; i < len(mr.fileSlice); i += 1 {
		if mr.fileSlice[i].Path == file.Path {
			ind = i
			break
		}
	}
	mr.fileSlice = append(mr.fileSlice[:ind], mr.fileSlice[ind+1:]...)
}
