package v2ray

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jasonlvhit/gocron"
)

var configFilePath string

func Start(config Config) *exec.Cmd {
	stream, _ := config.ToJSON()
	service := exec.Command("./xray")
	service.Stdin = strings.NewReader(stream)
	var out bytes.Buffer
	service.Stdout = &out
	err := service.Start()
	if err != nil {
		log.Fatal(err)
	}
	return service
}

type UpdateClient struct {
	Id        string
	ExpiredAt string
}

type AddClient struct {
	Email     string
	ExpiredAt string
}

type Client struct {
	Id        string
	AlterId   int8
	Email     string
	ExpiredAt string
	Deleted   bool
}

type Settings struct {
	Clients []Client
}

type Inbound struct {
	Port     int16
	Listen   string
	Protocol string
	Settings Settings
}

type Config struct {
	Log       map[string]interface{} `json:"config"`
	Stats     map[string]interface{}
	Api       map[string]interface{}
	Policy    map[string]interface{}
	Allocate  map[string]interface{}
	Inbounds  []Inbound
	Outbounds []map[string]interface{}
	Routing   map[string]interface{}
}

type V2ray struct {
	Config  Config
	Service *exec.Cmd
}

func (v2ray *V2ray) StartCheckExpiredClientTask() {
	gocron.Every(1).Day().At("00:00").Do(task, &v2ray)
}

func task(v2ray *V2ray) {
	isRestartV2rayService := false
	for _, client := range v2ray.Config.Inbounds[0].Settings.Clients {
		expiredTime, _ := time.Parse(time.RFC3339, client.ExpiredAt)
		if time.Now().Unix() > expiredTime.Unix() {
			client.Deleted = true
			client.AlterId = -128
			isRestartV2rayService = true
		}
	}
	if isRestartV2rayService {
		v2ray.Service.Process.Kill()
		Start(v2ray.Config)
	}
}

// config

func (config *Config) AddClient(addClient AddClient) Client {
	client := Client{Email: addClient.Email, Id: uuid.NewString(), AlterId: 0, Deleted: false, ExpiredAt: addClient.ExpiredAt}
	config.Inbounds[0].Settings.Clients = append(config.Inbounds[0].Settings.Clients, client)
	config.Save()
	return client
}

func (config *Config) DisableClient(id string) Client {
	var c Client
	for _, client := range config.Inbounds[0].Settings.Clients {
		if client.Id == id {
			client.Deleted = true
			client.AlterId = -128
			c = client
		}
	}
	config.Save()
	return c
}

func (config *Config) EnableClient(updateClient UpdateClient) Client {
	var c Client
	for _, client := range config.Inbounds[0].Settings.Clients {
		if client.Id == updateClient.Id {
			client.ExpiredAt = updateClient.ExpiredAt
			client.Deleted = false
			client.AlterId = 0
		}
	}
	config.Save()
	return c
}

func (config *Config) QueryAllClients() []Client {
	var clients []Client
	if len(config.Inbounds) > 0 {
		clients = append(clients, config.Inbounds[0].Settings.Clients...)
		return clients
	}
	return clients
}

func (config *Config) QueryRemovedClients() []Client {
	var clients []Client
	for _, client := range config.Inbounds[0].Settings.Clients {
		if client.Deleted {
			clients = append(clients, client)
		}
	}
	return clients
}

func (config *Config) ToByteArray() ([]byte, error) {
	byteArray, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return byteArray, nil
}

func (config *Config) ToJSON() (string, error) {
	byteArray, err := json.Marshal(config)
	if err != nil {
		return "", err
	}
	return string(byteArray[:]), nil
}

func (config *Config) Save() error {
	byteArray, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, byteArray, 0644)
}

func Create(path string) (Config, error) {
	configFilePath = path
	config := Config{}
	file, err := os.Open(path)
	if err != nil {
		return config, err
	}
	byteArray, err := ioutil.ReadAll(file)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(byteArray, &config)
	if err != nil {
		return config, err
	}
	return config, err
}
