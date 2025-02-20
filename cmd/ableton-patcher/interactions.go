package main

import (
	"crypto/dsa"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unspok3n/ableton-patcher/internal/ableton"
)

func (app *application) mainMenu() {
	for {
		fmt.Print("1. Patch\n2. Unpatch\n3. Deauthorize\n4. Generate license\n5. Generate DSA key pair")
		fmt.Print("\nSelect option: ")
		option := GetLine()
		switch option {
		case "1":
			app.patcher(false)
		case "2":
			app.patcher(true)
		case "3":
			app.deauthorize()
		case "4":
			app.licenseGenerator()
		case "5":
			app.keyGenerator()
		default:
			fmt.Print("Invalid option\n\n")
			continue
		}
	}
}

func (app *application) deauthorize() {
	selectedInstsData := installationDataSelector()

	for _, data := range selectedInstsData {
		unlockFile := filepath.Join(data.Path, "/Unlock/Unlock.json")
		os.Remove(unlockFile)
		fmt.Printf("\nDeauthorized (%s)", data.Name)
	}

	fmt.Print("\n\n")
}

func (app *application) licenseGenerator() {
	fmt.Print("Enter HWID: ")
	hwid := GetLine()

	fmt.Print("Enter major version: ")
	version := GetLine()
	versionInt, err := strconv.Atoi(version)
	if err != nil {
		LogFatalError("parse version", err)
	}

	var edition int
	fmt.Print("1. Suite\n2. Standard\n3. Intro\n4. Lite")
	for {
		fmt.Print("\nSelect edition: ")
		editionNumber := GetLine()
		switch editionNumber {
		case "1":
			edition = 2
		case "2":
			edition = 0
		case "3":
			edition = 3
		case "4":
			edition = 4
		default:
			fmt.Println("Invalid edition")
			continue
		}
		break
	}

	license, err := ableton.GenerateLicense(*app.key, hwid, edition, versionInt)
	if err != nil {
		LogFatalError("generate license", err)
	}

	if err := ableton.WriteAuthorizationFile(license, fmt.Sprintf("Authorize_%s.auz", hwid)); err != nil {
		LogFatalError("write authorization file", err)
	}

	fmt.Print("Done\n\n")
}

func (app *application) keyGenerator() {
	// There's gotta be a better way of doing this
	var priv dsa.PrivateKey
	for {
		var params dsa.Parameters
		if err := dsa.GenerateParameters(&params, rand.Reader, dsa.L1024N160); err != nil {
			LogFatalError("generate parameters", err)
		}

		priv.Parameters = params
		if err := dsa.GenerateKey(&priv, rand.Reader); err != nil {
			LogFatalError("generate key", err)
		}

		if priv.Y.BitLen() == 1024 && priv.G.BitLen() < 1024 {
			break
		}
	}

	keyString, err := ableton.PrivateDSAToHex(&priv)
	if err != nil {
		LogFatalError("key to hex", err)
	}

	fmt.Printf("\n%s\n\n", keyString)

	fmt.Print("Update config? (y/n): ")
	answer := GetLine()
	if answer == "y" {
		app.config.PrivateKey = keyString
		app.key = &priv
		err := app.config.Save(app.configPath)
		if err != nil {
			LogFatalError("save config", err)
		}
		fmt.Print("Config saved\n\n")
	}
}

func (app *application) patcher(reverse bool) {
	public, err := ableton.PublicDSAToHex(&app.key.PublicKey)
	if err != nil {
		LogFatalError("key to hex", err)
	}

	if len(public) != len(app.config.OriginalPublicKey) {
		LogFatalError("compare keys", fmt.Errorf("key length mismatch: %d vs %d", len(public), len(app.config.OriginalPublicKey)))
	}

	selectedInsts := installationSelector()

	for _, installation := range selectedInsts {
		fileBytes, err := os.ReadFile(installation.Path)
		if err != nil {
			LogFatalError("open file", err)
		}

		var newContents string
		if reverse {
			newContents = strings.Replace(string(fileBytes), public, app.config.OriginalPublicKey, -1)
		} else {
			newContents = strings.Replace(string(fileBytes), app.config.OriginalPublicKey, public, -1)
		}

		err = os.WriteFile(installation.Path, []byte(newContents), 0)
		if err != nil {
			LogFatalError("write file", err)
		}

		fmt.Printf("\nPatch/Unpatch applied (%s)", installation.Path)
	}

	fmt.Print("\n\n")
}

func installationSelector() []ableton.Installation {
	installations, err := ableton.FindInstallations()
	if err != nil {
		LogFatalError("find installations", err)
	}

	fmt.Println()
	installationsLen := len(installations)
	if installationsLen == 0 {
		fmt.Print("\nNo installations were found, please enter the path to the binary file: ")
		input := GetLine()
		_, err = os.Stat(input)
		if err != nil {
			LogFatalError("installation path not found", err)
		}
		installations = append(installations, ableton.Installation{Path: input, Name: "Custom Installation"})
	}

	for i, installation := range installations {
		fmt.Printf("%d. %s (%s)\n", i+1, installation.Name, installation.Path)
	}

	var selectedInsts []ableton.Installation
	for {
		fmt.Print("Enter the installation number(s): ")
		input := GetLine()
		requestedInsts := strings.Split(input, " ")
		for _, result := range requestedInsts {
			resultInt, err := strconv.Atoi(result)
			if err != nil {
				fmt.Printf("invalid installation number: %s\n", result)
				continue
			}

			if resultInt > installationsLen || resultInt == 0 {
				fmt.Printf("invalid installation number: %d\n\n", resultInt)
				continue
			}

			selectedInsts = append(selectedInsts, installations[resultInt-1])
		}
		break
	}

	return selectedInsts
}

func installationDataSelector() []ableton.InstallationData {
	data, err := ableton.FindInstallationData()
	if err != nil {
		LogFatalError("find installation data", err)
	}

	fmt.Println()
	dataLen := len(data)
	if dataLen == 0 {
		fmt.Print("No installation data was found, please enter the path to the data directory: ")
		input := GetLine()
		_, err = os.Stat(input)
		if err != nil {
			LogFatalError("installation data directory does not exists", err)
		}
		data = append(data, ableton.InstallationData{Path: input, Name: "Custom Installation Data"})
	}

	for i, installationData := range data {
		fmt.Printf("%d. %s (%s)\n", i+1, installationData.Name, installationData.Path)
	}

	var selectedInstsData []ableton.InstallationData
	for {
		fmt.Print("Enter the installation data number(s): ")
		input := GetLine()
		requestedInstsData := strings.Split(input, " ")
		for _, result := range requestedInstsData {
			resultInt, err := strconv.Atoi(result)
			if err != nil {
				fmt.Printf("invalid installation data number: %s", result)
				continue
			}

			if resultInt > dataLen || resultInt == 0 {
				fmt.Printf("invalid installation data number: %d", resultInt)
				continue
			}

			selectedInstsData = append(selectedInstsData, data[resultInt-1])
		}
		break
	}

	return selectedInstsData
}
