package internal

import "net/http"


type Server interface{
	ListenAndServe()  error
}

type apiServer struct{
	server      *http.Server
}


func NewApiServer(addr string, handler http.Handler) *apiServer{
	return &apiServer{
		server:&http.Server{
			Addr:addr,
			Handler:handler,},
	}
}

func (s *apiServer) ListenAndServe()error{
	return s.server.ListenAndServe()
}