package har

import (
	"net/http"
	"time"
)

const iso8601 string = "2006-01-02T15:04:05+07:00"

type Builder struct {
	Creator Creator
	Pages   map[string]*Page
	Entries map[string]*Entry
}

func New(creator, version string) *Builder {
	return &Builder{
		Creator: Creator{
			Name:    creator,
			Version: version,
		},
		Pages:   make(map[string]*Page),
		Entries: make(map[string]*Entry),
	}
}

func (b *Builder) Har() Har {
	pages := []Page{}
	entries := []Entry{}

	for _, page := range b.Pages {
		pages = append(pages, *page)
	}

	for _, entry := range b.Entries {
		entries = append(entries, *entry)
	}

	return Har{
		Log: Log{
			Version: "1.2",
			Creator: b.Creator,
			Pages:   pages,
			Entries: entries,
		},
	}
}

func (b *Builder) Page(id, title string) *Page {
	p := Page{
		StartedDateTime: time.Now().UTC().Format(iso8601),
		Title:           title,
		ID:              id,
		PageTiming: PageTiming{
			OnContentLoad: -1,
			OnLoad:        -1,
		},
	}

	b.Pages[p.ID] = &p

	return &p
}

func mapHeaders(headers http.Header) []NVP {
	out := []NVP{}

	for header, values := range headers {
		for _, value := range values {
			out = append(out, NVP{
				Name:  header,
				Value: value,
			})
		}
	}

	return out
}

func mapCookies(cookies []*http.Cookie) []Cookie {
	out := []Cookie{}

	for _, cookie := range cookies {
		out = append(out, Cookie{
			Name:     cookie.Name,
			Value:    cookie.Value,
			Path:     cookie.Path,
			Domain:   cookie.Domain,
			Expires:  cookie.Expires.Format(iso8601),
			HTTPOnly: cookie.HttpOnly,
			Secure:   cookie.Secure,
		})
	}

	return out
}

func (b *Builder) NewRequest(pageId, id string, req *http.Request) *Entry {
	entry := &Entry{
		Pageref: pageId,
		Request: Request{
			Method:      req.Method,
			URL:         req.URL.String(),
			HTTPVersion: req.Proto,
			Cookies:     mapCookies(req.Cookies()),
			Headers:     mapHeaders(req.Header),
			BodySize:    int(req.ContentLength),
		},
	}
	b.Entries[id] = entry

	return entry
}

func (e *Entry) SetResponse(res *http.Response) {
	e.Response = Response{
		Status:      res.StatusCode,
		StatusText:  res.Status,
		HTTPVersion: res.Proto,
		BodySize:    int(res.ContentLength),
		HeadersSize: -1,
		Headers:     mapHeaders(res.Header),
		Cookies:     mapCookies(res.Cookies()),
	}
}
