package router

import "net/http"

type Router struct{
	handler  	http.Handler
}

func NewServerMux()*Router{

	mux:=http.NewServeMux()

	handler:=cors(mux)

	return &Router{
		handler: handler, 
	}
}

func (h *Router) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
) {
	h.handler.ServeHTTP(w, r)
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Origin", "127.0.0.1:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}