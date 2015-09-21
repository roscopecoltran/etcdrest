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
	"strings"

	etcd "github.com/coreos/go-etcd/etcd"
	"github.com/gorilla/mux"
	"github.com/mickep76/etcdmap"
	"github.com/xeipuuv/gojsonschema"
)

type JSONError struct {
	message string `json:"message"`
}

// Env variables.
type Env struct {
	EtcdConn string
	Bind     string
	Port     string
}

// Endpoint configuration.
type Endpoint struct {
	Path      string `json:"path"`
	Desc      string `json:"desc"`
	SchemaURL string `json:"schema"`
}

// getEnv variables.
func getEnv() Env {
	env := Env{}
	for _, e := range os.Environ() {
		a := strings.Split(e, "=")
		switch a[0] {
		case "ETCD_CONN":
			env.EtcdConn = a[1]
		case "ETCD_REST_PORT":
			env.Port = a[1]
		case "ETCD_REST_BIND":
			env.Bind = a[1]
		}
	}

	return env
}

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

func getAllEntries(w http.ResponseWriter, r *http.Request) {
	res, err := client.Get(*etcdDir+"/host", true, true)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusNoContent)
		w.Write(errToJSON(err))
		return
	}
	j, err := etcdmap.JSONIndent(res.Node, "    ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(j)
}

func getEntry(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	res, err := client.Get(*etcdDir+"/host/"+name, true, true)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	j, err := etcdmap.JSONIndent(res.Node, "    ")
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

func updateEntry(w http.ResponseWriter, r *http.Request) {
}

func createEntry(endpoint Endpoint) func(w http.ResponseWriter, r *http.Request) {
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

		docLoader := gojsonschema.NewStringLoader(string(j))

		schemaLoader := gojsonschema.NewReferenceLoader(endpoint.SchemaURL)

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

			b, _ := json.MarshalIndent(&m, "", "    ")

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			w.Write(b)
			return
		}

		if err = etcdmap.CreateJSON(client, "/host/"+name, j); err != nil {
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

func deleteEntry(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]

	_, err := client.Delete("/host/"+name, true)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNoContent)
}

var client *etcd.Client
var etcdDir *string
var schema []byte
var endpoints []Endpoint

func main() {
	// Define endpoints.
	endpoints = []Endpoint{
		Endpoint{
			Path:      "/hosts",
			Desc:      "Host",
			SchemaURL: "http://localhost:8080/schemas/host.json",
		},
		Endpoint{
			Path:      "/hosts/{host}/interfaces",
			Desc:      "Host Interface",
			SchemaURL: "http://localhost:8080/schemas/interface.json",
		},
	}

	// Get app env variables.
	env := getEnv()

	// Set defaults
	if env.Bind == "" {
		env.Bind = "127.0.0.1"
	}
	if env.Port == "" {
		env.Port = "8080"
	}

	// Options.
	version := flag.Bool("version", false, "Version")
	etcdNode := flag.String("etcd-node", "", "Etcd node")
	etcdPort := flag.String("etcd-port", "2379", "Etcd port")
	etcdDir = flag.String("etcd-dir", "/", "Etcd directory")
	bindAddr := flag.String("bind-addr", env.Bind, "Bind to address")
	port := flag.String("port", env.Port, "Port")
	flag.Parse()

	// Print version.
	if *version {
		fmt.Printf("etcd-drowsy %s\n", Version)
		os.Exit(0)
	}

	// Validate input.
	if env.EtcdConn == "" && *etcdNode == "" {
		log.Fatalf("You need to specify Etcd host.")
	}

	// Setup Etcd client.
	conn := env.EtcdConn
	if *etcdNode != "" {
		conn = fmt.Sprintf("http://%v:%v", *etcdNode, *etcdPort)
	}
	client = etcd.NewClient([]string{conn})

	r := mux.NewRouter()

	for _, e := range endpoints {
		r.HandleFunc(e.Path, getAllEntries).
			Methods("GET")
		r.HandleFunc(e.Path+"/{name}", getEntry).
			Methods("GET")
		r.HandleFunc(e.Path+"/{name}", createEntry(e)).
			Methods("PUT")
		r.HandleFunc(e.Path+"/{name}", deleteEntry).
			Methods("DELETE")
		r.HandleFunc(e.Path+"/{name}", updateEntry).
			Methods("PATCH")
	}

	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("static/"))))
	http.ListenAndServe(*bindAddr+":"+*port, r)
}
