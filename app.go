package main

import (
	"context"
	"log"
)

type Server interface {
	Start(ctx context.Context, srvErr chan error, srvChan chan bool) error
}

type App struct {
	Ctx         context.Context
	servers     []Server
	srvErr      chan error
	srvChans    []chan bool
	startErrors []error
}

func (a *App) AddServer(s Server) {
	a.srvChans = append(a.srvChans, make(chan bool))
	a.servers = append(a.servers, s)
	a.startErrors = append(a.startErrors, nil)
}

func (a *App) Start() {
	a.srvErr = make(chan error, len(a.srvChans))
	for idx := range a.servers {
		if idx == len(a.servers)-1 {
			a.startService(idx)
		} else {
			go a.startService(idx)
		}
	}
}

func (a *App) startService(idx int) {
	server := a.servers[idx]
	log.Println("Starting service: ", idx, server)
	err := server.Start(a.Ctx, a.srvErr, a.srvChans[idx])
	if err != nil {
		a.startErrors[idx] = err
		log.Fatal("Error Starting Server: ", idx, err)
		panic(err)
	}
}

func (a *App) Done(onStopped func()) {
	// Wait for interruption.
	select {
	case err := <-a.srvErr:
		log.Println("Server error: ", err)
		// Error when starting HTTP server or GRPC server
		return
	case <-a.Ctx.Done():
		// Wait for first CTRL+C.
		// Stop receiving signal notifications as soon as possible.
		if onStopped != nil {
			onStopped()
		}
	}
}
