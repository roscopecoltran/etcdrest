package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	etcd "github.com/coreos/go-etcd/etcd"
)

func setupEtcd() {
	env := getEnv()
	client = etcd.NewClient([]string{env.EtcdConn})
	s := "/"
	etcdDir = &s
}

func TestHandleIndexReturnsWithStatusOK(t *testing.T) {
	request, _ := http.NewRequest("GET", "/hosts", nil)
	response := httptest.NewRecorder()

	setupEtcd()

	getAllEntries(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("Non-expected status code%v:\n\tbody: %v", "200", response.Code)
	}
}
