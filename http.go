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

type HttpApiFunc func(s Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error

func createServer(s Server, graphDriver string) {
	log.Println("Creating introspection server for", s.container.Id)
	var sockPath string
	switch graphDriver {
	case "devmaper":
		sockPath = fmt.Sprintf("%s/%s/mnt/%s/rootfs/int.sock", *graphDir, graphDriver, s.container.Id)
	default:
		sockPath = fmt.Sprintf("%s/%s/mnt/%s/int.sock", *graphDir, graphDriver, s.container.Id)
	}

	if err := syscall.Unlink(sockPath); err != nil && !os.IsNotExist(err) {
		log.Print(err)
		return
	}

	r := createRouter(s)

	l, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatalf("Listen: %s", err)
		return
	}

	go s.monitor(l)

	httpSrv := http.Server{Addr: sockPath, Handler: r}
	httpSrv.Serve(l)
}

func createRouter(s Server) *mux.Router {
	r := mux.NewRouter()
	m := map[string]map[string]HttpApiFunc{
		"GET": {
			"/inspect":     getContainer,
			"/name":        getContainerName,
			"/id":          getContainerId,
			"/port/{port}": getContainerPort,
		},
	}

	for method, routes := range m {
		for route, fct := range routes {
			f := makeHttpHandler(s, fct)

			if route == "" {
				r.Methods(method).HandlerFunc(f)
				continue
			}

			r.Path(route).Methods(method).HandlerFunc(f)
		}
	}

	return r
}

func getContainer(s Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	c, err := s.client.FetchContainer(s.container.Id)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error: %", err), 500)
	}
	writeJson(w, c)
	return nil
}

func getContainerName(s Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	c, err := s.client.FetchContainer(s.container.Id)
	if err != nil {
		http.Error(w, "Error fetching container", 500)
		return err
	}

	return writeJson(w, strings.TrimPrefix(c.Name, "/"))
}
func getContainerId(s Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	return writeJson(w, s.container.Id)
}

func getContainerPort(s Server, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	c, err := s.client.FetchContainer(s.container.Id)
	if err != nil {
		http.Error(w, "Could not fetch container", 500)
		return err
	}

	var binds = []string{}
	for portStr, binding := range c.HostConfig.PortBindings {
		if strings.TrimSuffix(portStr, "/tcp") == vars["port"] || strings.TrimSuffix(portStr, "/udp") == vars["port"] {
			for _, b := range binding {
				if b.HostPort != "" {
					if b.HostIp == "" {
						b.HostIp = "0.0.0.0"
					}
					binds = append(binds, fmt.Sprintf("%s:%s", b.HostIp, b.HostPort))
				}
			}
			return writeJson(w, binds)
		}
	}
	return writeJson(w, binds)
}

func makeHttpHandler(s Server, h HttpApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(s, w, r, mux.Vars(r)); err != nil {
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
