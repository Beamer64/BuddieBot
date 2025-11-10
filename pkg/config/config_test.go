package config

import (
	"os"
	"reflect"
	"testing"
)

func TestReadConfig(t *testing.T) {
	if os.Getenv("INTEGRATION") != "true" {
		t.Skip("skipping due to INTEGRATION env var not being set to 'true'")
	}

	cfg, err := ReadConfig()
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(cfg.Configs, &configuration{}) {
		t.Fatal("cfg.Configuration is empty")
	}

	if reflect.DeepEqual(cfg.Cmd, &command{}) {
		t.Fatal("cfg.CommandMessages is empty")
	}

	if len(cfg.LoadingMessages) == 0 {
		t.Fatal("cfg.LoadingMessages is empty")
	}

}
