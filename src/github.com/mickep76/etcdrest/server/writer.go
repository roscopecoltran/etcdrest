package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

// envelope struct.
type envelope struct {
	code   int         `json:"code"`
	data   interface{} `json:"data,omitempty"`
	errors []string    `json:"errors,omitempty"`
}

func (c *config) write(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	envelope := c.envelope
	switch strings.ToLower(r.URL.Query().Get("envelope")) {
	case "true":
		envelope = true
	case "false":
		envelope = false
	}

	indent := c.indent
	switch strings.ToLower(r.URL.Query().Get("indent")) {
	case "true":
		indent = true
	case "false":
		indent = false
	}

	var outp interface{}
	if envelope == false {
		outp = data
	} else {
		outp = envelope{
			code: http.StatusOK,
			data: data,
		}
	}

	var b []byte
	if indent == false {
		b, _ = json.Marshal(data, "", "  ")
	} else {
		b, _ = json.MarshalIndent(data, "", "  ")
	}
	w.Write(b)
}

func (c *config) writeErrors(w http.ResponseWriter, r *http.Request, errors interface{}, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	envelope := c.envelope
	switch strings.ToLower(r.URL.Query().Get("envelope")) {
	case "true":
		envelope = true
	case "false":
		envelope = false
	}

	indent := c.indent
	switch strings.ToLower(r.URL.Query().Get("indent")) {
	case "true":
		indent = true
	case "false":
		indent = false
	}

	var outp interface{}
	if envelope == false {
		outp = data
	} else {
		outp = envelope{
			code:   http.StatusOK,
			errors: errors,
		}
	}

	var b []byte
	if indent == false {
		b, _ = json.Marshal(data, "", "  ")
	} else {
		b, _ = json.MarshalIndent(data, "", "  ")
	}
	w.Write(b)
}
