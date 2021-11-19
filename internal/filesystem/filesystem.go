package filesystem

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"sync"

	"github.com/seggga/practice/internal/domain"
	"go.uber.org/zap"
)

type FileSystem struct {
	dir        string
	fileSystem fs.FS
	slogger    *zap.SugaredLogger
}

// New creates a FileSystem instance
func New(dir string, slogger *zap.SugaredLogger) *FileSystem {
	fsys := os.DirFS(dir)
	return &FileSystem{
		dir:        dir,
		fileSystem: fsys,
		slogger:    slogger,
	}
}

// FindSubfolders implements search for subdirectories inside the given path
func (f *FileSystem) FindSubfolders(path string) ([]string, error) {
	f.slogger.Debugf("start scan folders in %s directory", path)
	var dirSlice []string
	err := fs.WalkDir(f.fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			dirSlice = append(dirSlice, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	f.slogger.Debugf("obtained dirslice on %s directory, %d subfolders", path, len(dirSlice))
	return dirSlice, nil
}

// FindFiles uses pool of workers. Each go-routine looks for files in a particular subfolder.
// The number of simultaneous go-routines corresponds to the number of cores (or set by config).
func (f *FileSystem) FindFiles(dirSlice []string) ([]domain.File, error) {
	f.slogger.Debug("started scan files")
	resultSlice := make([]domain.File, 0)
	var mutex sync.Mutex
	// jobsQueue controls maximum amount of goruotines scanning subfolders
	jobsQueue := make(chan struct{}, 2) // TODO: change 2 on value from config (number of simultaneously running go-routines)
	defer close(jobsQueue)

	// initialize waitgroup with the number of subfolders
	var wg sync.WaitGroup
	wg.Add(len(dirSlice))

	// start worker's manager - jogsQueue writer
	go func() {
		for _, someDir := range dirSlice {
			// control the number of simultaneously executing goroutines
			jobsQueue <- struct{}{}
			// goroutite to scan particular subfolder
			go func(someDir string) {
				defer func() {
					<-jobsQueue
					wg.Done()
				}()
				// obtain files from specified directory
				files, err := fs.ReadDir(f.fileSystem, someDir)
				for _, someFile := range files {
					if !someFile.IsDir() {
						file := new(domain.File)
						file.Dir = someDir
						file.Name = someFile.Name()
						fileInfo, err := someFile.Info()
						if err != nil {
							// log error
							return
						}
						file.Path = path.Join(f.dir, someDir, file.Name)
						file.SizeInBytes = int(fileInfo.Size())
						mutex.Lock()
						resultSlice = append(resultSlice, *file)
						mutex.Unlock()
					}
				}
				if err != nil {
					fmt.Println(err)
					return
				}
			}(someDir)
		}
	}()
	// wait for all the data to be captured
	wg.Wait()
	f.slogger.Debugf("obtained fileslice, %d files", len(resultSlice))
	return resultSlice, nil
}

// RemoveFile removes a file specified
func (f *FileSystem) RemoveFile(file domain.File) error {
	f.slogger.Debugf("trying to remove [ %s ] file", file.Path)
	err := os.Remove(file.Path)
	if err != nil {
		f.slogger.Errorf("failed remove file", err)
	}
	return err
}
