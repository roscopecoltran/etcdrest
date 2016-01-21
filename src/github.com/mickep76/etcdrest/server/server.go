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

func CreateUpdateOrPatch(cfg *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		path := route.Path + "/" + name

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
		}
		if err := r.Body.Close(); err != nil {
			writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
		}

		// Patch existing document using JSON Patch RFC 6902
		if r.Method == "PATCH" {
			patch := body

			res, err := kapi.Get(context.TODO(), path, &client.GetOptions{Recursive: true})
			if err != nil {
				// Path doesn't exist
				if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Error retrieving data
				writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
				return
			}

			data, err := etcdmap.JSON(res.Node)
			if err != nil {
				writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
			}

			p, err := jsonpatch.DecodePatch(patch)
			if err != nil {
				writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
			}

			var err2 error
			body, err2 = p.ApplyIndent(data, "  ")
			if err2 != nil {
				writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
			}
		}

		docLoader := gojsonschema.NewStringLoader(string(body))
		schemaLoader := gojsonschema.NewReferenceLoader(route.Schema)

		result, err := gojsonschema.Validate(schemaLoader, docLoader)
		if err != nil {
			writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
		}

		if !result.Valid() {
			var errors []string
			for _, e := range result.Errors() {
				errors = append(errors, fmt.Sprintf("%s: %s", strings.Replace(e.Context().String("/"), "(root)", path, 1), e.Description()))
			}

			writeErrors(cfg, w, r, errors, http.StatusBadRequest)
			return
		}

		var d interface{}
		if err := json.Unmarshal(body, &d); err != nil {
			writeError(cfg, w, r, err.Error(), http.StatusBadRequest)
		}

		if err = etcdmap.Create(kapi, path, reflect.ValueOf(d)); err != nil {
			writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
			return
		}

		write(cfg, w, r, d)
	}
}

func GetOneOrAll(cfg *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := route.Path
		name := mux.Vars(r)["name"]
		if name != "" {
			path = path + "/" + name
		}

		res, err := kapi.Get(context.TODO(), path, &client.GetOptions{Recursive: true})
		if err != nil {
			// Path doesn't exist
			if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Error retrieving data
			writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
			return
		}

		m := etcdmap.Map(res.Node)
		write(cfg, w, r, m)
	}
}

func Delete(cfg *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		path := route.Path + "/" + name

		if _, err := kapi.Delete(context.Background(), path, &client.DeleteOptions{Recursive: true}); err != nil {
			// Path doesn't exist
			if cerr, ok := err.(client.Error); ok && cerr.Code == 100 {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			// Error retrieving data
			writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
	}
}

func Run(cfg *config.Config) {
	// Connect to etcd.
	log.Infof("Connecting to etcd: %s", cfg.Etcd.Peers)
	kapi := etcd.NewKeyAPI(cfg)

	// Create new router.
	r := mux.NewRouter()

	for _, route := range *cfg.Routes {
		path := "/" + cfg.APIVersion + route.Endpoint
		log.Infof("Add endpoint: %s etcd path: %s", path, route.Path)
		r.HandleFunc(path, GetOneOrAll(cfg, &route, kapi)).
			Methods("GET")
		r.HandleFunc(path+"/{name}", GetOneOrAll(cfg, &route, kapi)).
			Methods("GET")
		r.HandleFunc(path+"/{name}", CreateUpdateOrPatch(cfg, &route, kapi)).
			Methods("PUT")
		r.HandleFunc(path+"/{name}", CreateUpdateOrPatch(cfg, &route, kapi)).
			Methods("PATCH")
		r.HandleFunc(path+"/{name}", Delete(cfg, &route, kapi)).
			Methods("DELETE")
	}

	// Fire up the server
	log.Infof("Bind to: %s", cfg.Bind)
	logr := handlers.LoggingHandler(os.Stdout, r)
	err := http.ListenAndServe(cfg.Bind, logr)
	if err != nil {
		log.Fatal(err.Error())
	}
}
