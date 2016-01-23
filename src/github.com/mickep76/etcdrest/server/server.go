package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	//	"reflect"
	"strings"

	"github.com/evanphx/json-patch"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"

	//	"github.com/mickep76/etcdrest/config"
	"github.com/mickep76/etcdrest/etcd"
	"github.com/mickep76/etcdrest/log"
)

// Config interface.
type Config interface {
	Bind(string) Config
	APIVersion(string) Config
	Envelope(bool) Config
	Indent(bool) Config
	EtcdRoute(string, string, string)
	Run() error
}

// config struct.
type config struct {
	bind       string
	apiVersion string
	envelope   bool
	indent     bool
	session    etcd.Session
	router     *mux.Router
}

// New config constructor.
func New(session etcd.Session) Config {
	return &config{
		bind:       "0.0.0.0:8080",
		apiVersion: "v1",
		envelope:   false,
		indent:     true,
		session:    session,
		router:     mux.NewRouter(),
	}
}

func (c *config) Bind(bind string) Config {
	c.bind = bind
	return c
}

func (c *config) APIVersion(apiVersion string) Config {
	c.apiVersion = apiVersion
	return c
}

func (c *config) Envelope(envelope bool) Config {
	c.envelope = envelope
	return c
}

func (c *config) Indent(indent bool) Config {
	c.indent = indent
	return c
}

func (c *config) patchDoc(origDoc, patchDoc []byte) ([]byte, int, error) {
	// Prepare JSON patch.
	patch, err := jsonpatch.DecodePatch(patchDoc)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Apply JSON patch.
	doc, err := patch.ApplyIndent(origDoc, "  ")
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return doc, http.StatusOK, nil
}

func (c *config) validateDoc(doc []byte, path string, schema string) (int, []error) {
	// Prepare document and JSON schema.
	docLoader := gojsonschema.NewStringLoader(string(doc))
	schemaLoader := gojsonschema.NewReferenceLoader(schema)

	// Validate document using JSON schema.
	res, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return http.StatusInternalServerError, []error{err}
	}

	if !res.Valid() {
		var errors []error
		for _, e := range res.Errors() {
			errors = append(errors, fmt.Errorf("%s: %s", strings.Replace(e.Context().String("/"), "(root)", path, 1), e.Description()))
		}

		return http.StatusBadRequest, errors
	}

	return http.StatusOK, nil
}

// putOrPatchDoc put or patch document.
func (c *config) putOrPatchDoc(path string, schema string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path = path + "/" + mux.Vars(r)["name"]

		// Get request body.
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			c.writeError(w, r, err, http.StatusInternalServerError)
			return
		}
		if err := r.Body.Close(); err != nil {
			c.writeError(w, r, err, http.StatusInternalServerError)
			return
		}

		// Patch document using JSON patch RFC 6902.
		var doc []byte
		if r.Method == "PATCH" {
			data, code, err := c.session.Get(path)
			if err != nil {
				c.writeError(w, r, err, code)
				return
			}

			origDoc, err := json.Marshal(&data)
			if err != nil {
				c.writeError(w, r, err, http.StatusInternalServerError)
				return
			}

			doc, code, err = c.patchDoc(origDoc, body)
			if err != nil {
				c.writeError(w, r, err, code)
				return
			}
		} else {
			doc = body
		}

		// Validate document using JSON schema
		if code, errors := c.validateDoc(doc, path, schema); err != nil {
			c.writeErrors(w, r, errors, code)
			return
		}

		var data interface{}
		if err := json.Unmarshal(doc, &data); err != nil {
			c.writeError(w, r, err, http.StatusInternalServerError)
			return
		}

		// Create document.
		if code, err := c.session.Put(path, data); err != nil {
			c.writeError(w, r, err, code)
			return
		}

		c.write(w, r, data)
	}
}

// getDoc get document.
func (c *config) getDoc(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		if name != "" {
			path = path + "/" + name
		}

		doc, code, err := c.session.Get(path)
		if err != nil {
			c.writeError(w, r, err, code)
		}

		c.write(w, r, doc)
	}
}

// deleteDoc delete document.
func (c *config) deleteDoc(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := path + "/" + mux.Vars(r)["name"]

		data, code, err := c.session.Get(path)
		if err != nil {
			c.writeError(w, r, err, code)
		}

		if code, err := c.session.Delete(path); err != nil {
			c.writeError(w, r, err, code)
		}

		c.write(w, r, data)
	}
}

// EtcdRoute add route for etcd.
func (c *config) EtcdRoute(endpoint, path, schema string) {
	url := "/" + c.apiVersion + endpoint
	log.Infof("Add endpoint: %s etcd path: %s", url, path)

	c.router.HandleFunc(url, c.getDoc(path)).Methods("GET")
	c.router.HandleFunc(url+"/{name}", c.getDoc(path)).Methods("GET")
	c.router.HandleFunc(url+"/{name}", c.putOrPatchDoc(path, schema)).Methods("PUT")
	c.router.HandleFunc(url+"/{name}", c.putOrPatchDoc(path, schema)).Methods("PATCH")
	c.router.HandleFunc(url+"/{name}", c.deleteDoc(path)).Methods("DELETE")
}

// Run server.
func (c *config) Run() error {
	log.Infof("Bind to: %s", c.bind)
	logr := handlers.LoggingHandler(os.Stderr, c.router)
	return http.ListenAndServe(c.bind, logr)
}
