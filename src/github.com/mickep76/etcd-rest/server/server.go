package server

import (
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

func all(cfg *config.Config, route *config.Route, kapi client.KeysAPI) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {

		// Retrieve data
		res, err := kapi.Get(context.TODO(), route.Path, &client.GetOptions{Recursive: true})
		if err != nil {
			// Directory doesn't exist
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

	r.HandleFunc("/hosts", all(cfg, &route, kapi)).
		Methods("GET")

	// Fire up the server
	log.Infof("Bind to: %s", cfg.Bind)
	logr := handlers.LoggingHandler(os.Stdout, r)
	err := http.ListenAndServe(cfg.Bind, logr)
	if err != nil {
		log.Fatal(err.Error())
	}
}
