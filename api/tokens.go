package api

import (
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jmhobbs/authy-cli/model"
	"github.com/xlzd/gotp"
)

type AuthenticatorTokensResult struct {
	Message             string        `json:"message"`
	Success             bool          `json:"success"`
	AuthenticatorTokens []model.Token `json:"authenticator_tokens"`
	// todo: Deleted
}

type AppSyncResult struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Apps    []model.App `json:"apps"`
	// todo: Deleted
}

func (a *Authy) AuthenticatorTokens(authyId, deviceId int, secretSeed string) ([]model.Token, error) {
	u, err := url.Parse(a.url("/json/users/%d/authenticator_tokens", authyId))
	if err != nil {
		return nil, err
	}

	dec, err := hex.DecodeString(secretSeed)
	if err != nil {
		return nil, err
	}

	totp := gotp.NewTOTP(strings.TrimRight(base32.StdEncoding.EncodeToString(dec), "="), 7, 10, nil)

	// todo: standardize on this, maybe as part of Authy.url
	q := url.Values{
		"locale":    []string{"en-US"},
		"api_key":   []string{"37b312a3d682b823c439522e1fd31c82"},
		"device_id": []string{strconv.Itoa(deviceId)},
		"apps":      []string{""}, // todo:
		"otp1":      []string{totp.Now()},
		"otp2":      []string{totp.AtTime(time.Now().Add(10 * time.Second))},
		"otp3":      []string{totp.AtTime(time.Now().Add(20 * time.Second))},
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(
		http.MethodGet,
		u.String(),
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(a.addHeaders(req))
	if err != nil {
		return nil, err
	}

	// todo: handle error HTTP codes

	var r AuthenticatorTokensResult

	err = json.NewDecoder(resp.Body).Decode(&r)

	// todo: handle success: false?

	return r.AuthenticatorTokens, err
}

func (a *Authy) AppSync(authyId, deviceId int, secretSeed string, apps []model.App) ([]model.App, error) {
	dec, err := hex.DecodeString(secretSeed)
	if err != nil {
		return nil, err
	}

	totp := gotp.NewTOTP(strings.TrimRight(base32.StdEncoding.EncodeToString(dec), "="), 7, 10, nil)

	form := url.Values{
		"locale":                   []string{"en-US"},
		"api_key":                  []string{"37b312a3d682b823c439522e1fd31c82"},
		"last_unlock_method_used":  []string{"none"},
		"last_unlock_date":         []string{"0"},
		"enabled_unlock_methods[]": []string{"none"},
		"otp1":                     []string{totp.Now()},
		"otp2":                     []string{totp.AtTime(time.Now().Add(10 * time.Second))},
		"otp3":                     []string{totp.AtTime(time.Now().Add(20 * time.Second))},
		"device_id":                []string{strconv.Itoa(deviceId)},
	}

	for _, app := range apps {
		form.Add(fmt.Sprintf("vs%s", app.Id), strconv.Itoa(app.Version))
	}

	req, err := http.NewRequest(http.MethodPost, a.url("/json/users/%d/devices/%d/apps/sync", authyId, deviceId), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(a.addHeaders(req))
	if err != nil {
		return nil, err
	}

	var r AppSyncResult

	return r.Apps, json.NewDecoder(resp.Body).Decode(&r)
}
