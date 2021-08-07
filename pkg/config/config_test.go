package config

import (
	"reflect"
	"testing"
)

func TestReadConfig(t *testing.T) {
	cfg, err := ReadConfig("config/", "../config/", "../../config/")
	if err != nil {
		t.Fatal(err)
	}

	if reflect.DeepEqual(cfg.ExternalServicesConfig, &ExternalServicesConfig{}) {
		t.Fatal("cfg.ExternalServicesConfig is empty")
	}

	if reflect.DeepEqual(cfg.GCPAuth, &GCPAuth{}) {
		t.Fatal("cfg.GCPAuth is empty")
	}

	if reflect.DeepEqual(cfg.CommandMessages, &CommandMessages{}) {
		t.Fatal("cfg.CommandMessages is empty")
	}

	if len(cfg.LoadingMessages) == 0 {
		t.Fatal("cfg.LoadingMessages is empty")
	}

}
