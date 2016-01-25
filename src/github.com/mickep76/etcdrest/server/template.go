package server

import (
	"fmt"
	"net/http"
	"strings"
	"text/template"

	"github.com/gorilla/mux"

	"github.com/mickep76/etcdrest/log"
)

func Center(size int, deco string, str string) string {
	if size < len(str) {
		return str
	}

	pad := (size - len(str)) / 2
	lpad := pad
	rpad := size - len(str) - lpad

	return fmt.Sprintf("%s%s%s", strings.Repeat(deco, lpad), str, strings.Repeat(deco, rpad))
}

func Substr(s string, b int, l int) string {
	return s[b:l]
}

var funcs = template.FuncMap{
	"center": Center,
	"substr": Substr,
}

var templates = template.Must(template.New("main").Funcs(funcs).ParseGlob("templates/*.tmpl"))

// RouteTempl add route for Go Text Template.
func (c *config) RouteTemplate(endpoint, path, templ string) {
	url := "/" + c.apiVersion + endpoint
	log.Infof("Add endpoint: %s etcd path: %s template: %s", url, path, templ)

	c.router.HandleFunc(url, c.getTemplate(path, templ)).Methods("GET")
}

// getDoc get document.
func (c *config) getTemplate(path string, templ string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		npath := path
		if name != "" {
			npath = npath + "/" + name
		}

		data, code, err := c.session.Get(npath)
		if err != nil {
			c.writeError(w, r, err, code)
		}

		input := map[string]interface{}{
			"vars":   mux.Vars(r),
			"params": r.URL.Query(),
			"data":   data,
		}

		// Write template.
		// Need to exec first otherwise multiple headers on error
		if err := templates.ExecuteTemplate(w, templ, input); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}
