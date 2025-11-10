package config

import (
	"bufio"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/beamer64/godagpi/dagpi"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Configs struct {
	Configs         *configuration
	Cmd             *command
	LoadingMessages []string
	Emojis          []string
	Clients         *dagpiClients
}

type configuration struct {
	Keys struct {
		ProdBotToken     string `yaml:"prodBotToken"`
		TestBotToken     string `yaml:"testBotToken"`
		WebHookToken     string `yaml:"webHookToken"`
		BotPublicKey     string `yaml:"botPublicKey"`
		TenorAPIkey      string `yaml:"tenorAPIkey"`
		DagpiAPIkey      string `yaml:"dagpiAPIkey"`
		NinjaAPIKey      string `yaml:"ninjaAPIKey"`
		DoggoAPIkey      string `yaml:"doggoAPIkey"`
		ModerationPATKey string `yaml:"moderationPATKey"`
		ImgbbAPIkey      string `yaml:"imgbbAPIkey"`
	} `yaml:"keys"`

	ApiURLs struct {
		SteamAPI       string `yaml:"steamAPI"`
		AffirmationAPI string `yaml:"affirmationAPI"`
		KanyeAPI       string `yaml:"kanyeAPI"`
		AdviceAPI      string `yaml:"adviceAPI"`
		DoggoAPI       string `yaml:"doggoAPI"`
		NinjaKatzAPI   string `yaml:"ninjaKatzAPI"`
		AlbumPickerAPI string `yaml:"albumPickerAPI"`
		FakePersonAPI  string `yaml:"fakePersonAPI"`
		XkcdAPI        string `yaml:"xkcdAPI"`
		ImgbbAPI       string `yaml:"imgbbAPI"`
		LandsatAPI     string `yaml:"landsatAPI"`
	} `yaml:"apiURLs"`

	DiscordIDs struct {
		CurrentBotAppID         string `yaml:"currentBotAppID"`
		ProdBotAppID            string `yaml:"prodBotAppID"`
		TestBotAppID            string `yaml:"testBotAppID"`
		MasterGuildID           string `yaml:"masterGuildID"`
		TestGuildID             string `yaml:"testGuildID"`
		WebHookID               string `yaml:"webHookID"`
		ErrorLogChannelID       string `yaml:"errorLogChannelID"`
		EventNotifChannelID     string `yaml:"eventNotifChannelID"`
		FeatureTestingChannelID string `yaml:"featureTestingChannelID"`
	} `yaml:"discordIDs"`

	Settings struct {
		BotPrefix            string `yaml:"botPrefix"`
		BotAdminRole         string `yaml:"botAdminRole"`
		Email                string `yaml:"email"`
		EmailPassword        string `yaml:"emailPassword"`
		EnableNSFWModeration bool   `yaml:"enableNSFWModeration"`
		NsfwChannelName      string `yaml:"nsfwChannelName"`
	} `yaml:"settings"`

	Database struct {
		TableName string `yaml:"tableName"`
		Region    string `yaml:"region"`
		AccessKey string `yaml:"accessKey"`
		SecretKey string `yaml:"secretKey"`
	} `yaml:"database"`

	ReqFileDirs struct {
		Datasets string `yaml:"datasets"`
	} `yaml:"reqFileDirs"`
}

type command struct {
	SlashName struct {
		Tuuck   string `yaml:"tuuck"`
		Pick    string `yaml:"pick"`
		Animals string `yaml:"animals"`
		Daily   string `yaml:"daily"`
		ImgSet  string `yaml:"img-set"`
		Play    string `yaml:"play"`
		Get     string `yaml:"get"`
	} `yaml:"slash-name"`

	PrefixName struct {
	} `yaml:"prefix-name"`

	Desc struct {
		Tuuck    string `yaml:"tuuck"`
		CoinFlip string `yaml:"coin-flip"`
		LMGTFY   string `yaml:"lmgtfy"`
		Pick     string `yaml:"pick"`
		Animals  string `yaml:"animals"`
		Daily    string `yaml:"daily"`
		ImgSet   string `yaml:"img-set"`
		Play     string `yaml:"play"`
		Get      string `yaml:"get"`
	} `yaml:"description"`

	Example struct {
		Tuuck    string `yaml:"tuuck"`
		CoinFlip string `yaml:"coin-flip"`
		LMGTFY   string `yaml:"lmgtfy"`
		Pick     string `yaml:"pick"`
		Animals  string `yaml:"animals"`
		Daily    string `yaml:"daily"`
		ImgSet   string `yaml:"img-set"`
		Play     string `yaml:"play"`
		Get      string `yaml:"get"`
	} `yaml:"example"`

	Msg struct {
		Invalid         string `yaml:"invalid"`
		NotBotAdmin     string `yaml:"notBotAdmin"`
		TenorAPIError   string `yaml:"tenorAPIError"`
		YoutubeAPIError string `yaml:"youtubeAPIError"`
	} `yaml:"message"`
}

type dagpiClients struct {
	Dagpi *dagpi.Client
}

