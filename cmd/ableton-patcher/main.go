package main

import (
	"crypto/dsa"
	"unspok3n/ableton-patcher/config"
	"unspok3n/ableton-patcher/internal/ableton"
)

const (
	configFilename = "ableton-patcher-config.yml"
)

type application struct {
	config     *config.PatcherConfig
	configPath string
	key        *dsa.PrivateKey
}

func main() {
	configFilePath, _ := FindFile(configFilename)
	c, err := config.Parse(configFilePath)
	if err != nil {
		LogFatalError("load config", err)
	}

	if configFilePath == "" {
		configFilePath, err = ExecutableDirFilePath(configFilename)
		if err != nil {
			LogFatalError("join config filename", err)
		}
	}

	key, err := ableton.HexToPrivateDSA(c.PrivateKey)
	if err != nil {
		LogFatalError("load private key", err)
	}

	app := &application{
		config:     c,
		configPath: configFilePath,
		key:        key,
	}

	app.mainMenu()
}
