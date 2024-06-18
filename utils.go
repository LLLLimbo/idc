package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

var charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type Token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	IdToken      string `json:"id_token"`
}

func MakeTokenRequest(state string, code string) (string, *Token, error) {
	//make a http call to oauth 2 token endpoint
	form := fmt.Sprintf("state=%s&code=%s&grant_type=authorization_code&redirect_uri=%s/idc/redirect/callback", state, code, *host)
	payload := strings.NewReader(form)
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/protocol/openid-connect/token", *keycloak), payload)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(*clientId, *clientSecret)
	if err != nil {
		return "", nil, err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", nil, err
	}
	token := &Token{}

	body, _ := io.ReadAll(resp.Body)
	log.Printf("Keycloak Token endpoint response: %s", string(body))
	_ = json.Unmarshal(body, token)
	return string(body), token, nil
}

func CreateSession(token *Token) (error, *Credential) {
	url := fmt.Sprintf("%s/ac/session/create", *authCenter)
	data, _ := json.Marshal(token)
	resp, _ := http.Post(url, "application/json", bytes.NewBuffer(data))
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	log.Printf("Auth Center response: %s", string(body))

	type CreateSessionResponse struct {
		Credential *Credential `json:"credential"`
	}

	respStruct := &CreateSessionResponse{}
	_ = json.Unmarshal(body, respStruct)
	return nil, respStruct.Credential
}

func stringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	rand.Seed(time.Now().UnixNano())

	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}

func randomStr(length int) string {
	return stringWithCharset(length, charset)
}
