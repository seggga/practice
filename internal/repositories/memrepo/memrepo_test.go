package memrepo

import (
	"testing"

	"github.com/seggga/practice/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestGetClones(t *testing.T) {

	fileSlice := []domain.File{
		{Dir: ".", Name: "clone1", Path: "./clone1", SizeInBytes: 11, CloneID: 0},
		{Dir: ".", Name: "clone2", Path: "./clone2", SizeInBytes: 12, CloneID: 0},
		{Dir: "test-folder1", Name: "unique1", Path: "test-folder1/unique1", SizeInBytes: 100, CloneID: 0},
		{Dir: "test-folder1", Name: "clone1", Path: "test-folder1/clone1", SizeInBytes: 11, CloneID: 0},
		{Dir: "test-folder1", Name: "clone2", Path: "test-folder1/clone2", SizeInBytes: 12, CloneID: 0},
		{Dir: "test-folder2", Name: "unique2", Path: "test-folder2/unique2", SizeInBytes: 200, CloneID: 0},
		{Dir: "test-folder2", Name: "clone1", Path: "test-folder2/clone1", SizeInBytes: 11, CloneID: 0},
		{Dir: "test-folder2", Name: "clone3", Path: "test-folder2/clone3", SizeInBytes: 13, CloneID: 0},
		{Dir: "test-folder3", Name: "unique3", Path: "test-folder3/unique3", SizeInBytes: 300, CloneID: 0},
		{Dir: "test-folder3", Name: "unique4", Path: "test-folder3/unique4", SizeInBytes: 400, CloneID: 0},
		{Dir: "test-folder3", Name: "clone1", Path: "test-folder3/clone1", SizeInBytes: 11, CloneID: 0},
		{Dir: "test-folder3", Name: "clone2", Path: "test-folder3/clone2", SizeInBytes: 12, CloneID: 0},
		{Dir: "test-folder3/test-folder4", Name: "clone2", Path: "test-folder3/test-folder4/clone2", SizeInBytes: 12, CloneID: 0},
		{Dir: "test-folder3/test-folder4", Name: "clone3", Path: "test-folder3/test-folder4/clone3", SizeInBytes: 13, CloneID: 0},
	}

	type simpleFile struct {
		Name        string
		SizeInBytes int
		Path        string
		Dir         string
	}

	expected := []simpleFile{
		{Dir: ".", Name: "clone1", Path: "./clone1", SizeInBytes: 11},
		{Dir: ".", Name: "clone2", Path: "./clone2", SizeInBytes: 12},
		{Dir: "test-folder1", Name: "clone1", Path: "test-folder1/clone1", SizeInBytes: 11},
		{Dir: "test-folder1", Name: "clone2", Path: "test-folder1/clone2", SizeInBytes: 12},
		{Dir: "test-folder2", Name: "clone1", Path: "test-folder2/clone1", SizeInBytes: 11},
		{Dir: "test-folder2", Name: "clone3", Path: "test-folder2/clone3", SizeInBytes: 13},
		{Dir: "test-folder3", Name: "clone1", Path: "test-folder3/clone1", SizeInBytes: 11},
		{Dir: "test-folder3", Name: "clone2", Path: "test-folder3/clone2", SizeInBytes: 12},
		{Dir: "test-folder3/test-folder4", Name: "clone2", Path: "test-folder3/test-folder4/clone2", SizeInBytes: 12},
		{Dir: "test-folder3/test-folder4", Name: "clone3", Path: "test-folder3/test-folder4/clone3", SizeInBytes: 13},
	}

	memRepo := New()
	memRepo.StoreFiles(fileSlice)
	memRepo.GetClones()

	var outSlice []simpleFile
	for _, f := range memRepo.fileSlice {
		fData := simpleFile{
			Name:        f.Name,
			Dir:         f.Dir,
			Path:        f.Path,
			SizeInBytes: f.SizeInBytes,
		}
		outSlice = append(outSlice, fData)
	}
	assert.ElementsMatch(t, outSlice, expected, "Incorrect GetClones result: slices are not equal")
}
