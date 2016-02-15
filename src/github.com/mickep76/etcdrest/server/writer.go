package server

import (
	"encoding/json"
	"net/http"
	"strings"
)

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

	if envelope == false {
		c.writeMIME(w, r, data)
	} else {
		e := map[string]interface{}{
			"code": http.StatusOK,
			"data": data,
		}

		c.writeMIME(w, r, e)
	}
}

func (c *config) writeErrors(w http.ResponseWriter, r *http.Request, errors []error, code int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	envelope := c.envelope
	switch strings.ToLower(r.URL.Query().Get("envelope")) {
	case "true":
		envelope = true
	case "false":
		envelope = false
	}

	var s []string
	for _, err := range errors {
		s = append(s, err.Error())
	}

	if envelope == false {
		c.writeMIME(w, r, s)
	} else {
		e := map[string]interface{}{
			"code": code,
			"data": s,
		}

		c.writeMIME(w, r, e)
	}
}

func (c *config) writeError(w http.ResponseWriter, r *http.Request, err error, code int) {
	c.writeErrors(w, r, []error{err}, code)
}

func (c *config) writeMIME(w http.ResponseWriter, r *http.Request, data interface{}) {
	indent := c.indent
	switch strings.ToLower(r.URL.Query().Get("indent")) {
	case "true":
		indent = true
	case "false":
		indent = false
	}

	var b []byte
	if indent == false {
		b, _ = json.Marshal(data)
	} else {
		b, _ = json.MarshalIndent(data, "", "  ")
	}
	w.Write(b)
}
