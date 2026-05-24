package config

import (
	"log"
	"os"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/BuddieBot/pkg/voice_chat"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Configs is the top-level config bundle. The yaml-loaded settings are
// embedded so callers can write `cfg.Keys.X` directly. Cmd holds the
// loaded cmd.yaml message strings; Player is wired up at bot init.
type Configs struct {
	*configuration
	Cmd    *command
	Player *voice_chat.Player
}

type configuration struct {
	Keys struct {
		ProdBotToken string `yaml:"prodBotToken"`
		TestBotToken string `yaml:"testBotToken"`
		NinjaAPIKey  string `yaml:"ninjaAPIKey"`
		DoggoAPIkey  string `yaml:"doggoAPIkey"`
	} `yaml:"keys"`

	ApiURLs struct {
		SteamAPI       string `yaml:"steamAPI"`
		AffirmationAPI string `yaml:"affirmationAPI"`
		AdviceAPI      string `yaml:"adviceAPI"`
		DoggoAPI       string `yaml:"doggoAPI"`
		NinjaKatzAPI   string `yaml:"ninjaKatzAPI"`
		FakePersonAPI  string `yaml:"fakePersonAPI"`
		XkcdAPI        string `yaml:"xkcdAPI"`
		LandsatAPI     string `yaml:"landsatAPI"`
	} `yaml:"apiURLs"`

	DiscordIDs struct {
		MasterGuildID       string `yaml:"masterGuildID"`
		TestGuildID         string `yaml:"testGuildID"`
		ErrorLogChannelID   string `yaml:"errorLogChannelID"`
		EventNotifChannelID string `yaml:"eventNotifChannelID"`
	} `yaml:"discordIDs"`

	Settings struct {
		BotPrefix    string `yaml:"botPrefix"`
		BotAdminRole string `yaml:"botAdminRole"`
	} `yaml:"settings"`

	Lavalink struct {
		ProdHost     string `yaml:"prodHost"`
		ProdPort     string `yaml:"prodPort"`
		ProdPassword string `yaml:"prodPassword"`
		TestHost     string `yaml:"testHost"`
		TestPort     string `yaml:"testPort"`
		TestPassword string `yaml:"testPassword"`

		// Resolved at load time based on isLaunchedByDebugger(). The rest of
		// the code reads these — never the Prod/Test fields directly.
		Host     string `yaml:"-"`
		Port     string `yaml:"-"`
		Password string `yaml:"-"`
	} `yaml:"lavalink"`
}

type command struct {
	Msg struct {
		Invalid     string `yaml:"invalid"`
		NotBotAdmin string `yaml:"notBotAdmin"`
	} `yaml:"message"`
}

func ReadConfig() (*Configs, error) {
	return marshalConfigs("config_files/", "../config_files/", "../../config_files/")
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

	if helper.IsLaunchedByDebugger() {
		cfg.Lavalink.Host = cfg.Lavalink.TestHost
		cfg.Lavalink.Port = cfg.Lavalink.TestPort
		cfg.Lavalink.Password = cfg.Lavalink.TestPassword
	} else {
		cfg.Lavalink.Host = cfg.Lavalink.ProdHost
		cfg.Lavalink.Port = cfg.Lavalink.ProdPort
		cfg.Lavalink.Password = cfg.Lavalink.ProdPassword
	}

	err = yaml.Unmarshal(commandFile, cmd)
	if err != nil {
		return nil, err
	}

	return &Configs{
		configuration: cfg,
		Cmd:           cmd,
	}, nil
}
