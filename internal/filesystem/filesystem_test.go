package filesystem

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

/*
test-folder (.)
├── clone1
├── clone2
├── test-folder1
│   ├── clone1
│   ├── clone2
│	└── unique1
├── test-folder2
│   ├── clone1
│   ├── clone3
│	└── unique2
└── test-folder3
    ├── clone1
    ├── clone2
    ├── unique3
	├── unique4
 	└── test-folder4
	    ├── clone2
    	└── clone3
*/
var sysVal int

func createTestFolder() {
	os.MkdirAll("test-folder", os.ModePerm)
	os.Create(filepath.Join("test-folder", "clone1"))
	os.Create(filepath.Join("test-folder", "clone2"))
	os.MkdirAll(filepath.Join("test-folder", "test-folder1"), os.ModePerm)
	os.MkdirAll(filepath.Join("test-folder", "test-folder2"), os.ModePerm)
	os.MkdirAll(filepath.Join("test-folder", "test-folder3"), os.ModePerm)
	os.MkdirAll(filepath.Join("test-folder", "test-folder3", "test-folder4"), os.ModePerm)
	os.Create(filepath.Join("test-folder", "test-folder1", "clone1"))
	os.Create(filepath.Join("test-folder", "test-folder1", "clone2"))
	os.Create(filepath.Join("test-folder", "test-folder1", "unique1"))
	os.Create(filepath.Join("test-folder", "test-folder2", "clone1"))
	os.Create(filepath.Join("test-folder", "test-folder2", "clone3"))
	os.Create(filepath.Join("test-folder", "test-folder2", "unique2"))
	os.Create(filepath.Join("test-folder", "test-folder3", "clone1"))
	os.Create(filepath.Join("test-folder", "test-folder3", "clone2"))
	os.Create(filepath.Join("test-folder", "test-folder3", "unique3"))
	os.Create(filepath.Join("test-folder", "test-folder3", "unique4"))
	os.Create(filepath.Join("test-folder", "test-folder3", "test-folder4", "clone2"))
	os.Create(filepath.Join("test-folder", "test-folder3", "test-folder4", "clone3"))
}

func deleteTestFolder() {
	os.RemoveAll("test-folder")
}

// cropped fileData-type for testing only
type fileDataReduced struct {
	Dir  string
	Name string
}

func TestFindSubfoldersRealFS(t *testing.T) {
	createTestFolder()
	defer deleteTestFolder()

	dir := "test-folder"
	FS := New(dir)
	dirSlice, _ := FS.FindSubfolders(dir)

	expectedSlice := []string{".", "test-folder1", "test-folder2", "test-folder3", "test-folder3/test-folder4"}
	assert.ElementsMatch(t, dirSlice, expectedSlice, "slices not equal")
}

func TestFindFilesRealFS(t *testing.T) {

	expectedSlice := []fileDataReduced{
		{Dir: ".", Name: "clone1"},
		{Dir: ".", Name: "clone2"},
		{Dir: "test-folder1", Name: "unique1"},
		{Dir: "test-folder1", Name: "clone1"},
		{Dir: "test-folder1", Name: "clone2"},
		{Dir: "test-folder2", Name: "unique2"},
		{Dir: "test-folder2", Name: "clone1"},
		{Dir: "test-folder2", Name: "clone3"},
		{Dir: "test-folder3", Name: "unique3"},
		{Dir: "test-folder3", Name: "unique4"},
		{Dir: "test-folder3", Name: "clone1"},
		{Dir: "test-folder3", Name: "clone2"},
		{Dir: "test-folder3/test-folder4", Name: "clone2"},
		{Dir: "test-folder3/test-folder4", Name: "clone3"},
	}

	createTestFolder()
	defer deleteTestFolder()

	dir := "test-folder"
	FS := New(dir)
	dirSlice, _ := FS.FindSubfolders(dir)
	fileSlice, _ := FS.FindFiles(dirSlice)

	outFileSlice := make([]fileDataReduced, len(fileSlice))
	for i := range fileSlice {
		outFileSlice[i].Dir = fileSlice[i].Dir
		outFileSlice[i].Name = fileSlice[i].Name
	}

	assert.ElementsMatch(t, outFileSlice, expectedSlice, "File data slice is not valid")
}
