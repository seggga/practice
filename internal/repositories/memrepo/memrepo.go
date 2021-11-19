package memrepo

import (
	"fmt"
	"sort"

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

// GetDeletable produces a slice with files to be deleted
func (mr *MemRepo) GetDeletable() {
	clones := mr.fileSlice
	clones = sortByID(clones)

	var cloneID = 0
	var cloneGroup, deletableFiles []domain.File
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
				deletableFiles = append(deletableFiles, cloneGroup[1:]...)
				// clear cloneGroup
				cloneGroup = cloneGroup[:0]
			}
			// create a new slice of clones
			cloneID = fileData.CloneID
		}
		cloneGroup = append(cloneGroup, fileData)
	}
	mr.fileSlice = deletableFiles
}

// ReadFiles reads file-data from MemRepo.Slices
func (mr *MemRepo) ReadFiles() []domain.File {
	mr.slogger.Debugf("obtained %d file-elements", len(mr.fileSlice))
	return mr.fileSlice
}
