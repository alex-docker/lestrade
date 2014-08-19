package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/gorilla/mux"
)

type HttpApiFunc func(s Server, w http.ResponseWriter, r *http.Request) error

func createServer(s Server, graphDriver string) {
	sockPath := fmt.Sprintf("%s/%s/mnt/%s/lestrade.sock", *graphDir, graphDriver, s.container.Id)

	if err := syscall.Unlink(sockPath); err != nil && !os.IsNotExist(err) {
		log.Print(err)
		return
	}

	inspectHandler := makeHttpHandler(s, getContainer)
	nameHandler := makeHttpHandler(s, getContainerName)
	idHandler := makeHttpHandler(s, getContainerId)

	r := mux.NewRouter()

	r.HandleFunc("/inspect", inspectHandler).Methods("Get")
	r.HandleFunc("/name", nameHandler).Methods("Get")
	r.HandleFunc("/id", idHandler).Methods("Get")

	l, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatalf("Listen: %s", err)
		return
	}

	go s.monitor(l, sockPath)

	httpSrv := http.Server{Addr: sockPath, Handler: r}
	httpSrv.Serve(l)
}

func getContainer(s Server, w http.ResponseWriter, r *http.Request) error {
	c, err := s.client.FetchContainer(s.container.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %", err), 500)
	}
	writeJson(w, c)
	return nil
}

func getContainerName(s Server, w http.ResponseWriter, r *http.Request) error {
	c, err := s.client.FetchContainer(s.container.Id)
	if err != nil {
		http.Error(w, "Error fetching container", 500)
		return err
	}

	return writeJson(w, strings.TrimPrefix(c.Name, "/"))
}
func getContainerId(s Server, w http.ResponseWriter, r *http.Request) error {
	return writeJson(w, s.container.Id)
}

func makeHttpHandler(s Server, h HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(s, w, r); err != nil {
			log.Print(err)
		}
	}
}

func writeJson(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")

	d, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "Error", 500)
		return err
	}
	w.Write(d)
	return nil
}
