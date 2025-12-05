//go:build !wasm
// +build !wasm

// This file is excluded from WASM builds.
// It contains gRPC client code that requires net/http packages
// which are not supported by TinyGo's WASM target.

package services

import (
	"context"
	"errors"
	"fmt"
	"log"

	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

const APP_ID = "sdl"

var ErrNoSuchEntity = errors.New("entity not found")

type ClientMgr struct {
	svcAddr           string
	canvasSvcClient   v1s.CanvasServiceClient
	systemsSvcClient  v1s.SystemsServiceClient
}

func NewClientMgr(svc_addr string) *ClientMgr {
	if svc_addr == "" {
		panic("Service Address is nil")
	}
	log.Println("Client Mgr Svc Addr: ", svc_addr)
	return &ClientMgr{svcAddr: svc_addr}
}

func (c *ClientMgr) Address() string {
	return c.svcAddr
}

func (c *ClientMgr) ClientContext(ctx context.Context, loggedInUserId string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return metadata.AppendToOutgoingContext(context.Background(), "LoggedInUserId", loggedInUserId)
}

func (c *ClientMgr) GetCanvasSvcClient() v1s.CanvasServiceClient {
	if c.canvasSvcClient == nil {
		conn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("cannot connect with server %v", err))
		}
		c.canvasSvcClient = v1s.NewCanvasServiceClient(conn)
	}
	return c.canvasSvcClient
}

func (c *ClientMgr) GetSystemsSvcClient() v1s.SystemsServiceClient {
	if c.systemsSvcClient == nil {
		conn, err := grpc.NewClient(c.svcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(fmt.Sprintf("cannot connect with server %v", err))
		}
		c.systemsSvcClient = v1s.NewSystemsServiceClient(conn)
	}
	return c.systemsSvcClient
}
