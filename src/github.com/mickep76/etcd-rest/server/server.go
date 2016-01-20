package server

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/coreos/etcd/client"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/mickep76/etcdmap"
	//	"github.com/xeipuuv/gojsonschema"
	"golang.org/x/net/context"

	"github.com/mickep76/etcd-rest/config"
	"github.com/mickep76/etcd-rest/etcd"
	"github.com/mickep76/etcd-rest/log"
)

func Get(cfg *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		path := route.Path
		name := mux.Vars(r)["name"]
		if name != "" {
			path = path + "/" + name
		}

		log.Infof("Get path: %s from etcd", path)
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

		/*
			docLoader := gojsonschema.NewStringLoader(string(body))
			schemaLoader := gojsonschema.NewStringLoader(route.SchemaJSON)

			result, err := gojsonschema.Validate(schemaLoader, docLoader)
			if err != nil {
				log.Fatalf(err.Error())
			}

			if !result.Valid() {
				var errors []string
				for _, e := range result.Errors() {
					errors = append(errors, fmt.Sprintf("%s: %s\n", strings.Replace(e.Context().String("/"), "(root)", route.Endpoint+"/"+name, 1), e.Description()))
				}
				b, _ := json.MarshalIndent(&errors, "", "	")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write(b)
				return
			}
		*/
		log.Infof("Create path: %s in etcd", path)
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

		log.Infof("Delete path: %s from etcd", path)
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

	route := config.Route{
		Regexp:   "^/hosts$",
		Path:     "/hosts",
		Endpoint: "/hosts",
	}

	r.HandleFunc("/hosts", Get(cfg, &route, kapi)).
		Methods("GET")
	r.HandleFunc("/hosts/{name}", Create(cfg, &route, kapi)).
		Methods("PUT")
	r.HandleFunc("/hosts/{name}", Get(cfg, &route, kapi)).
		Methods("GET")
	r.HandleFunc("/hosts/{name}", Delete(cfg, &route, kapi)).
		Methods("DELETE")

	// Fire up the server
	log.Infof("Bind to: %s", cfg.Bind)
	logr := handlers.LoggingHandler(os.Stdout, r)
	err := http.ListenAndServe(cfg.Bind, logr)
	if err != nil {
		log.Fatal(err.Error())
	}
}
