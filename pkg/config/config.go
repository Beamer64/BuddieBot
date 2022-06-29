package config

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"strings"
)

type Configs struct {
	Configs         *Configuration
	Cmd             *Command
	LoadingMessages []string
}

type Configuration struct {
	Keys struct {
		ProdBotToken   string `yaml:"prodBotToken"`
		TestBotToken   string `yaml:"testBotToken"`
		WebHookToken   string `yaml:"webHookToken"`
		BotPublicKey   string `yaml:"botPublicKey"`
		TenorAPIkey    string `yaml:"tenorAPIkey"`
		DagpiAPIkey    string `yaml:"dagpiAPIkey"`
		InsultAPI      string `yaml:"insultAPI"`
		SteamAPI       string `yaml:"steamAPI"`
		AffirmationAPI string `yaml:"affirmationAPI"`
		KanyeAPI       string `yaml:"kanyeAPI"`
		AdviceAPI      string `yaml:"adviceAPI"`
		DoggoAPI       string `yaml:"doggoAPI"`
		AlbumPickerAPI string `yaml:"albumPickerAPI"`
	} `yaml:"keys"`

	DiscordIDs struct {
		MasterGuildID       string `yaml:"masterGuildID"`
		TestGuildID         string `yaml:"testGuildID"`
		WebHookID           string `yaml:"webHookID"`
		ErrorLogChannelID   string `yaml:"errorLogChannelID"`
		EventNotifChannelID string `yaml:"eventNotifChannelID"`
	} `yaml:"discordIDs"`

	Settings struct {
		BotPrefix     string `yaml:"botPrefix"`
		BotAdminRole  string `yaml:"botAdminRole"`
		Email         string `yaml:"email"`
		EmailPassword string `yaml:"emailPassword"`
	} `yaml:"settings"`

	Database struct {
		TableName string `yaml:"tableName"`
		Region    string `yaml:"region"`
		AccessKey string `yaml:"accessKey"`
		SecretKey string `yaml:"secretKey"`
	}

	Server struct {
		SSHKeyBody                  string `yaml:"sshKeyBody"`
		MachineIP                   string `yaml:"machineIP"`
		Type                        string `yaml:"type"`
		Project_ID                  string `yaml:"project_id"`
		Private_key_ID              string `yaml:"private_key_id"`
		Private_key                 string `yaml:"private_key"`
		Client_email                string `yaml:"client_email"`
		Client_ID                   string `yaml:"client_id"`
		Auth_URI                    string `yaml:"auth_uri"`
		Token_URI                   string `yaml:"token_uri"`
		Auth_provider_x509_cert_URL string `yaml:"auth_provider_x509_cert_url"`
		Client_x509_cert_URL        string `yaml:"client_x509_cert_url"`
		Zone                        string `yaml:"zone"`
	} `yaml:"vm"`
}

type Command struct {
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
		StartServer  string `yaml:"start-server"`
		StopServer   string `yaml:"stop-server"`
		ServerStatus string `yaml:"server-status"`
	} `yaml:"prefix-name"`

	Desc struct {
		Tuuck        string `yaml:"tuuck"`
		StartServer  string `yaml:"start-server"`
		StopServer   string `yaml:"stop-server"`
		CoinFlip     string `yaml:"coin-flip"`
		ServerStatus string `yaml:"server-status"`
		LMGTFY       string `yaml:"lmgtfy"`
		Pick         string `yaml:"pick"`
		Animals      string `yaml:"animals"`
		Daily        string `yaml:"daily"`
		ImgSet       string `yaml:"img-set"`
		Play         string `yaml:"play"`
		Get          string `yaml:"get"`
	} `yaml:"description"`

	Example struct {
		Tuuck        string `yaml:"tuuck"`
		StartServer  string `yaml:"start-server"`
		StopServer   string `yaml:"stop-server"`
		CoinFlip     string `yaml:"coin-flip"`
		ServerStatus string `yaml:"server-status"`
		LMGTFY       string `yaml:"lmgtfy"`
		Pick         string `yaml:"pick"`
		Animals      string `yaml:"animals"`
		Daily        string `yaml:"daily"`
		ImgSet       string `yaml:"img-set"`
		Play         string `yaml:"play"`
		Get          string `yaml:"get"`
	} `yaml:"example"`

	Msg struct {
		Invalid          string `yaml:"invalid"`
		WindUp           string `yaml:"windUp"`
		WindDown         string `yaml:"windDown"`
		FinishOpperation string `yaml:"finishOpperation"`
		ServerUP         string `yaml:"serverUP"`
		ServerDOWN       string `yaml:"serverDOWN"`
		CheckStatusUp    string `yaml:"checkStatusUp"`
		CheckStatusDown  string `yaml:"checkStatusDown"`
		NotBotAdmin      string `yaml:"notBotAdmin"`
		MCServerError    string `yaml:"mcServerError"`
		TenorAPIError    string `yaml:"tenorAPIError"`
		YoutubeAPIError  string `yaml:"youtubeAPIError"`
		InsultAPIError   string `yaml:"insultAPIError"`
	} `yaml:"message"`
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

func ReadConfig(possibleConfigPaths ...string) (*Configs, error) {

	var configDir string
	for _, cp := range possibleConfigPaths {
		if !strings.HasSuffix(cp, "/") {
			return nil, errors.New(cp + " is not a valid path, needs to end with '/'")
		}

		// attempt to open dir
		files, err := ioutil.ReadDir(cp)
		if err != nil {
			fmt.Printf("Couldn't find file in dir %s\n", cp)
			continue
		}

		// build a map of necessary files
		fmap := make(map[string]bool)
		fmap["config.yaml"] = false
		fmap["cmd.yaml"] = false
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
	configFile, err := ioutil.ReadFile(configDir + "config.yaml")
	if err != nil {
		return nil, err
	}

	fmt.Println("Reading from cmd file...")
	commandFile, err := ioutil.ReadFile(configDir + "cmd.yaml")
	if err != nil {
		return nil, err
	}

	var cfg *Configuration
	var command *Command

	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(commandFile, &command)
	if err != nil {
		return nil, err
	}

	fmt.Println("Reading from loading messages file...")
	msgs, err := grabLoadingMessages(configDir + "loadingMessages.txt")
	if err != nil {
		return nil, err
	}

	return &Configs{
		Configs:         cfg,
		Cmd:             command,
		LoadingMessages: msgs,
	}, nil
}

// don't move this out (circular dependency)
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
