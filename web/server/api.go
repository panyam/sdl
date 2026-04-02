package server

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	gfn "github.com/panyam/goutils/fn"
	v1s "github.com/panyam/sdl/gen/go/sdl/v1/services"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type SDLApi struct {
	mux *http.ServeMux
}

func (n *SDLApi) Handler() http.Handler {
	return n.mux
}

func NewSDLApi(grpcAddr string) *SDLApi {
	out := SDLApi{
		mux: http.NewServeMux(),
	}
	gwmux, err := out.createSvcMux(grpcAddr)
	if err != nil {
		log.Println("error creating grpc mux: ", err)
		panic(err)
	}
	out.mux.Handle("/v1/", gwmux)
	log.Println("Registered gRPC-gateway at /v1/")

	return &out
}

func (web *SDLApi) createSvcMux(grpc_addr string) (*runtime.ServeMux, error) {
	svcMux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			md := metadata.Pairs()
			return md
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			s := status.Convert(err)
			httpStatus := runtime.HTTPStatusFromCode(s.Code())

			log.Printf("gRPC Gateway Error: code=%s, http_status=%d, msg=%s, details=%v\n", s.Code(), httpStatus, s.Message(), s.Details())

			body := struct {
				Error   string        `json:"error"`
				Message string        `json:"message"`
				Code    int           `json:"code"`
				Details []interface{} `json:"details,omitempty"`
			}{
				Error:   s.Code().String(),
				Message: s.Message(),
				Code:    int(s.Code()),
				Details: gfn.Map(s.Proto().Details, func(detail *anypb.Any) any {
					var msg proto.Message
					msg, err = anypb.UnmarshalNew(detail, proto.UnmarshalOptions{})
					if err != nil {
						log.Printf("Failed to unmarshal error detail: %v", err)
					}
					return msg
				}),
			}

			writer.Header().Del("Trailer")
			writer.Header().Set("Content-Type", marshaler.ContentType(body))
			writer.WriteHeader(httpStatus)
			if err := marshaler.NewEncoder(writer).Encode(body); err != nil {
				log.Printf("Failed to marshal error response: %v", err)
				runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, err)
			}
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	ctx := context.Background()

	// Register WorkspaceService gateway
	err := v1s.RegisterWorkspaceServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
	if err != nil {
		log.Fatal("Unable to register workspace service: ", err)
		return nil, err
	}

	return svcMux, nil
}
