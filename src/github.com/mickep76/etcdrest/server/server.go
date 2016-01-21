package server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/coreos/etcd/client"
	"github.com/evanphx/json-patch"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mickep76/etcdmap"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/net/context"

	"github.com/mickep76/etcdrest/config"
	"github.com/mickep76/etcdrest/etcd"
	"github.com/mickep76/etcdrest/log"
)

// Server struct.
type Server struct {
	Bind       string
	APIVersion string
	Envelope   bool
	
	router	*mux.Router
	etcd *etcd.Etcd
}

func patchDoc(origDoc, patchDoc []byte) ([]byte, int, error)
	// Prepare JSON patch.
	patch, err := jsonpatch.DecodePatch(patch)
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	// Apply JSON patch.			
	doc, err = patch.ApplyIndent(inpDoc, "  ")
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	return doc, http.StatusOK, nil
}

func validateDoc(doc []byte, schema string) (int, []error) {
	// Prepare document and JSON schema.
	docLoader := gojsonschema.NewStringLoader(string(doc))
	schemaLoader := gojsonschema.NewReferenceLoader(schema)
		
	// Validate document using JSON schema.
	res, err := gojsonschema.Validate(schemaLoader, docLoader)
	if err != nil {
		return http.StatusInternalServerError, err
	}
		
	if !res.Valid() {
		var errors []string
		for _, e := range res.Errors() {
			errors = append(errors, fmt.Sprintf("%s: %s", strings.Replace(e.Context().String("/"), "(root)", path, 1), e.Description()))
		}

		return  http.StatusBadRequest, errors
	}

	return http.StatusOK, nil
}

// putOrPatchDoc put or patch document.
func putOrPatchDoc(srv *Server, path string, schema string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path = path + "/" + mux.Vars(r)["name"]

		// Get request body.
		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			writeError(srv, w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := r.Body.Close(); err != nil {
			writeError(srv, w, r, err.Error(), http.StatusInternalServerError)
			return
		}

		// Patch document using JSON patch RFC 6902.
		if r.Method == "PATCH" {
			origDoc, code, err := etcd.Get(path)
			if code != http.StatusOK {
				write(srv, w, r, doc, code)
				return
			}

			doc, code, err := patchDoc(origDoc, body)
			if err != nil {
				write(srv, w, r, doc, code)
				return
			}
		}

		// Validate document using JSON schema
		code, err := validateDoc(body, schema)
		if err != nil {
			write(srv, w, r, doc, code)
			return
		}

		// Create document.
		data, code, err := etcd.Create(path, body)
		if err != nil {
			write(srv, w, r, doc, code)
			return
		}

		write(srv, w, r, data, http.StatusOK)
	}
}

// getDoc get document.
func getDoc(srv *Server, path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		if name != "" {
			path = path + "/" + name
		}

		doc, code, err := etcd.Get(path)
		if code != http.StatusOK {
			write(srv, w, r, doc, code)
		}

		writeDoc(cfg, w, r, doc)
	}
}

// deleteDoc delete document.
func deleteDoc(srv *Server, path string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := rpath + "/" + mux.Vars(r)["name"]

		doc, code, err := etcd.Get(path)
		if code != http.StatusOK {
			write(srv, w, r, doc, code)
		}
		
		code, err := etcd.Delete(path)
		writeDoc(srv, w, r, doc, code)
	}
}

// New server.
func New(routes cfg.Routes) (*Server) {
	srv := &Server{}

	// Connect to etcd.
//	log.Infof("Connecting to etcd: %s", cfg.Etcd.Peers)
//	srv.kapi := etcd.NewKeyAPI(cfg)

	// Create new router.
	srv.router := mux.NewRouter()

	// Add routes.
	for _, route := range routes {
		path := "/" + srv.APIVersion + route.Endpoint
		log.Infof("Add endpoint: %s etcd path: %s", path, route.Path)
		srv.router.HandleFunc(path, getDoc(srv, path)).Methods("GET")
		srv.router.HandleFunc(path+"/{name}", getDoc(srv, path)).Methods("GET")
		srv.router.HandleFunc(path+"/{name}", putOrPatch(srv, path, route.Schema)).Methods("PUT")
		srv.router.HandleFunc(path+"/{name}", putOrPatch(cfg, path, route.Schema)).Methods("PATCH")
		srv.router.HandleFunc(path+"/{name}", deleteDoc(cfg, path)).Methods("DELETE")
	}
}

// Run server.
func (srv *Server) Run() {
	// Fire up the server
	log.Infof("Bind to: %s", cfg.Bind)
	logr := handlers.LoggingHandler(os.Stdout, r)
	err := http.ListenAndServe(cfg.Bind, logr)
	if err != nil {
		log.Fatal(err.Error())
	}
}