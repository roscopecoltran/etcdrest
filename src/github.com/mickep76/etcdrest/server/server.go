package server

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/coreos/etcd/client"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mickep76/etcdmap"
	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/net/context"

	"github.com/mickep76/etcdrest/config"
	"github.com/mickep76/etcdrest/etcd"
	"github.com/mickep76/etcdrest/log"
)

func Get(cfg *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
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

func Create(cfg *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		path := route.Path + "/" + name

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err := r.Body.Close(); err != nil {
			panic(err)
		}

		docLoader := gojsonschema.NewStringLoader(string(body))
		schemaLoader := gojsonschema.NewReferenceLoader(route.Schema)

		result, err := gojsonschema.Validate(schemaLoader, docLoader)
		if err != nil {
			log.Fatalf(err.Error())
		}

		if !result.Valid() {
			var errors []string
			for _, e := range result.Errors() {
				errors = append(errors, fmt.Sprintf("%s: %s", strings.Replace(e.Context().String("/"), "(root)", route.Endpoint+"/"+name, 1), e.Description()))
			}

			writeErrors(cfg, w, r, errors, http.StatusBadRequest)
			return
		}

		if err = etcdmap.CreateJSON(kapi, path, body); err != nil {
			writeError(cfg, w, r, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(body)
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
		log.Infof("Add endpoint: /%s%s etcd path: %s", cfg.APIVersion, route.Endpoint, route.Path)
		r.HandleFunc("/"+cfg.APIVersion+route.Endpoint, Get(cfg, &route, kapi)).
			Methods("GET")
		r.HandleFunc("/"+cfg.APIVersion+route.Endpoint+"/{name}", Create(cfg, &route, kapi)).
			Methods("PUT")
		r.HandleFunc("/"+cfg.APIVersion+route.Endpoint+"/{name}", Get(cfg, &route, kapi)).
			Methods("GET")
		r.HandleFunc("/"+cfg.APIVersion+route.Endpoint+"/{name}", Delete(cfg, &route, kapi)).
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
