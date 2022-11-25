package middleware

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go-postgres/models"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"

	_ "github.com/lib/pq"
)

// response format
type response struct {
	ID      int64  `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

// create connection with postgres db
func createConnection() *sql.DB {

	// collect values form env
	connStr := "postgres://" + os.Getenv("CLUSTER_USERNAME") + ":" + os.Getenv("CLUSTER_PASSWORD") + "@" + os.Getenv("CLUSTER_HOST") + "/" + os.Getenv("CLUSTER_DATABASE") + "?sslmode=disable"

	// Open the connection
	db, err := sql.Open("postgres", connStr)

	if err != nil {
		panic(err)
	}

	// check the connection
	err = db.Ping()

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected!")
	// return the connection
	return db
}

// CreateUser create a user in the postgres db
func CreateUser(w http.ResponseWriter, r *http.Request) {

	// create an empty user of type models.User
	var user models.User

	// decode the json request to user
	err := json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}

	// call insert user function and pass the user
	insertID := insertUser(user)

	// format a response object
	res := response{
		ID:      insertID,
		Message: "User created successfully",
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

// GetUser will return a single user by its id
func GetUser(w http.ResponseWriter, r *http.Request) {
	// get the userid from the request params, key is "id"
	params := mux.Vars(r)

	// convert the id type from string to int
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	// call the getUser function with user id to retrieve a single user
	user, err := getUser(int64(id))

	if err != nil {
		log.Fatalf("Unable to get user. %v", err)
	}

	// send the response
	json.NewEncoder(w).Encode(user)
}

// GetAllUser will return all the users
func GetAllUser(w http.ResponseWriter, r *http.Request) {

	// get all the users in the db
	users, err := getAllUsers()

	if err != nil {
		log.Fatalf("Unable to get all user. %v", err)
	}

	// send all the users as response
	json.NewEncoder(w).Encode(users)
}

// UpdateUser update user's detail in the postgres db
func UpdateUser(w http.ResponseWriter, r *http.Request) {

	// get the userid from the request params, key is "id"
	params := mux.Vars(r)

	// convert the id type from string to int
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	// create an empty user of type models.User
	var user models.User

	// decode the json request to user
	err = json.NewDecoder(r.Body).Decode(&user)

	if err != nil {
		log.Fatalf("Unable to decode the request body.  %v", err)
	}

	// call update user to update the user
	updatedRows := updateUser(int64(id), user)

	// format the message string
	msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", updatedRows)

	// format the response message
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

// DeleteUser delete user's detail in the postgres db
func DeleteUser(w http.ResponseWriter, r *http.Request) {

	// get the userid from the request params, key is "id"
	params := mux.Vars(r)

	// convert the id in string to int
	id, err := strconv.Atoi(params["id"])

	if err != nil {
		log.Fatalf("Unable to convert the string into int.  %v", err)
	}

	// call the deleteUser, convert the int to int64
	deletedRows := deleteUser(int64(id))

	// format the message string
	msg := fmt.Sprintf("User updated successfully. Total rows/record affected %v", deletedRows)

	// format the reponse message
	res := response{
		ID:      int64(id),
		Message: msg,
	}

	// send the response
	json.NewEncoder(w).Encode(res)
}

//------------------------- handler functions ----------------
// insert one user in the DB
func insertUser(user models.User) int64 {

	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create the insert sql query
	// returning userid will return the id of the inserted user
	sqlStatement := `INSERT INTO users (name, location, age) VALUES ($1, $2, $3) RETURNING userid`

	// the inserted id will store in this id
	var id int64

	// execute the sql statement
	// Scan function will save the insert id in the id
	err := db.QueryRow(sqlStatement, user.Name, user.Location, user.Age).Scan(&id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	fmt.Printf("Inserted a single record %v", id)

	// return the inserted id
	return id
}

// get one user from the DB by its userid
func getUser(id int64) (models.User, error) {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create a user of models.User type
	var user models.User

	// create the select sql query
	sqlStatement := `SELECT * FROM users WHERE userid=$1`

	// execute the sql statement
	row := db.QueryRow(sqlStatement, id)

	// unmarshal the row object to user
	err := row.Scan(&user.ID, &user.Name, &user.Age, &user.Location)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return user, nil
	case nil:
		return user, nil
	default:
		log.Fatalf("Unable to scan the row. %v", err)
	}

	// return empty user on error
	return user, err
}

// get one user from the DB by its userid
func getAllUsers() ([]models.User, error) {
	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	var users []models.User

	// create the select sql query
	sqlStatement := `SELECT * FROM users`

	// execute the sql statement
	rows, err := db.Query(sqlStatement)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// close the statement
	defer rows.Close()

	// iterate over the rows
	for rows.Next() {
		var user models.User

		// unmarshal the row object to user
		err = rows.Scan(&user.ID, &user.Name, &user.Age, &user.Location)

		if err != nil {
			log.Fatalf("Unable to scan the row. %v", err)
		}

		// append the user in the users slice
		users = append(users, user)

	}

	// return empty user on error
	return users, err
}

// update user in the DB
func updateUser(id int64, user models.User) int64 {

	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create the update sql query
	sqlStatement := `UPDATE users SET name=$2, location=$3, age=$4 WHERE userid=$1`

	// execute the sql statement
	res, err := db.Exec(sqlStatement, id, user.Name, user.Location, user.Age)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// check how many rows affected
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}

// delete user in the DB
func deleteUser(id int64) int64 {

	// create the postgres db connection
	db := createConnection()

	// close the db connection
	defer db.Close()

	// create the delete sql query
	sqlStatement := `DELETE FROM users WHERE userid=$1`

	// execute the sql statement
	res, err := db.Exec(sqlStatement, id)

	if err != nil {
		log.Fatalf("Unable to execute the query. %v", err)
	}

	// check how many rows affected
	rowsAffected, err := res.RowsAffected()

	if err != nil {
		log.Fatalf("Error while checking the affected rows. %v", err)
	}

	fmt.Printf("Total rows/record affected %v", rowsAffected)

	return rowsAffected
}
