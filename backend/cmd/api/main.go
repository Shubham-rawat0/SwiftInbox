package api

import (
	internal "github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api"
	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/api/router"
)

func main(){
	handler:=router.NewServerMux()
	server:=internal.NewApiServer(":3000",handler)
	
	err:=server.ListenAndServe()
	if err!=nil{
		panic(err)
	}
}
