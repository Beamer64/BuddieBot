package config

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type Config struct {
	ExternalServicesConfig *ExternalServicesConfig
	GCPAuth                *GCPAuth
	CommandMessages        *CommandMessages
	CommandDescriptions    *CommandDescriptions
	LoadingMessages        []string
	Version                string
}

type ExternalServicesConfig struct {
	Token           string `json:"Token"`
	BotPrefix       string `json:"BotPrefix"`
	SSHKeyBody      string `json:"SSHKeyBody"`
	MachineIP       string `json:"MachineIP"`
	TenorAPIkey     string `json:"TenorAPIkey"`
	YoutubeAPIKey   string `json:"YoutubeAPIKey"`
	DiscordEmail    string `json:"DiscordEmail"`
	DiscordPassword string `json:"DiscordPassword"`
	BotAdminRole    string `json:"BotAdminRole"`
}

type GCPAuth struct {
	Type                        string `json:"type"`
	Project_ID                  string `json:"project_id"`
	Private_key_ID              string `json:"private_key_id"`
	Private_key                 string `json:"private_key"`
	Client_email                string `json:"client_email"`
	Client_ID                   string `json:"client_id"`
	Auth_URI                    string `json:"auth_uri"`
	Token_URI                   string `json:"token_uri"`
	Auth_provider_x509_cert_URL string `json:"auth_provider_x509_cert_url"`
	Client_x509_cert_URL        string `json:"client_x509_cert_url"`
	Zone                        string `json:"zone"`
}

type CommandMessages struct {
	Invalid          string `json:"Invalid"`
	WindUp           string `json:"WindUp"`
	WindDown         string `json:"WindDown"`
	FinishOpperation string `json:"FinishOpperation"`
	ServerUP         string `json:"ServerUP"`
	ServerDOWN       string `json:"ServerDOWN"`
	CheckStatusUp    string `json:"CheckStatusUp"`
	CheckStatusDown  string `json:"CheckStatusDown"`
	NotBotAdmin      string `json:"NotBotAdmin"`
}

type CommandDescriptions struct {
	Tuuck     string `json:"Tuuck"`
	Start     string `json:"Start"`
	Stop      string `json:"Stop"`
	CoinFlip  string `json:"CoinFlip"`
	Horoscope string `json:"Horoscope"`
	Gif       string `json:"Gif"`
	McStatus  string `json:"McStatus"`
	Version   string `json:"Version"`
	LMGTFY    string `json:"LMGTFY"`
}

type ServerCommandOut struct {
	Command      string `json:"CommandMessages"`
	CreatedAt    string `json:"CreatedAt"`
	ID           string `json:"ID"`
	Image        string `json:"Image"`
	Labels       string `json:"Labels"`
	LocalVolumes string `json:"LocalVolumes"`
	Mounts       string `json:"Mounts"`
	Names        string `json:"Names"`
	Networks     string `json:"Networks"`
	Ports        string `json:"Ports"`
	RunningFor   string `json:"RunningFor"`
	Size         string `json:"Size"`
	State        string `json:"State"`
	Status       string `json:"Status"`
}

func ReadConfig(possibleConfigPaths ...string) (*Config, error) {

	var configDir string
	for _, cp := range possibleConfigPaths {
		if !strings.HasSuffix(cp, "/") {
			return nil, errors.New(cp + " is not a valid path, needs to end with '/'")
		}

		// attempt to open dir
		files, err := ioutil.ReadDir(cp)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		// build a map of neccesary files
		fmap := make(map[string]bool)
		fmap["auth.json"] = false
		fmap["config.json"] = false
		fmap["cmdMessages.json"] = false
		fmap["cmdDescriptions.json"] = false
		fmap["loadingMessages.txt"] = false

		// loops thru all files in dir, check if any of them are required
		for _, f := range files {
			for reqFile := range fmap {
				if reqFile == f.Name() {
					fmap[reqFile] = true
				}
			}
		}

		// check if all values are set to true, meaning that all files were found
		allFound := true
		for _, v := range fmap {
			if !v {
				allFound = false
				break
			}
		}

		if !allFound {
			fmt.Printf("missing one or more required files in directory %s: \n%+v\n", cp, fmap)
		} else {
			configDir = cp
			fmt.Printf("SUCCESS found config dir %s\n", configDir)
			break
		}
	}

	fmt.Println("Reading from config file...")
	configFile, err := ioutil.ReadFile(configDir + "config.json")
	if err != nil {
		return nil, err
	}

	fmt.Println("Reading from auth file...")
	authFile, err := ioutil.ReadFile(configDir + "auth.json")
	if err != nil {
		return nil, err
	}

	fmt.Println("Reading from cmdMsg file...")
	commandMsgFile, err := ioutil.ReadFile(configDir + "cmdMessages.json")
	if err != nil {
		return nil, err
	}

	fmt.Println("Reading from cmdMsg file...")
	commandDescFile, err := ioutil.ReadFile(configDir + "cmdDescriptions.json")
	if err != nil {
		return nil, err
	}

	var escfg *ExternalServicesConfig
	var gcpauth *GCPAuth
	var comMsg *CommandMessages
	var comDesc *CommandDescriptions

	err = json.Unmarshal(configFile, &escfg)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(authFile, &gcpauth)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(commandMsgFile, &comMsg)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(commandDescFile, &comDesc)
	if err != nil {
		return nil, err
	}

	fmt.Println("Reading from loading messages file...")
	msgs, err := grabLoadingMessages(configDir + "loadingMessages.txt")
	if err != nil {
		return nil, err
	}

	fmt.Println("Looking for version.txt")
	file, err := os.Open(configDir + "version.txt")
	if err != nil {
		fmt.Println("WARNING didn't find version.txt if directory")
	}
	contents, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println("WARNING couldn't read version.txt")
	}

	fv := strings.Replace(string(contents), "\n", "", -1)
	if len(fv) > 8 {
		fv = fv[0:7]
	}

	return &Config{
		ExternalServicesConfig: escfg,
		GCPAuth:                gcpauth,
		CommandMessages:        comMsg,
		CommandDescriptions:    comDesc,
		LoadingMessages:        msgs,
		Version:                fv,
	}, nil
}

func grabLoadingMessages(loadingMessagesPath string) ([]string, error) {
	file, err := os.Open(loadingMessagesPath)
	if err != nil {
		return nil, err
	}

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	err = file.Close()
	if err != nil {
		return nil, err
	}

	return lines, nil
}
