package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"syscall"

	"github.com/cpuguy83/docker-grand-ambassador/docker"
	"github.com/gorilla/mux"
)

var socket = flag.String("sock", "/var/run/docker.sock", "Path to Docker socket")
var graphDir = flag.String("g", "/docker", "Path to Docker graph")

type Server struct {
	container *docker.Container
	client    docker.Docker
}

type HttpApiFunc func(s Server, w http.ResponseWriter, r *http.Request) error

func main() {
	flag.Parse()

	client, err := docker.NewClient(*socket)
	if err != nil {
		log.Fatal(err)
	}
	containers, err := client.FetchAllContainers()
	if err != nil {
		log.Fatal(err)
	}

	daemonInfo, err := client.Info()
	if err != nil {
		log.Fatal(err)
	}

	var servers = map[string]Server{}

	for _, c := range containers {
		server := Server{c, client}
		servers[c.Id] = server
		go createServer(server, daemonInfo.Driver)
	}
	<-make(chan struct{})
}

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
	fmt.Fprint(w, string(d))
	return nil
}
