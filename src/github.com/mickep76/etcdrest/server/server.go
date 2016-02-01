package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/evanphx/json-patch"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/xeipuuv/gojsonschema"

	"github.com/mickep76/etcdrest/etcd"
	"github.com/mickep76/etcdrest/log"
)

var session etcd.Session

// Config interface.
type Config interface {
	TemplDir(string) Config
	SchemaURI(string) Config
	Bind(string) Config
	Envelope(bool) Config
	Indent(bool) Config
	RouteEtcd(string, string, string)
	RouteTemplate(string, string)
	RouteStatic(string, string)
	Run() error
}

// config struct.
type config struct {
	templDir  string
	schemaURI string
	bind      string
	envelope  bool
	indent    bool
	session   etcd.Session
	router    *mux.Router
}

// New config constructor.
func New(session etcd.Session) Config {
	return &config{
		templDir:  "templates",
		schemaURI: "file://schemas",
		bind:      "0.0.0.0:8080",
		envelope:  false,
		indent:    true,
		session:   session,
		router:    mux.NewRouter(),
	}
}

func (c *config) TemplDir(templDir string) Config {
	c.templDir = templDir
	return c
}

func (c *config) SchemaURI(schemaURI string) Config {
	c.schemaURI = schemaURI
	return c
}

func (c *config) Bind(bind string) Config {
	c.bind = bind
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

func (c *config) patchDoc(doc, patch []byte) ([]byte, error) {
	// Prepare JSON patch.
	p, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return nil, err
	}

	// Apply JSON patch.
	return p.ApplyIndent(doc, "  ")
}

func (c *config) validateDoc(doc []byte, path string, schema string) (int, []error) {
	// Prepare document and JSON schema.
	docLoader := gojsonschema.NewStringLoader(string(doc))
	log.Infof("Using schema URI: %s/%s", c.schemaURI, schema)
	schemaLoader := gojsonschema.NewReferenceLoader(c.schemaURI + "/" + schema)

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
		npath := path + "/" + mux.Vars(r)["name"]

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
			data, code, err := c.session.Get(npath)
			if err != nil {
				c.writeError(w, r, err, code)
				return
			}

			origDoc, err := json.Marshal(&data)
			if err != nil {
				c.writeError(w, r, err, http.StatusInternalServerError)
				return
			}

			doc, err = c.patchDoc(origDoc, body)
			if err != nil {
				c.writeError(w, r, err, code)
				return
			}
		} else {
			doc = body
		}

		// Validate document using JSON schema
		if code, errors := c.validateDoc(doc, npath, schema); errors != nil {
			c.writeErrors(w, r, errors, code)
			return
		}

		var data interface{}
		if err := json.Unmarshal(doc, &data); err != nil {
			c.writeError(w, r, err, http.StatusInternalServerError)
			return
		}

		// Create document.
		if code, err := c.session.Put(npath, data); err != nil {
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
		npath := path
		if name != "" {
			npath = npath + "/" + name
		}

		doc, code, err := c.session.Get(npath)
		if err != nil {
			c.writeError(w, r, err, code)
			return
		}

		c.write(w, r, doc)
	}
}

// deleteDoc delete document.
func (c *config) deleteDoc(path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		npath := path + "/" + mux.Vars(r)["name"]

		data, code, err := c.session.Get(npath)
		if err != nil {
			c.writeError(w, r, err, code)
			return
		}

		if code, err := c.session.Delete(npath); err != nil {
			c.writeError(w, r, err, code)
			return
		}

		c.write(w, r, data)
	}
}

// RouteEtcd add route for etcd.
func (c *config) RouteEtcd(endpoint, path, schema string) {
	url := endpoint
	log.Infof("Add endpoint: %s etcd path: %s schema: %s", url, path, schema)

	c.router.HandleFunc(url, c.getDoc(path)).Methods("GET")
	c.router.HandleFunc(url+"/{name}", c.getDoc(path)).Methods("GET")
	c.router.HandleFunc(url+"/{name}", c.putOrPatchDoc(path, schema)).Methods("PUT")
	c.router.HandleFunc(url+"/{name}", c.putOrPatchDoc(path, schema)).Methods("PATCH")
	c.router.HandleFunc(url+"/{name}", c.deleteDoc(path)).Methods("DELETE")
}

// RouteStatic add route for file system path.
func (c *config) RouteStatic(endpoint, path string) {
	url := endpoint
	log.Infof("Add endpoint: %s path: %s", url, path)

	static := http.StripPrefix(url, http.FileServer(http.Dir(path)))
	c.router.PathPrefix(url).Handler(static)
	http.Handle("/", c.router)
}

// Run server.
func (c *config) Run() error {
	session = c.session

	log.Infof("Bind to: %s", c.bind)
	logr := handlers.LoggingHandler(os.Stderr, c.router)
	return http.ListenAndServe(c.bind, logr)
}
