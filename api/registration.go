package api

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

type RegistrationResult struct {
	Message     string `json:"message"`
	RequestId   string `json:"request_id"`
	ApprovalPin int    `json:"approval_pin"`
	Provider    string `json:"provider"`
	Success     bool   `json:"success"`
}

type RegistrationStatus struct {
	Message struct {
		RequestStatus string `json:"request_status"`
	} `json:"message"`
	Status  string `json:"status"`
	Pin     string `json:"pin"`
	Success bool   `json:"success"`
}

type DeviceRegistration struct {
	Device struct {
		Id         int    `json:"id"`
		SecretSeed string `json:"secret_seed"`
		ApiKey     string `json:"api_key"`
		Reinstall  bool   `json:"reinstall"`
	} `json:"device"`
	AuthyId int `json:"authy_id"`
}

func (a *Authy) RegistrationStart(authyId int, signature string) (RegistrationResult, error) {
	var r RegistrationResult

	form := url.Values{}
	form.Add("signature", signature)
	form.Add("via", "push")
	form.Add("device_app", "authy")
	form.Add("api_key", "37b312a3d682b823c439522e1fd31c82")

	req, err := http.NewRequest(http.MethodPost, a.url("/json/users/%d/devices/registration/start?locale=en-US", authyId), strings.NewReader(form.Encode()))
	if err != nil {
		return r, err
	}

	resp, err := a.client.Do(a.addHeaders(req))
	if err != nil {
		return r, err
	}

	// todo: handle error HTTP codes

	err = json.NewDecoder(resp.Body).Decode(&r)

	return r, err
}

func (a *Authy) RegistrationStatus(authyId int, signature, registrationId string) (RegistrationStatus, error) {
	var r RegistrationStatus

	req, err := http.NewRequest(
		http.MethodGet,
		a.url("/json/users/%d/devices/registration/%s/status?locale=en-US&signature=%s&api_key=37b312a3d682b823c439522e1fd31c82", authyId, registrationId, signature),
		nil,
	)
	if err != nil {
		return r, err
	}

	resp, err := a.client.Do(a.addHeaders(req))
	if err != nil {
		return r, err
	}

	// todo: handle error HTTP codes

	err = json.NewDecoder(resp.Body).Decode(&r)

	return r, err
}

func (a *Authy) RegistrationComplete(authyId int, deviceUUID, pin string) (DeviceRegistration, error) {
	var d DeviceRegistration

	form := url.Values{}
	form.Add("pin", pin)
	form.Add("uuid", deviceUUID)
	form.Add("device_app", "authy")
	form.Add("device_name", "Authy Desktop on localhost")
	form.Add("api_key", "37b312a3d682b823c439522e1fd31c82")

	req, err := http.NewRequest(http.MethodPost, a.url("/json/users/%d/devices/registration/complete", authyId), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if err != nil {
		return d, err
	}

	resp, err := a.client.Do(a.addHeaders(req))
	if err != nil {
		return d, err
	}

	// todo: handle error HTTP codes

	err = json.NewDecoder(resp.Body).Decode(&d)

	return d, err
}
