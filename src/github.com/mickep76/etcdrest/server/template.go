package server

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"text/template"

	"github.com/gorilla/mux"
	"github.com/robertkrimen/otto"

	"github.com/mickep76/etcdrest/log"
)

func jsSet(key string, val interface{}) error {
	return vm.Set(key, val)
}

func jsRun(code string) (otto.Value, error) {
	return vm.Run(code)
}

func getSubnet(ip, netmask string) (string, error) {
	// Check ip
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IP address %s", ip)
	}

	// Check netmask
	maskIP := net.ParseIP(netmask).To4()
	if maskIP == nil {
		return "", fmt.Errorf("invalid Netmask %s", netmask)
	}

	// Get prefix
	mask := net.IPMask(maskIP)
	prefix, _ := mask.Size()

	// Get network
	sPrefix := strconv.Itoa(prefix)
	_, network, err := net.ParseCIDR(ip + "/" + sPrefix)
	if err != nil {
		return "", err
	}

	return network.IP.String() + "/" + sPrefix, nil
}

func center(size int, deco string, str string) string {
	if size < len(str) {
		return str
	}

	pad := (size - len(str)) / 2
	lpad := pad
	rpad := size - len(str) - lpad

	return fmt.Sprintf("%s%s%s", strings.Repeat(deco, lpad), str, strings.Repeat(deco, rpad))
}

func substr(s string, b int, l int) string {
	return s[b:l]
}

func lastval(l []string) string {
	for i := len(l) - 1; i >= 0; i-- {
		if l[i] != "" {
			return l[i]
		}
	}

	return ""
}

func lastvaln(l ...string) string {
	return lastval(l)
}

func get(path string) (interface{}, error) {
	data, code, err := session.Get(path, false, "")

	if code == http.StatusNotFound {
		return nil, nil
	}

	return data, err
}

func getKeys(paths ...string) ([]string, error) {
	arr, code, err := session.GetKeys(paths...)

	if code == http.StatusNotFound {
		return nil, nil
	}

	return arr, err
}

func replace(oldStr string, newStr string, str string) string {
	return strings.Replace(str, oldStr, newStr, -1)
}

var funcs = template.FuncMap{
	"center":    center,
	"substr":    substr,
	"get":       get,
	"getkeys":   getKeys,
	"lastval":   lastval,
	"lastvaln":  lastvaln,
	"replace":   replace,
	"setjs":     jsSet,
	"runjs":     jsRun,
	"getsubnet": getSubnet,
}

var templates *template.Template
var vm *otto.Otto

// RouteTempl add route for Go Text Template.
func (c *config) RouteTemplate(endpoint, templ string) {
	if templates == nil {
		templates = template.Must(template.New("main").Funcs(funcs).ParseGlob(c.templDir + "/*.tmpl"))
		vm = otto.New()
	}

	url := endpoint
	log.Infof("Add endpoint: %s template: %s", url, templ)
	c.router.HandleFunc(url, c.getTemplate(templ)).Methods("GET")
}

func (c *config) getTemplate(templ string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		input := map[string]interface{}{
			"vars":       mux.Vars(r),
			"params":     r.URL.Query(),
			"server_uri": c.serverURI,
		}

		// Write template.
		b := new(bytes.Buffer)
		if err := templates.ExecuteTemplate(b, templ, input); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Write(b.Bytes())
	}
}
