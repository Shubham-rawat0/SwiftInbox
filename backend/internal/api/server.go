package internal

import (
	"context"
	"net/http"
)

type apiServer struct{
	server      *http.Server
}


func NewApiServer(addr string, handler http.Handler)*apiServer{
	return &apiServer{
		server:&http.Server{
			Addr:addr,
			Handler:handler,},
	}
}

func (s *apiServer) ListenAndServe()error{
	return s.server.ListenAndServe()
}

func (s *apiServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}