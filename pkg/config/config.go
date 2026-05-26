package config

import (
	"log"
	"os"
	"strings"

	"github.com/Beamer64/BuddieBot/pkg/database"
	"github.com/Beamer64/BuddieBot/pkg/helper"
	"github.com/Beamer64/BuddieBot/pkg/voice_chat"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// Configs is the top-level bundle. *configuration is embedded so callers
// can write `cfg.Keys.X` directly. Player and DB are wired up at bot init.
type Configs struct {
	*configuration
	Cmd    *command
	Player *voice_chat.Player
	DB     *database.DB
}

type configuration struct {
	Keys struct {
		ProdBotToken string `yaml:"prodBotToken"`
		TestBotToken string `yaml:"testBotToken"`
		NinjaAPIKey  string `yaml:"ninjaAPIKey"`
		DoggoAPIkey  string `yaml:"doggoAPIkey"`
	} `yaml:"keys"`

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

		// Resolved at load based on IsLaunchedByDebugger; the rest of the
		// code reads these, never the Prod/Test fields directly.
		Host     string `yaml:"-"`
		Port     string `yaml:"-"`
		Password string `yaml:"-"`
	} `yaml:"lavalink"`

	Database struct {
		ProdPath string `yaml:"prodPath"`
		DevPath  string `yaml:"devPath"`

		// Resolved at load based on IsLaunchedByDebugger.
		Path string `yaml:"-"`
	} `yaml:"database"`
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

		files, err := os.ReadDir(cp)
		if err != nil {
			log.Printf("Couldn't find file in dir %s\n", cp)
			continue
		}

		requiredConfigFiles := map[string]bool{
			"config.yaml": false,
			"cmd.yaml":    false,
		}
		for _, f := range files {
			if _, ok := requiredConfigFiles[f.Name()]; ok {
				requiredConfigFiles[f.Name()] = true
			}
		}

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
		cfg.Database.Path = cfg.Database.DevPath
	} else {
		cfg.Lavalink.Host = cfg.Lavalink.ProdHost
		cfg.Lavalink.Port = cfg.Lavalink.ProdPort
		cfg.Lavalink.Password = cfg.Lavalink.ProdPassword
		cfg.Database.Path = cfg.Database.ProdPath
	}
	if cfg.Database.Path == "" {
		return nil, errors.New("database path not configured (set database.prodPath / database.devPath)")
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
