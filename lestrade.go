package main

import (
	"flag"
	"log"
	"net"

	"github.com/cpuguy83/docker-grand-ambassador/docker"
)

var socket = flag.String("sock", "/var/run/docker.sock", "Path to Docker socket")
var graphDir = flag.String("g", "/docker", "Path to Docker graph")

var Servers = map[string]Server{}

type Server struct {
	container *docker.Container
	client    docker.Docker
	sigChan   chan bool
}

func (s *Server) monitor(l net.Listener) {
	log.Println("Starting monitor for:", s.container.Id)
	for sig := range s.sigChan {
		if sig {
			break
		}
	}
	go log.Println("Stopping introspection server for", s.container.Id)
	l.Close()
}

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

	go func() {
		for _, c := range containers {
			c, err = client.FetchContainer(c.Id)
			if err != nil {
				continue
			}
			if c.State.Running {
				server := Server{c, client, make(chan bool)}
				Servers[c.Id] = server
				go createServer(server, daemonInfo.Driver)
			}
		}
	}()

	events := client.GetEvents()
	go handleEvents(events, client, daemonInfo.Driver)

	<-make(chan struct{})
}

func handleEvents(eventChan chan *docker.Event, client docker.Docker, graphDriver string) {
	for e := range eventChan {
		switch e.Status {
		case "start", "restart":
			go handleStartEvent(e, client, graphDriver)
		case "stop", "die", "kill":
			go handleStopEvent(e, client, graphDriver)
		}
	}
}

func handleStartEvent(e *docker.Event, client docker.Docker, graphDriver string) {
	c, err := client.FetchContainer(e.ContainerId)
	if err != nil {
		log.Println("EventStart: Could not fetch container:", err)
		return
	}

	if _, exists := Servers[c.Id]; !exists {
		s := Server{c, client, make(chan bool)}
		Servers[c.Id] = s
	}
	createServer(Servers[c.Id], graphDriver)
}

func handleStopEvent(e *docker.Event, client docker.Docker, graphDriver string) {
	if s, exists := Servers[e.ContainerId]; exists {
		s.sigChan <- true
	}
}
