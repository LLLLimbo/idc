package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type AppConfig struct {
	Redirect      string   `json:"redirect"`
	TokenEndpoint string   `json:"token_endpoint"`
	SessionType   []string `json:"session_type"`
}

type AppContext struct {
	Apps map[string]*AppConfig
}

func NewAppContext() *AppContext {
	return &AppContext{
		Apps: make(map[string]*AppConfig),
	}
}

func (ac *AppContext) ReadApps() {
	file, err := os.Open(*apps)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&ac.Apps)
	if err != nil {
		fmt.Println("Error decoding json:", err)
	}
}

func (ac *AppContext) GetAppConfig(appId string) *AppConfig {
	return ac.Apps[appId]
}

func (ac *AppContext) GetRedirect(appId string) string {
	redirect := ac.Apps[appId].Redirect
	return redirect
}
