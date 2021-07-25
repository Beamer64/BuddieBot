package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"
)

type Config struct {
	Token       string `json:"Token"`
	BotPrefix   string `json:"BotPrefix"`
	SSHKeyBody  string `json:"SSHKeyBody"`
	MachineIP   string `json:"MachineIP"`
	TenorAPIkey string `json:"TenorAPIkey"`
}

type Auth struct {
	Type                        string `json:"type"`
	Project_id                  string `json:"project_id"`
	Private_key_id              string `json:"private_key_id"`
	Private_key                 string `json:"private_key"`
	Client_email                string `json:"client_email"`
	Client_id                   string `json:"client_id"`
	Auth_uri                    string `json:"auth_uri"`
	Token_uri                   string `json:"token_uri"`
	Auth_provider_x509_cert_url string `json:"auth_provider_x509_cert_url"`
	Client_x509_cert_url        string `json:"client_x509_cert_url"`
	Zone                        string `json:"zone"`
}

type Command struct {
	Tuuck            string `json:"Tuuck"`
	Start            string `json:"Start"`
	Stop             string `json:"Stop"`
	Invalid          string `json:"Invalid"`
	WindUp           string `json:"WindUp"`
	WindDown         string `json:"WindDown"`
	FinishOpperation string `json:"FinishOpperation"`
	Horoscope        string `json:"Horoscope"`
	Gif              string `json:"Gif"`
}

func ReadConfig(configFilePath string, authFilePath string, commandFilePath string) (*Config, *Auth, *Command, error) {
	fmt.Println("Reading from config file...")
	configFile, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	//fmt.Println(string(configFile))

	fmt.Println("No Errors")

	fmt.Println("Reading from auth file...")
	authFile, err := ioutil.ReadFile(authFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	//fmt.Println(string(authFile))

	fmt.Println("No Errors")

	fmt.Println("Reading from command file...")
	commandFile, err := ioutil.ReadFile(commandFilePath)
	if err != nil {
		return nil, nil, nil, err
	}
	//fmt.Println(string(authFile))

	fmt.Println("No Errors")

	var config *Config
	var auth *Auth
	var comm *Command

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, nil, nil, err
	}

	err = json.Unmarshal(authFile, &auth)
	if err != nil {
		return nil, nil, nil, err
	}

	err = json.Unmarshal(configFile, &config)
	if err != nil {
		return nil, nil, nil, err
	}

	err = json.Unmarshal(commandFile, &comm)
	if err != nil {
		return nil, nil, nil, err
	}

	return config, auth, comm, nil
}

func GrabLoadingMessage(messageFile string) string {
	file, err := os.Open(messageFile)
	if err != nil {
		log.Fatal(err)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	rand.Seed(time.Now().Unix())

	message := lines[rand.Intn(len(lines))]

	return message
}

func Recovered() {
	if r := recover(); r != nil {
		fmt.Println("Recovered in recover()", r)
	}
}
