package commands

import (
	"context"
	"fmt"
	"time"

	v1 "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// getGRPCConnection returns a gRPC connection to the server
func getGRPCConnection() (*grpc.ClientConn, error) {
	serverAddr := "localhost:9090"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gRPC server at %s: %v", serverAddr, err)
	}

	return conn, nil
}

// getWorkspaceClient returns a WorkspaceService client
func getWorkspaceClient() (v1.WorkspaceServiceClient, *grpc.ClientConn, error) {
	conn, err := getGRPCConnection()
	if err != nil {
		return nil, nil, err
	}

	client := v1.NewWorkspaceServiceClient(conn)
	return client, conn, nil
}

// withWorkspaceClient handles the common pattern of getting a client, creating context, and cleaning up
func withWorkspaceClient(fn func(client v1.WorkspaceServiceClient, ctx context.Context) error) error {
	client, conn, err := getWorkspaceClient()
	if err != nil {
		return fmt.Errorf("cannot connect to SDL server: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return fn(client, ctx)
}
