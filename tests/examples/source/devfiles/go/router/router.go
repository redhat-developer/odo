package router

import (
	"net/http"

	"go-postgres/middleware"

	"github.com/gorilla/mux"
)

// Router is exported and used in main.go
func Router() *mux.Router {

	router := mux.NewRouter()

	router.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("pong"))
	})

	router.HandleFunc("/api/user/{id}", middleware.GetUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/user", middleware.GetAllUser).Methods("GET", "OPTIONS")
	router.HandleFunc("/api/newuser", middleware.CreateUser).Methods("POST", "OPTIONS")
	router.HandleFunc("/api/user/{id}", middleware.UpdateUser).Methods("PUT", "OPTIONS")
	router.HandleFunc("/api/deleteuser/{id}", middleware.DeleteUser).Methods("DELETE", "OPTIONS")

	return router
}
