package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"

	etcd "github.com/coreos/go-etcd/etcd"
	"github.com/gorilla/mux"
	"github.com/mickep76/etcdmap"
	//	"github.com/xeipuuv/gojsonschema"
	"github.com/gorilla/handlers"
)

// JSONError structure.
type JSONError struct {
	message string `json:"message"`
}

// Env variables.
type Env struct {
	Peers string
	Bind  string
}

// Route structure.
type Route struct {
	Path     string `json:"path"`
	Endpoint string `json:"endpoint"`
	Schema   string `json:"schema"`
}

// getEnv variables.
func getEnv() Env {
	env := Env{}
	env.Peers = "http://127.0.0.1:4001,http://127.0.0.1:2379"
	env.Bind = "127.0.0.1:8080"

	for _, e := range os.Environ() {
		a := strings.Split(e, "=")
		switch a[0] {
		case "ETCD_PEERS":
			env.Peers = a[1]
		case "ETCD_BIND":
			env.Bind = a[1]
		}
	}

	return env
}

// errToJSON convert error to JSON.
func errToJSON(err error) []byte {
	e := JSONError{
		message: err.Error(),
	}

	j, err := json.Marshal(&e)
	if err != nil {
		return []byte{}
	}

	return j
}

// getAllEntries return all entries for an endpoint.
func getAllEntries(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res, err := client.Get(route.Endpoint, true, true)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusNoContent)
			w.Write(errToJSON(err))
			return
		}
		j, err := etcdmap.JSONIndent(res.Node, "	")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	}
}

// getEntry return a single entry for an endpoint.
func getEntry(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		res, err := client.Get(route.Endpoint+"/"+name, true, true)
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		j, err := etcdmap.JSONIndent(res.Node, "	")
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errToJSON(err))
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(j)
	}
}

// createEntry for an endpoint.
func createEntry(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			panic(err)
		}
		if err := r.Body.Close(); err != nil {
			panic(err)
		}

		j := body
		/*
			docLoader := gojsonschema.NewStringLoader(string(j))

			schemaLoader := gojsonschema.NewReferenceLoader(route.Schema)

			fmt.Println(schemaLoader)

			result, err := gojsonschema.Validate(schemaLoader, docLoader)
			if err != nil {
				panic(err.Error())
			}

			if !result.Valid() {
				m := map[string]string{}
				for _, e := range result.Errors() {
					m[e.Field()] = e.Description()
				}

				b, _ := json.MarshalIndent(&m, "", "	")

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write(b)
				return
			}
		*/

		if err = etcdmap.CreateJSON(client, route.Endpoint+"/"+name, j); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(errToJSON(err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusCreated)
		w.Write(body)
	}
}

func deleteEntry(route Route) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		_, err := client.Delete(route.Endpoint+"/"+name, true)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
	}
}

var client *etcd.Client
var schemas []byte
var routes []Route

func main() {
	// Get env variables.
	env := getEnv()

	// Options.
	version := flag.Bool("version", false, "Version")
	peers := flag.String("peers", env.Peers, "Comma separated list of etcd nodes")
	bind := flag.String("bind", env.Bind, "Bind to address and port")
	flag.Parse()

	// Print version.
	if *version {
		fmt.Printf("etcd-rest %s\n", Version)
		os.Exit(0)
	}

	// Setup etcd client.
	log.Printf("Connecting to etcd: %s", *peers)
	client = etcd.NewClient(strings.Split(*peers, ","))

	// Get routes.
	res, err := client.Get("/routes", true, true)
	if err != nil {
		log.Fatal(err.Error())
	}

	for _, v := range etcdmap.Map(res.Node) {
		switch reflect.ValueOf(v).Kind() {
		case reflect.Map:
			m := v.(map[string]interface{})

			path, ok := m["path"]
			endpoint, ok2 := m["endpoint"]
			schema, ok3 := m["schema"]
			if ok && ok2 && ok3 {
				routes = append(routes, Route{
					Path:     path.(string),
					Endpoint: endpoint.(string),
					Schema:   schema.(string),
				})
				log.Printf("Adding endpoint: %s", path)
			}
		}
	}

	// Check that there are any registered endpoints
	if len(routes) < 1 {
		log.Fatal("No registered endpoints")
	}

	// Create new router.
	r := mux.NewRouter()
	logr := handlers.LoggingHandler(os.Stdout, r)

	for _, e := range routes {
		r.HandleFunc(e.Path, getAllEntries(e)).
			Methods("GET")
		r.HandleFunc(e.Path+"/{name}", getEntry(e)).
			Methods("GET")
		r.HandleFunc(e.Path+"/{name}", createEntry(e)).
			Methods("PUT")
		r.HandleFunc(e.Path+"/{name}", deleteEntry(e)).
			Methods("DELETE")
	}

	http.ListenAndServe(*bind, logr)
}