func ReadConfig() (*Configs, error) {
	configs, err := marshalConfigs("config_files/", "../config_files/", "../../config_files/")
	if err != nil {
		return nil, err
	}

	configs, err = marshalDatasets(configs, "datasets/", "../datasets/", "../../datasets/")
	if err != nil {
		return nil, err
	}

	return configs, nil
}

func marshalDatasets(configs *Configs, possibleDsPaths ...string) (*Configs, error) {
	var datasetDir string
	for _, cp := range possibleDsPaths {
		if !strings.HasSuffix(cp, "/") {
			return nil, errors.New(cp + " is not a valid path, needs to end with '/'")
		}

		// attempt to open dir
		files, err := os.ReadDir(cp)
		if err != nil {
			log.Printf("Couldn't find file in dir %s\n", cp)
			continue
		}

		// build a map of necessary Dataset files
		requiredDatasetFiles := map[string]bool{
			"loading_messages.txt": false,
			"emojis.txt":           false,
		}

		// loops through all files in dir, check if any of them are required
		for _, f := range files {
			if _, ok := requiredDatasetFiles[f.Name()]; ok {
				requiredDatasetFiles[f.Name()] = true
			}
		}

		allDatasetsFound := true
		for _, v := range requiredDatasetFiles {
			if !v {
				allDatasetsFound = false
				break
			}
		}

		if !allDatasetsFound {
			log.Printf("missing one or more required Dataset files in directory %s: \n %+v \n", cp, requiredDatasetFiles)
		} else {
			log.Printf("SUCCESS found Required Files dir %s\n", cp)
			datasetDir = cp
			break
		}
	}

	log.Println("Reading from loading messages file...")
	msgs, err := grabStringLists(datasetDir + "loading_messages.txt")
	if err != nil {
		return nil, err
	}

	log.Println("Reading from emojis file...")
	emojis, err := grabStringLists(datasetDir + "emojis.txt")
	if err != nil {
		return nil, err
	}

	configs.Emojis = emojis
	configs.LoadingMessages = msgs
	configs.Configs.ReqFileDirs.Datasets = datasetDir

	return configs, nil
}

func marshalConfigs(possibleConfigPaths ...string) (*Configs, error) {
	var configDir string
	for _, cp := range possibleConfigPaths {
		if !strings.HasSuffix(cp, "/") {
			return nil, errors.New(cp + " is not a valid path, needs to end with '/'")
		}

		// attempt to open dir
		files, err := os.ReadDir(cp)
		if err != nil {
			log.Printf("Couldn't find file in dir %s\n", cp)
			continue
		}

		// build a map of necessary Config files
		requiredConfigFiles := map[string]bool{
			"config.yaml": false,
			"cmd.yaml":    false,
		}

		// loops through all files in dir, check if any of them are required
		for _, f := range files {
			if _, ok := requiredConfigFiles[f.Name()]; ok {
				requiredConfigFiles[f.Name()] = true
			}
		}

		// check if all values are set to true, meaning that all files were found
		allConfigsFound := true
		for _, v := range requiredConfigFiles {
			if !v {
				allConfigsFound = false
				break
			}
		}

		if !allConfigsFound {
			log.Printf("missing one or more required Config files in directory %s: \n %+v \n", cp, requiredConfigFiles)
		} else {
			log.Printf("SUCCESS found Required Files dir %s\n", cp)
			configDir = cp
			break
		}
	}

	log.Println("Reading from config file...")
	configFile, err := os.ReadFile(configDir + "config.yaml")
	if err != nil {
		return nil, err
	}

	log.Println("Reading from cmd file...")
	commandFile, err := os.ReadFile(configDir + "cmd.yaml")
	if err != nil {
		return nil, err
	}

	cfg := &configuration{}
	cmd := &command{}

	err = yaml.Unmarshal(configFile, cfg)
	if err != nil {
		return nil, err
	}

	if isLaunchedByDebugger() {
		cfg.DiscordIDs.CurrentBotAppID = cfg.DiscordIDs.TestBotAppID
	} else {
		cfg.DiscordIDs.CurrentBotAppID = cfg.DiscordIDs.ProdBotAppID
	}

	err = yaml.Unmarshal(commandFile, cmd)
	if err != nil {
		return nil, err
	}

	clients := registerClients(cfg)

	return &Configs{
		Configs: cfg,
		Cmd:     cmd,
		Clients: clients,
	}, nil
}

func registerClients(cfg *configuration) *dagpiClients {
	return &dagpiClients{
		Dagpi: &dagpi.Client{Auth: cfg.Keys.DagpiAPIkey},
	}
}

// finds and returns []string from txt file
func grabStringLists(strListPath string) ([]string, error) {
	file, err := os.Open(strListPath)
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

// IsLaunchedByDebugger Determines if application is being run by the debugger.
func isLaunchedByDebugger() bool {
	// gops executable must be in the path. See https://github.com/google/gops
	gopsOut, err := exec.Command("gops", strconv.Itoa(os.Getppid())).Output()
	if err == nil && strings.Contains(string(gopsOut), "\\dlv.exe") {
		// our parent process is (probably) the Delve debugger
		return true
	}
	return false
}
