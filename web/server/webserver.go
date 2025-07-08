package server

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/felixge/httpsnoop"
	"github.com/panyam/sdl/web/services"
)

type WebAppServer struct {
	Address       string
	GrpcAddress   string
	AllowLocalDev bool
	CanvasService *services.CanvasService
}

// func (s *WebAppServer) Start(ctx context.Context, mux http.Handler, gw_addr string, srvErr chan error, stopChan chan bool) {
func (s *WebAppServer) Start(ctx context.Context, srvErr chan error, stopChan chan bool) error {
	app := NewWebServer(s.GrpcAddress, s.CanvasService)
	return s.StartWithHandler(ctx, app.Handler(), srvErr, stopChan)
}

func (s *WebAppServer) StartWithHandler(ctx context.Context, handler http.Handler, srvErr chan error, stopChan chan bool) error {
	log.Println("Starting http web server on: ", s.Address)
	// handler := otelhttp.NewHandler(mux, "gateway", otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string { return fmt.Sprintf("%s %s %s", operation, r.Method, r.URL.Path) }))
	handler = withLogger(handler)
	if s.AllowLocalDev {
		handler = CORS(handler)
	}
	server := &http.Server{
		Addr:        s.Address,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		Handler:     handler,
	}

	go func() {
		<-stopChan
		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatalln(err)
			panic(err)
		}
	}()
	srvErr <- server.ListenAndServe()
	return nil
}

func withLogger(handler http.Handler) http.Handler {
	// the create a handler
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// pass the handler to httpsnoop to get http status and latency
		m := httpsnoop.CaptureMetrics(handler, writer, request)
		// printing exracted data
		if false && m.Code != 200 { // turn off frequent logs
			log.Printf("http[%d]-- %s -- %s, Query: %s\n", m.Code, m.Duration, request.URL.Path, request.URL.RawQuery)
		}
	})
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// fmt.Println(r.Header)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, Origin, Cache-Control, X-Requested-With")
		//w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Methods", "PUT, DELETE")

		if r.Method == "OPTIONS" {
			w.WriteHeader(204)
			return
		}

		next.ServeHTTP(w, r)
	})
}
