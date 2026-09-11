package cleanup

import (
	"context"
	"log"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
)

func CleanUpExpired(q *postgres.Queries) (int32,error){
	count,err:=q.DeleteExpiredMailbox(context.Background(),time.Now())
	if err!=nil{
		log.Println("Cleanup error:",err.Error())
	}
	if count==0{
		log.Println("nothing to delete")
	}else{
		log.Printf("deleted %d mailboxes",count)
	}
	return count,err
}