package main

import "os"

func DefaultServiceAddress() string {
	port := os.Getenv("LEETCOACH_GRPC_PORT")
	if port != "" {
		return port
	}
	return ":9095"
}

func DefaultGatewayAddress() string {
	gateway_addr := os.Getenv("LEETCOACH_WEB_PORT")
	if gateway_addr != "" {
		return gateway_addr
	}
	return ":8080"
}
