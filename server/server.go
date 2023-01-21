package main

import (
	"cards-api/jsonapi"
	"cards-api/questiondb"
	"database/sql"
	"log"
	"sync"

	"github.com/alexflint/go-arg"
)

var args struct {
	// We can specify the question_db in the commandline, otherwise it is set to a default
	DbPath   string `arg:"env:QUESTION_DB`
	BindJson string `arg:"env:QUESTION_BIND_JSON"`
}

func main() {
	arg.MustParse(&args)

	if args.DbPath == "" {
		args.DbPath = "question.db" // default DB location
	}
	if args.BindJson == "" {
		args.BindJson = ":8080" // default port
	}

	log.Printf("using database '%v'\n", args.DbPath)
	db, err := sql.Open("sqlite3", args.DbPath)

	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	questiondb.TryCreate(db)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		log.Printf("starting JSON API server...\n")
		jsonapi.Serve(db, args.BindJson)
		wg.Done()
	}()

	wg.Wait()
}
