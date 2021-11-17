package main

import (
	"fmt"
	"os"

	"github.com/seggga/practice/internal/filesystem"
	"github.com/seggga/practice/internal/repositories/memrepo"
	"github.com/seggga/practice/internal/services/removersrv"
)

func main() {
	// define filesystem
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		return
	}
	fs := filesystem.New(os.DirFS(dir))
	// define storage
	stor := memrepo.New()
	service := removersrv.New(fs, stor)

	// obtain files
	err = service.FindFiles(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	// find clones
	err = service.GetClones() // TODO проверить, действительно ли нужно возвращать здесь ошибку
	if err != nil {
		fmt.Println(err)
		return
	}

	// clones removal
	err = service.RemoveClones()
	if err != nil {
		fmt.Println(err)
		return
	}
}
