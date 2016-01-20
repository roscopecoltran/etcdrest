package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mickep76/etcdrest/config"
)

// Envelope struct.
type Envelope struct {
	Code   int         `json:"code"`
	Data   interface{} `json:"data,omitempty"`
	Errors []string    `json:"errors,omitempty"`
}

func write(c *config.Config, w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	envelope := c.Envelope
	switch strings.ToLower(r.URL.Query().Get("envelope")) {
	case "true":
		envelope = true
	case "false":
		envelope = false
	}

	if envelope == false {
		b, _ := json.MarshalIndent(data, "", "  ")
		w.Write(b)
		return
	}

	s := Envelope{
		Code: http.StatusOK,
		Data: data,
	}

	b, _ := json.MarshalIndent(&s, "", "  ")
	w.Write(b)
}

func writeErrors(c *config.Config, w http.ResponseWriter, r *http.Request, e []string, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	envelope := c.Envelope
	switch strings.ToLower(r.URL.Query().Get("envelope")) {
	case "true":
		envelope = true
	case "false":
		envelope = false
	}

	if envelope == false {
		b, _ := json.MarshalIndent(e, "", "  ")
		w.Write(b)
		return
	}

	s := Envelope{
		Code:   code,
		Errors: e,
	}

	b, _ := json.MarshalIndent(&s, "", "  ")
	w.Write(b)
}

func writeError(c *config.Config, w http.ResponseWriter, r *http.Request, e string, code int) {
	writeErrors(c, w, r, []string{e}, code)
}
