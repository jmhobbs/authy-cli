package api

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/jmhobbs/authy-cli/har"
)

type harRecordingTransport struct {
	transport   http.RoundTripper
	builder     *har.Builder
	sawARequest bool
}

func (t *harRecordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.sawARequest = true

	h := md5.New()
	h.Write([]byte(req.URL.String()))
	id := fmt.Sprintf("%d_%x", time.Now().UnixMicro(), h.Sum(nil))

	t.builder.Page(id, req.URL.String())

	entry := t.builder.NewRequest(id, id, req)
	resp, err := t.transport.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	entry.SetResponse(resp)
	return resp, err
}
