package main

import (
	"encoding/json"
)

type Credential struct {
	Id          string `json:"id"`
	AccessToken string `json:"access_token"`
	UserId      string `json:"user_id"`
	UserName    string `json:"user_name"`
	CreateDate  string `json:"create_date"`
	ExpiresIn   int    `json:"expires_in"`
	TenantId    string `json:"tenant_id"`
}

func (cred *Credential) ToJsonStr() string {
	bytes, _ := json.Marshal(cred)
	return string(bytes)
}
