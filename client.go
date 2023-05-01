package main

import (
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/motemen/go-loghttp"
	"github.com/xlzd/gotp"
)

type Authy struct {
	client    http.Client
	base      string
	requestId string
}

// todo: make this a variadic options based config
func NewAuthy(out io.Writer, base, requestId string) *Authy {
	if base == "" {
		base = "https://api.authy.com"
	}
	if requestId == "" {
		requestId = uuid.Must(uuid.NewRandom()).String()
	}

	return &Authy{
		client: http.Client{
			Transport: &loghttp.Transport{
				// todo: these should be storing into a buffer or something and then
				// we have the option to dump to disk or screen or whatever
				// OUTPUT TO A HAR MAYBE?!
				// https://github.com/mrichman/hargo/blob/master/types.go#L11
				LogRequest: func(req *http.Request) {
					reqBytes, err := httputil.DumpRequest(req, true)
					if err != nil {
						log.Printf("unable to dump request %s %s : %v", req.Method, req.URL, err)
					} else {
						_, err = out.Write(reqBytes)
						if err != nil {
							log.Printf("unable to write request: %v", err)
							log.Println(string(reqBytes))
						}
						out.Write([]byte("\n\n"))
					}
				},
				LogResponse: func(resp *http.Response) {
					respBytes, err := httputil.DumpResponse(resp, true)
					if err != nil {
						log.Printf("unable to dump response %s %s : %v", resp.Request.Method, resp.Request.URL, err)
					} else {
						_, err = out.Write(respBytes)
						if err != nil {
							log.Printf("unable to write response: %v", err)
							log.Println(string(respBytes))
							out.Write([]byte("\n\n"))
						}
					}
				},
			},
		},
		base:      base,
		requestId: requestId,
	}
}

func (a *Authy) addHeaders(req *http.Request) *http.Request {
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) AuthyDesktop/2.2.3 Chrome/96.0.4664.110 Electron/16.0.8 Safari/537.36")
	req.Header.Set("x-authy-api-key", "37b312a3d682b823c439522e1fd31c82")
	req.Header.Set("x-authy-device-app", "authy")
	req.Header.Set("x-authy-private-ip", "127.0.0.1,::1")
	req.Header.Set("x-authy-request-id", a.requestId)
	req.Header.Set("x-user-agent", "AuthyDesktop 2.2.3")
	// Even the GET requests have this
	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	return req
}

func (a *Authy) RotateRequestId() {
	a.requestId = uuid.Must(uuid.NewRandom()).String()
}

func (a *Authy) url(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s", a.base, fmt.Sprintf(format, args...))
}

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

type RegistrationResult struct {
	Message     string `json:"message"`
	RequestId   string `json:"request_id"`
	ApprovalPin int    `json:"approval_pin"`
	Provider    string `json:"provider"`
	Success     bool   `json:"success"`
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

type RegistrationStatus struct {
	Message struct {
		RequestStatus string `json:"request_status"`
	} `json:"message"`
	Status  string `json:"status"`
	Pin     string `json:"pin"`
	Success bool   `json:"success"`
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

type DeviceRegistration struct {
	Device struct {
		Id         int    `json:"id"`
		SecretSeed string `json:"secret_seed"`
		ApiKey     string `json:"api_key"`
		Reinstall  bool   `json:"reinstall"`
	} `json:"device"`
	AuthyId int `json:"authy_id"`
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

type AuthenticatorTokensResult struct {
	Message             string               `json:"message"`
	Success             bool                 `json:"success"`
	AuthenticatorTokens []AuthenticatorToken `json:"authenticator_tokens"`
	// todo: Deleted
}

type AuthenticatorToken struct {
	AccountType             string      `json:"account_type"`
	Digits                  int         `json:"digits"`
	EncryptedSeed           string      `json:"encrypted_seed"`
	Issuer                  interface{} `json:"issuer"`
	KeyDerivationIterations interface{} `json:"key_derivation_iterations"`
	Logo                    interface{} `json:"logo"`
	Name                    string      `json:"name"`
	OriginalName            string      `json:"original_name"`
	PasswordTimestamp       int         `json:"password_timestamp"`
	Salt                    string      `json:"salt"`
	UniqueId                string      `json:"unique_id"`
}

func (a *Authy) AuthenticatorTokens(authyId, deviceId int, secretSeed string) ([]AuthenticatorToken, error) {
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

type AppSyncResult struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
	Apps    []App  `json:"apps"`
	// todo: deleted
}

type App struct {
	Id                string `json:"_id"`
	Name              string `json:"name"`
	SerialId          int    `json:"serial_id"`
	Version           int    `json:"version"`
	AssetsGroup       string `json:"assets_group"`
	BackgroundColor   any    `json:"background_color"`
	CircleBackground  any    `json:"circle_background"`
	CircleColor       any    `json:"circle_color"`
	CustomAssets      bool   `json:"custom_assets"`
	GeneratingAssets  bool   `json:"generating_assets"`
	LabelsColor       any    `json:"labels_color"`
	LabelsShadowColor any    `json:"labels_shadow_color"`
	TimerColor        string `json:"timer_color"`
	TokenColor        any    `json:"token_color"`
	AuthyId           int    `json:"authy_id"`
	SecretSeed        string `json:"secret_seed"`
	Digits            int    `json:"digits"`
	MemberSince       int    `json:"member_since"`
	TransactionalOtp  bool   `json:"transactional_otp"`
}

func (a *Authy) AppSync(authyId, deviceId int, secretSeed string) ([]App, error) {
	dec, err := hex.DecodeString(secretSeed)
	if err != nil {
		return nil, err
	}

	totp := gotp.NewTOTP(strings.TrimRight(base32.StdEncoding.EncodeToString(dec), "="), 7, 10, nil)

	// todo: use vsXXXX for existing apps

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
