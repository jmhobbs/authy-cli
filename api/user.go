package api

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type UserStatusResponse struct {
	ForceOTT             bool   `json:"force_ott"`
	PrimaryEmailVerified bool   `json:"primary_email_verified"`
	Message              string `json:"message"`
	DevicesCount         int    `json:"devices_count"`
	AuthyId              int    `json:"authy_id"`
	Success              bool   `json:"success"`
}

func (a *Authy) UserStatus(countryCode, phone string) (UserStatusResponse, error) {
	u := UserStatusResponse{}

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/json/users/%s-%s/status?locale=en-US", a.base, countryCode, phone), nil)
	if err != nil {
		return u, err
	}

	resp, err := a.client.Do(a.addHeaders(req))
	if err != nil {
		return u, err
	}

	// todo: handle error HTTP codes

	err = json.NewDecoder(resp.Body).Decode(&u)

	return u, err
}
