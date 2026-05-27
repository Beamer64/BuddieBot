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

	if reflect.DeepEqual(cfg.configuration, &configuration{}) {
		t.Fatal("cfg.configuration is empty")
	}
}
