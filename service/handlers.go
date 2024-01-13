package service

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"postgres-test/db"

	"github.com/gorilla/mux"
)

func Run() {

	router := mux.NewRouter()
	dbConn := db.NewDbConnection()
	log.Println("Service is running")

	router.HandleFunc("/get/users", func(w http.ResponseWriter, r *http.Request) {
		res, err := dbConn.GetUsers()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		}

	}).Methods(http.MethodGet)

	router.HandleFunc("/get/user/{id}", func(w http.ResponseWriter, r *http.Request) {
		res, err := dbConn.GetUserById(mux.Vars(r)["id"])
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		}
	}).Methods(http.MethodGet)

	router.HandleFunc("/create/user", func(w http.ResponseWriter, r *http.Request) {
		var user db.User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		res, err := dbConn.CreateUser(&user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		}
	}).Methods(http.MethodPost)

	router.HandleFunc("/balance/up", func(w http.ResponseWriter, r *http.Request) {
		var cash db.Credition
		err := json.NewDecoder(r.Body).Decode(&cash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		res, err := dbConn.AddMoney(&cash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		}
	}).Methods(http.MethodPatch)

	router.HandleFunc("/reserve/up", func(w http.ResponseWriter, r *http.Request) {
		var cash db.Credition
		err := json.NewDecoder(r.Body).Decode(&cash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		res, err := dbConn.ReserveMoneyFromBalance(&cash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write(res)
		}
	}).Methods(http.MethodPatch)

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("API_PORT"), nil))
}
