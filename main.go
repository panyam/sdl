package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"

	"github.com/joho/godotenv"
	"github.com/panyam/leetcoach/web"
)

var (
	// svc_addr = flag.String("svc_addr", DefaultServiceAddress(), "Address where the gRPC endpoint is running")
	gw_addr = flag.String("gw_addr", DefaultGatewayAddress(), "Address where the http grpc gateway endpoint is running")
)

func main() {

	// logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	envfile := ".env"
	if os.Getenv("LEETCOACH_ENV") == "dev" {
		envfile = ".env.dev"
		logger := slog.New(NewPrettyHandler(os.Stdout, PrettyHandlerOptions{
			SlogOpts: slog.HandlerOptions{
				Level: slog.LevelDebug,
			},
		}))
		slog.SetDefault(logger)
	}
	log.Println("loading env file: ", envfile)
	err := godotenv.Load(envfile)
	if err != nil {
		log.Fatal("Error loading .env file", envfile, err)
	}

	flag.Parse()
	app := App{Ctx: context.Background()}
	// app.AddServer(&svc.Server{Address: *svc_addr})
	app.AddServer(&web.Server{Address: *gw_addr})
	app.Start()
	app.Done(nil)
}

func DefaultGatewayAddress() string {
	gateway_addr := os.Getenv("LEETCOACH_WEB_PORT")
	if gateway_addr != "" {
		return gateway_addr
	}
	return ":8080"
}
