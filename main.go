package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func getenv(name, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		value = defaultValue
	}

	return value
}

var (
	dbHost     = getenv("DB_HOST", "127.0.0.1")
	dbPort     = getenv("DB_PORT", "5432")
	dbUser     = getenv("DB_USER", "dbuser")
	dbPassword = getenv("DB_PASSWORD", "dbuser")
	dbName     = getenv("DB_NAME", "dbuser")
)

type handlerError struct {
	Error   error  `json:"-"`
	Message string `json:"error"`
	Code    int    `json:"code"`
}

// httpHandler rest service http layer handler
type httpHandler func(http.ResponseWriter, *http.Request) *handlerError

func (fn httpHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := fn(w, r); err != nil {
		log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL, err.Error)
		http.Error(w, err.Message, err.Code)
	}
}

func helloworld(rw http.ResponseWriter, request *http.Request) *handlerError {
	rw.Write([]byte("Hello world!"))
	return nil
}

type apiError struct {
	Error   error  `json:"-"`
	Message string `json:"error"`
	Code    int    `json:"code"`
}

// apiHandler DB handler mixin
type apiHandler struct {
	DB      *sql.DB
	Handler func(w http.ResponseWriter, r *http.Request, db *sql.DB) *apiError
}

func (api apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("api", "golang-pg-api/1.0")
	w.Header().Add("Content-Type", "application/json; charset=utf8")

	// if handler return an &apiError
	err := api.Handler(w, r, api.DB)
	if err != nil {
		// http log
		log.Printf("%s %s %s %s", r.RemoteAddr, r.Method, r.URL, err.Error)

		// response proper http status code
		w.WriteHeader(err.Code)

		// response JSON
		resp := json.NewEncoder(w)
		errJSON := resp.Encode(err)
		if errJSON != nil {
			log.Println("Encode JSON for error response was failed.")

			return
		}

		return
	}

	// the success case has been already taken care of in the api.Handler
	log.Printf("%s %s %s", r.RemoteAddr, r.Method, r.URL)
}

func apiStringHelper(code int, payload string, w http.ResponseWriter) *apiError {

	w.WriteHeader(code)
	_, err := w.Write([]byte(payload))

	if err != nil {
		return &apiError{
			err,
			"internal server error",
			http.StatusInternalServerError,
		}
	}

	return nil
}

// pingHandler handle '/' request
func pingHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) *apiError {

	err := db.Ping()
	if err != nil {
		return &apiError{
			err,
			"internal server error",
			http.StatusInternalServerError,
		}
	}

	log.Println("success ping database")
	return apiStringHelper(
		200,
		"{\"status\": \"success\"}",
		w)
}

// pingHandler handle '/' request
func dummyQueryHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) *apiError {

	err := db.Ping()
	if err != nil {
		return &apiError{
			err,
			"db ping error",
			http.StatusInternalServerError,
		}
	}

	row := db.QueryRow("select (1 + 4) * 20")
	if row == nil {
		return &apiError{
			err,
			"sql query error",
			http.StatusInternalServerError,
		}
	}

	var result int
	err = row.Scan(&result)
	if err != nil {
		return &apiError{
			err,
			"sql query error",
			http.StatusInternalServerError,
		}
	}

	log.Println("sql success")
	return apiStringHelper(
		200,
		"{\"result\": \""+strconv.Itoa(result)+"\"}",
		w)
}

func main() {
	dbinfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.Handle("/helloworld", httpHandler(helloworld))
	http.Handle("/ping", apiHandler{db, pingHandler})
	http.Handle("/dummyQuery", apiHandler{db, dummyQueryHandler})

	PORT := 3000
	log.Printf("Listening on :%d", PORT)
	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(PORT), nil))
}
