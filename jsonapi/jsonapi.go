package jsonapi

import (
	"bytes"
	"cards-api/questiondb"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

func setJsonHeader(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

func fromJson[T any](body io.Reader, target T) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(body)
	json.Unmarshal(buf.Bytes(), &target)
}

func returnJson[T any](w http.ResponseWriter, withData func() (T, error)) {
	setJsonHeader(w)

	data, serverErr := withData()

	if serverErr != nil {
		w.WriteHeader(500)
		serverErrJson, err := json.Marshal(&serverErr)
		if err != nil {
			log.Print(err)
			return
		}
		w.Write(serverErrJson) // send the error back to client
		return
	}

	dataJson, err := json.Marshal(&data)
	if err != nil {
		log.Print(err)
		w.WriteHeader(500)
		return
	}

	w.Write(dataJson)
}

func returnErr(w http.ResponseWriter, err error, code int) {
	returnJson(w, func() (interface{}, error) {
		errorMessage := struct {
			Err string
		}{
			Err: err.Error(),
		}
		w.WriteHeader(code)
		return errorMessage, nil
	})
}

func CreateQuestion(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "POST" {
			return
		}

		entry := questiondb.TodoEntry{}
		fromJson(req.Body, &entry)

		if err := questiondb.CreateQuestion(db, entry.Description); err != nil {
			returnErr(w, err, 400)
		}

		returnJson(w, func() (interface{}, error) {
			return questiondb.GetAllQuestions(db)
		})
	})
}

func UpdateQuestion(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "PUT" {
			return
		}

		entry := questiondb.TodoEntry{}
		fromJson(req.Body, &entry)
		fmt.Println("===", entry)

		if err := questiondb.UpdateQuestion(db, int(entry.Id), entry.Done); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON UpdateQuestion: %v\n", entry.Description)
			return questiondb.GetAllQuestions(db)
		})
	})
}

func DeleteQuestion(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "DELETE" {
			return
		}

		entry := questiondb.TodoEntry{}
		fromJson(req.Body, &entry)

		if err := questiondb.DeleteQuestion(db, int(entry.Id)); err != nil {
			returnErr(w, err, 400)
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Printf("JSON DeleteQuestion: %v\n", entry.Id)
			return questiondb.GetAllQuestions(db)
		})
	})
}

func GetAllQuestions(db *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			return
		}

		returnJson(w, func() (interface{}, error) {
			log.Println("JSON GetAllQuestions")
			return questiondb.GetAllQuestions(db)
		})
	})
}

// bind = ip address that the server is running on
// ip will default to localhost if not provided (http://127.0.0.1:8080)
func Serve(db *sql.DB, bind string) {
	http.Handle("/todo/create", CreateQuestion(db))
	http.Handle("/todo/update", UpdateQuestion(db))
	http.Handle("/todo/delete", DeleteQuestion(db))
	http.Handle("/todo/getall", GetAllQuestions(db))
	log.Printf(`
	==================================

	JSON API server listening on %v

	==================================`, bind)
	err := http.ListenAndServe(bind, nil)
	if err != nil {
		// fatalf terminates the server is there is an error
		log.Fatalf("JSON server error: %v", err)
	}
}
