package server

import (
	"context"

	goal "github.com/panyam/goapplib"
)

type WebAppServer struct {
	goal.WebAppServer
}

// NewWebAppServerConfig creates a WebAppServer configuration
func NewWebAppServerConfig(address, grpcAddress string, allowLocalDev bool) goal.WebAppServer {
	return goal.WebAppServer{
		Address:       address,
		GrpcAddress:   grpcAddress,
		AllowLocalDev: allowLocalDev,
	}
}

func (s *WebAppServer) Start(ctx context.Context, srvErr chan error, stopChan chan bool) error {
	sdlApp, _, _ := NewSdlApp(s.GrpcAddress)
	return s.StartWithHandler(ctx, sdlApp.Handler(), srvErr, stopChan)
}
