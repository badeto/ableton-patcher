package ableton

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	defaultWindowsInstallationLocation = "C:\\ProgramData\\Ableton"
)

type Installation struct {
	Path string
	Name string
}

type InstallationData struct {
	Path string
	Name string
}

func FindInstallations() ([]Installation, error) {
	var installations []Installation
	err := filepath.Walk(defaultWindowsInstallationLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.Contains(info.Name(), "Live") {
			binDir := filepath.Join(path, "/Program")
			instName := info.Name()

			err := filepath.Walk(binDir, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}

				if !info.IsDir() && strings.Contains(info.Name(), "Live") && strings.Contains(info.Name(), ".exe") {
					installations = append(installations, Installation{path, instName})
				}
				return nil
			})
			if err != nil {
				return err
			}

			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return installations, err
	}

	return installations, err
}

func FindInstallationData() ([]InstallationData, error) {
	var data []InstallationData

	appData, err := os.UserConfigDir()
	if err != nil {
		return data, err
	}
	defaultDataLocation := filepath.Join(appData, "/Ableton")
	err = filepath.Walk(defaultDataLocation, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.Contains(info.Name(), "Live") {
			dataPath := path
			dataName := info.Name()
			err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return nil
				}

				if info.IsDir() && strings.Contains(info.Name(), "Unlock") {
					data = append(data, InstallationData{dataPath, dataName})
				}
				return nil
			})
			if err != nil {
				return err
			}

			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return data, err
	}

	return data, err
}
