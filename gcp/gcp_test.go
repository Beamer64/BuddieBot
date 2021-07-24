package gcp

import (
	"testing"
)

func TestStopMachine(t *testing.T) {
	client, err := NewGCPClient("../config/auth.json", "pokernotifications-229105", "us-central1-a")
	if err != nil {
		t.Fatal(err)
	}
	err = client.StopMachine("instance-2-minecraft")
	if err != nil {
		t.Fatal(err)
	}
}

func TestStartMachine(t *testing.T) {
	client, err := NewGCPClient("../config/auth.json", "pokernotifications-229105", "us-central1-a")
	if err != nil {
		t.Fatal(err)
	}
	err = client.StartMachine("instance-2-minecraft")
	if err != nil {
		t.Fatal(err)
	}
}
