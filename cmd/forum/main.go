package main

import (
	"database/sql"
	"fmt"
	"forum/internal/env" // imports Env struct, where we store the db connection
	"forum/internal/handler"
	"forum/internal/handler/auth"
	"forum/internal/handler/auth/oauth"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "./db/storage.db")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("DB: {}", db)
	// Makes an environment for Database connection
	env := &env.Env{DB: db}

	// One crate statement for the audit
	db.Exec(`CREATE TABLE IF NOT EXISTS "postlikes" (
		"postid"	INTEGER NOT NULL,
		"userid"	INTEGER NOT NULL,
		"like"	INTEGER NOT NULL,
		FOREIGN KEY("postid") REFERENCES "posts"("postid") ON DELETE CASCADE ON UPDATE CASCADE,
		FOREIGN KEY("userid") REFERENCES "users"("id") ON DELETE CASCADE ON UPDATE CASCADE,
		PRIMARY KEY("userid","postid")
	);`)

	// Home page
	http.HandleFunc("/", handler.Home(env))

	// Adding comment and post
	http.HandleFunc("/createpost", handler.CreatePost(env))
	http.HandleFunc("/addcomment", handler.AddComment(env))

	// Posts filtering
	http.HandleFunc("/search", handler.Search(env))

	// Reacting to posts/comments
	http.HandleFunc("/react", handler.React(env))

	// Viewing information about user
	http.HandleFunc("/user", handler.UserDetails(env))

	// Viewing a post
	http.HandleFunc("/post", handler.ViewPost(env))

	// Web page registration/authentication
	http.HandleFunc("/register", auth.Register())
	http.HandleFunc("/registerauth", auth.RegisterAuth(env))
	http.HandleFunc("/login", auth.Login(env))
	http.HandleFunc("/loginauth", auth.LoginAuth(env))
	http.HandleFunc("/logout", auth.Logout(env))

	// OAuth2
	http.HandleFunc("/oauth/", oauth.AuthenticateUser(env))
	http.HandleFunc("/createaccount", oauth.CreateAccount(env))
	http.HandleFunc("/createusername", oauth.CreateUsername())

	fs := http.FileServer(http.Dir("./assets"))
	http.Handle("/assets/", http.StripPrefix("/assets/", fs))

	http.HandleFunc("/favicon.ico", ignoreFavicon)
	if err := http.ListenAndServe(":8000", nil); err != nil {
		fmt.Println("Listening port at: 8000...")
		log.Fatal(err)
	}

}

func ignoreFavicon(w http.ResponseWriter, r *http.Request) {}
