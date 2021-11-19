package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/seggga/practice/internal/filesystem"
	"github.com/seggga/practice/internal/repositories/memrepo"
	"github.com/seggga/practice/internal/services/cloremover"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/yaml.v2"
)

func main() {
	// read config
	config, err := ReadConfig()
	if err != nil {
		fmt.Println("error reading config, program exit")
		return
	}

	// init logger
	loggerConfig, err := newLoggerConfig(*config)
	if err != nil {
		fmt.Printf("error reading config, %v", err)
	}
	logger, _ := loggerConfig.Build()
	// logger := zap.New(zapcore.NewCore(zap.Encode))
	logger.Info("hello, log!")
	// define filesystem
	if config.Dir == "" {
		fmt.Println("directory (Dir) not set in config.yaml, program exit")
		return
	}
	fs := filesystem.New(os.DirFS(config.Dir))
	// define storage
	stor := memrepo.New()
	service := cloremover.New(fs, stor)

	// obtain files
	err = service.FindFiles(config.Dir)
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
	// remove clones
	err = service.RemoveClones()
	if err != nil {
		fmt.Println(err)
		return
	}
}

type Config struct {
	LogLevel string `yaml:"loglevel"`
	Cores    int    `yaml:"cores"`
	Dir      string `yaml:"dir"`
}

// ReadConfig implements filling config from yaml-file
func ReadConfig() (*Config, error) {
	// read config file
	configData, err := ioutil.ReadFile("./configs/config.yaml")
	if err != nil {
		return nil, err
	}
	// decode config
	cfg := new(Config)
	err = yaml.Unmarshal(configData, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func newLoggerConfig(config Config) (*zap.Config, error) {
	var level zap.AtomicLevel
	switch config.LogLevel {
	case "debug":
		level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "info":
		level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zapcore.ErrorLevel)
	default:
		return nil, fmt.Errorf("incorrect loglevel value")
	}
	return &zap.Config{
		Encoding:    "json",
		Level:       level,
		OutputPaths: []string{"stdout"},
	}, nil
}
