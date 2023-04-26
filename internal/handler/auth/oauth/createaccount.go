package oauth

import (
	"forum/internal/env"
	"forum/internal/handler/query"
	"forum/internal/hash"
	"forum/internal/session"
	"math/rand"
	"net/http"
)

// Display error message on createusername.html IF user entered existing username
var ErrorMessage string
var UserRegistered bool

func CreateAccount(env *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// We only accepot POST method to this address
		if r.Method != "POST" {
			http.Error(w, "Only POST request allowed", 400)
			return
		}

		db := env.DB
		username := r.FormValue("username")

		// Check if user entered username already exists
		// If it exists, redirect user back and display an error message
		if query.RowExists("SELECT username FROM users WHERE username = ?", username, db) {
			ErrorMessage = "This username is taken!"
			http.Redirect(w, r, "/createusername", 302)
			return
		}

		// Generate a random password
		password, err := generatePassword()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Add newly created user to database
		res, err := db.Exec("INSERT INTO users (email, username, password) VALUES (?, ?, ?)", UserEmail, username, password)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Takes the just added users ID in the database
		userid, err := res.LastInsertId()
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		// Create a session
		session.Create(int(userid), w, r, db)

		// After successful authentication reset ErrorMessage and UserEmail
		// ErrorMessage is reset because there is no error to display anymore
		// UserEmail makes sure the page is only accessible during authentication
		ErrorMessage = ""
		UserEmail = ""

		// Tracks if user completed registration and is redirected to home page
		// where we want to display a message that the registration was succesful
		UserRegistered = true

		http.Redirect(w, r, "/", 302)
		return

	}
}

// Takes random lettes from the variable LETTERS and then hashes it
func generatePassword() (string, error) {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	randomLetterBytes := make([]rune, 8)
	for i := range randomLetterBytes {
		randomLetterBytes[i] = letters[rand.Intn(len(letters))]
	}

	password, err := hash.Password(string(randomLetterBytes))
	if err != nil {
		return password, err
	}

	return password, nil
}
