package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"

	"github.com/jmhobbs/authy-cli/har"
)

type Authy struct {
	harTransport *harRecordingTransport
	client       *http.Client
	base         string
	requestId    string
}

// todo: make this a variadic options based config

func New(base, requestId string) *Authy {
	if base == "" {
		base = "https://api.authy.com"
	}
	if requestId == "" {
		requestId = uuid.Must(uuid.NewRandom()).String()
	}

	t := &harRecordingTransport{
		transport:   http.DefaultTransport,
		builder:     har.New("authy-cli", "0.0.1"),
		sawARequest: false,
	}

	return &Authy{
		harTransport: t,
		client: &http.Client{
			Transport: t,
		},
		base:      base,
		requestId: requestId,
	}
}

func (a *Authy) MadeRequests() bool {
	return a.harTransport.sawARequest
}

func (a *Authy) WriteHAR(out io.Writer) error {
	return json.NewEncoder(out).Encode(a.harTransport.builder.Har())
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
