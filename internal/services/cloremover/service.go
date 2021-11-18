package cloremover

import (
	"fmt"
	"sort"

	"github.com/seggga/practice/internal/domain"
	"github.com/seggga/practice/internal/ports"
)

type service struct {
	fs      ports.FileSystem
	storage ports.Storager
	dir     string
	// mutex sync.Mutex
}

func New(fs ports.FileSystem, stor ports.Storager) *service {
	return &service{
		fs:      fs,
		storage: stor,
		// mutex: sync.Mutex{},
	}
}

func (srv *service) FindFiles(path string) error {
	// obtain all the subfolders in the given path
	dirSlice, err := srv.fs.FindSubfolders(path)
	if err != nil {
		return err
	}
	// obtain all the files in all subfolders
	fileSlice, err := srv.fs.FindFiles(dirSlice)
	if err != nil {
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

// PrintClones prints out  found data to the console
func (srv *service) PrintClones(viewFiles int, dirLimit int) {

	clones := srv.storage.ReadFiles()
	clones = sortByID(clones)

	// determine the amount of clones of each CloneID
	cloneMap := make(map[int]int)
	for _, fileData := range clones {
		cloneMap[fileData.CloneID] += 1
	}
	// first string
	fmt.Printf("Directory [ %s ] contains clone-files as folows\n", srv.dir)

	var id, showCounter, limitCounter int
	outputMap := make(map[int]int)
	for _, fileData := range clones {
		if id != fileData.CloneID {
			// ID mismatch - means start of another group of clones with the new id
			showCounter += 1
			if showCounter > viewFiles {
				// that's enough to print clone-files
				return
			}

			id = fileData.CloneID
			outputMap[showCounter] = id

			fmt.Println()
			fmt.Printf("[%2d]: %s - %d bytes, %3d clones:\n", showCounter, fileData.Name, fileData.SizeInBytes, cloneMap[id])
			limitCounter = 1
		}

		if dirLimit > 0 && limitCounter > dirLimit {
			continue // that's enough to print directories for theese clone-files
		}
		fmt.Println(fileData.Dir)
		limitCounter += 1
	}

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
