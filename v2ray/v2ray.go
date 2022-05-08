package v2ray

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jasonlvhit/gocron"
)

var configFilePath string

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

type V2ray struct {
	Log       map[string]interface{} `json:"config"`
	Stats     map[string]interface{}
	Api       map[string]interface{}
	Policy    map[string]interface{}
	Allocate  map[string]interface{}
	Inbounds  []Inbound
	Outbounds []map[string]interface{}
	Routing   map[string]interface{}
}

func (v2ray *V2ray) AddClient(addClient AddClient) {
	client := Client{Email: addClient.Email, Id: uuid.NewString(), AlterId: 0, Deleted: false, ExpiredAt: addClient.ExpiredAt}
	v2ray.Inbounds[0].Settings.Clients = append(v2ray.Inbounds[0].Settings.Clients, client)
}

func (v2ray *V2ray) DisableClient(id string) {
	for _, client := range v2ray.Inbounds[0].Settings.Clients {
		if client.Id == id {
			client.Deleted = true
			client.AlterId = -128
		}
	}
}

func (v2ray *V2ray) EnableClient(updateClient UpdateClient) {
	for _, client := range v2ray.Inbounds[0].Settings.Clients {
		if client.Id == updateClient.Id {
			client.ExpiredAt = updateClient.ExpiredAt
			client.Deleted = false
			client.AlterId = 0
		}
	}
}

func (v2ray *V2ray) QueryAllClients() []Client {
	var clients []Client
	clients = append(clients, v2ray.Inbounds[0].Settings.Clients...)
	return clients

}

func (v2ray *V2ray) QueryRemovedClients() []Client {
	var clients []Client
	for _, client := range v2ray.Inbounds[0].Settings.Clients {
		if client.Deleted {
			clients = append(clients, client)
		}
	}
	return clients
}

func task(v2ray *V2ray) {
	isRestartV2rayService := false
	for _, client := range v2ray.Inbounds[0].Settings.Clients {
		expiredTime, _ := time.Parse(time.RFC3339, client.ExpiredAt)
		if time.Now().Unix() > expiredTime.Unix() {
			client.Deleted = true
			client.AlterId = -128
			isRestartV2rayService = true
		}
	}
	if isRestartV2rayService {
		//restart v2ray service

	}
}

func (v2ray *V2ray) runCheckExpiredTask() {
	gocron.Every(1).Day().At("00:00").Do(task, &v2ray)
}

func (v2ray *V2ray) ToByteArray() ([]byte, error) {
	byteArray, err := json.Marshal(v2ray)
	if err != nil {
		return nil, err
	}
	return byteArray, nil
}

func (v2ray *V2ray) ToJSON() (string, error) {
	byteArray, err := json.Marshal(v2ray)
	if err != nil {
		return "", err
	}
	return string(byteArray[:]), nil
}

func (v2ray *V2ray) Save() error {
	byteArray, err := json.Marshal(v2ray)
	if err != nil {
		return err
	}
	return os.WriteFile(configFilePath, byteArray, 0644)
}

func Create(path string) (V2ray, error) {
	configFilePath = path
	v2ray := V2ray{}
	file, err := os.Open(path)
	if err != nil {
		return v2ray, err
	}
	byteArray, err := ioutil.ReadAll(file)
	if err != nil {
		return v2ray, err
	}

	err = json.Unmarshal(byteArray, &v2ray)
	if err != nil {
		return v2ray, err
	}
	return v2ray, err
}
