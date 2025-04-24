package web

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	v1 "github.com/panyam/leetcoach/gen/go/leetcoach/v1"
	svc "github.com/panyam/leetcoach/services"
	oa "github.com/panyam/oneauth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type LCApi struct {
	mux            *http.ServeMux
	AuthMiddleware *oa.Middleware
	ClientMgr      *svc.ClientMgr
}

func (n *LCApi) Handler() http.Handler {
	return n.mux
}

func NewLCApi(grpcAddr string, middleware *oa.Middleware, clients *svc.ClientMgr) *LCApi {
	out := LCApi{
		AuthMiddleware: middleware,
		ClientMgr:      clients,
		mux:            http.NewServeMux(),
	}
	gwmux, err := out.createSvcMux(grpcAddr)
	if err != nil {
		log.Println("error creating grpc mux: ", err)
		panic(err)
	}
	out.mux.Handle("/v1/", gwmux)
	return &out
}

func (web *LCApi) createSvcMux(grpc_addr string) (*runtime.ServeMux, error) {
	svcMux := runtime.NewServeMux(
		runtime.WithMetadata(func(ctx context.Context, request *http.Request) metadata.MD {
			//
			// Step 2 - Extend the context
			//
			metadata.AppendToOutgoingContext(ctx)

			loggedInUserId := web.AuthMiddleware.GetLoggedInUserId(request)

			md := metadata.Pairs()
			username, _, ok := request.BasicAuth()
			if ok { // TODO: Enable if turning on basic auth
				// TODO - Validate password if doing basic auth
				// send 401 if password is invalid
				md.Append("LoggedInUserId", username)
			} else if loggedInUserId != "" {
				log.Println("LoggedInUserId: ", loggedInUserId)
				md.Append("LoggedInUserId", loggedInUserId)
			}
			return md
		}),
		runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, writer http.ResponseWriter, request *http.Request, err error) {
			//creating a new HTTTPStatusError with a custom status, and passing error
			log.Println("Error in request: ", err)
			newError := runtime.HTTPStatusError{
				HTTPStatus: 400,
				Err:        err,
			}
			// using default handler to do the rest of heavy lifting of marshaling error and adding headers
			runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, writer, request, &newError)
		}),
	)

	// TODO - Secure credentials for etc
	opts := []grpc.DialOption{grpc.WithInsecure()}
	ctx := context.Background()
	err := v1.RegisterDesignServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
	if err != nil {
		log.Fatal("Unable to register collections service: ", err)
		panic(err)
	}
	err = v1.RegisterContentServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
	if err != nil {
		log.Fatal("Unable to register content service: ", err)
		panic(err)
	}
	err = v1.RegisterTagServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
	if err != nil {
		log.Fatal("Unable to register tags service: ", err)
		panic(err)
	}
	if os.Getenv("LEETCOACH_ENV") == "dev" {
		// Register the Admin endpoint too
		/*
			err = v1.RegisterAdminServiceHandlerFromEndpoint(ctx, svcMux, grpc_addr, opts)
			if err != nil {
				log.Fatal("Unable to register admin service: ", err)
				panic(err)
			}
		*/
	}
	return svcMux, err
}
