package handler

import (
	"encoding/json"
	"fmt"
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

	router.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		var jsonresp []byte
		res, err := dbConn.GetUsers()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": err.Error()}
			jsonresp, _ = json.Marshal(resp)
		} else {
			w.WriteHeader(http.StatusOK)
			jsonresp, _ = json.Marshal(res)
		}
		w.Write(jsonresp)
	}).Methods(http.MethodGet)

	router.HandleFunc("/user/{uuid}", func(w http.ResponseWriter, r *http.Request) {
		var jsonresp []byte
		res, err := dbConn.GetUserById(mux.Vars(r)["uuid"])
		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": err.Error()}
			jsonresp, _ = json.Marshal(resp)
		} else {
			w.WriteHeader(http.StatusOK)
			jsonresp, _ = json.Marshal(res)
		}
		w.Write(jsonresp)
	}).Methods(http.MethodGet)

	router.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		var user db.User
		var jsonresp []byte
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
		}
		res, err := dbConn.CreateUser(&user)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": err.Error()}
			jsonresp, _ = json.Marshal(resp)
		} else {
			w.WriteHeader(http.StatusOK)
			resp := map[string]db.UserInfo{"created_user": res}
			jsonresp, _ = json.Marshal(resp)
		}
		w.Write(jsonresp)
	}).Methods(http.MethodPost)

	router.HandleFunc("/addition", func(w http.ResponseWriter, r *http.Request) {
		var cash db.Credition
		var jsonresp []byte
		err := json.NewDecoder(r.Body).Decode(&cash)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = dbConn.AddMoney(&cash)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": err.Error()}
			jsonresp, _ = json.Marshal(resp)

		} else {
			w.WriteHeader(http.StatusOK)
			resp := map[string]string{"message": fmt.Sprintf("add %d rub %d pen", cash.Rubles, cash.Pennies)}
			jsonresp, _ = json.Marshal(resp)
		}
		w.Write(jsonresp)
	}).Methods(http.MethodPatch)

	router.HandleFunc("/transfer", func(w http.ResponseWriter, r *http.Request) {
		var transfer db.Transfer
		var jsonresp []byte
		err := json.NewDecoder(r.Body).Decode(&transfer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = dbConn.TransferMoney(&transfer)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": err.Error()}
			jsonresp, _ = json.Marshal(resp)
		} else {
			w.WriteHeader(http.StatusOK)
			resp := map[string]string{"message": fmt.Sprintf("add %d rub %d pen transfered from the balance", transfer.Rubles, transfer.Pennies)}
			jsonresp, _ = json.Marshal(resp)
		}
		w.Write(jsonresp)
	}).Methods(http.MethodPatch)

	router.HandleFunc("/buy", func(w http.ResponseWriter, r *http.Request) {
		var buyer db.Buyer
		var jsonresp []byte
		err := json.NewDecoder(r.Body).Decode(&buyer)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		err = db.NewDbConnection().BuyService(&buyer)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			resp := map[string]string{"error": err.Error()}
			jsonresp, _ = json.Marshal(resp)
		} else {
			w.WriteHeader(http.StatusOK)
			resp := map[string]string{"message": fmt.Sprintf("successful purchase %v", buyer.ServicesUid)}
			jsonresp, _ = json.Marshal(resp)
		}
		w.Write(jsonresp)
	}).Methods(http.MethodPatch)

	http.Handle("/", router)
	log.Fatal(http.ListenAndServe(":"+os.Getenv("API_PORT"), nil))
}
