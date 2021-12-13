package handler

import (
	"database/sql"
	"forum/internal/env"
	"forum/internal/session"
	"forum/internal/tpl"
	"net/http"
)

// "createpost.html" uses "base" template, which has a navbar what uses data from UserInfo
type CreatePostPage struct {
	UserInfo session.User
}

func CreatePost(env *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		createPostPage := CreatePostPage{
			UserInfo: session.UserInfo,
		}

		if r.Method == "POST" {

			if err := r.ParseForm(); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}

			db := env.DB // intializes db connection

			// Add data to Posts table
			if err := addPosts(db, r); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			// Get the ID of the tags used in create post
			if err := addTags(db, r); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			// Add data to PostTags table
			if err := addPostTags(db, r); err != nil {
				http.Error(w, err.Error(), 500)
				return
			}

			http.Redirect(w, r, "/", 302)
			return
		}

		tpl.RenderTemplates(w, "createpost.html", createPostPage, "./templates/createpost.html", "./templates/base.html")
	}

}

// Adds tag names to Tags table
func addTags(db *sql.DB, r *http.Request) error {

	stmt, err := db.Prepare("INSERT OR IGNORE INTO tags (name) VALUES (?)")
	if err != nil {
		return err
	}

	for _, tag := range r.Form["tags"] {
		stmt.Exec(tag)
	}

	return nil
}

/*
1. Get the ID of the tags
2. Get the ID of the post
3. Add all the used tag IDs and post ID into the PostTags table
*/
func addPostTags(db *sql.DB, r *http.Request) error {
	var tagIDs []string

	for _, tag := range r.Form["tags"] {

		var tagid string
		if err := db.QueryRow("SELECT id FROM tags WHERE name = ?", tag).Scan(&tagid); err != nil {
			return err
		}

		tagIDs = append(tagIDs, tagid)

	}

	row := db.QueryRow("SELECT postid FROM posts WHERE title = ?", r.FormValue("title"))

	var postid string
	if err := row.Scan(&postid); err != nil {
		return err
	}

	for _, id := range tagIDs {
		stmt, err := db.Prepare("INSERT INTO posttags (postid, tagid) VALUES (?, ?)")
		if err != nil {
			return err
		}

		stmt.Exec(postid, id)
	}

	return nil
}

/*
1. Get the ID of the user by using UUID from the cookie
2. Add the post title, body and ID of the user into Posts table
*/
func addPosts(db *sql.DB, r *http.Request) error {
	cookie, err := r.Cookie("session")
	if err != nil { // If there is no cookie then the session has expired
		return err
	}

	userid, err := GetUserID(db, cookie.Value) // GetUserID is from addcomment.go file
	if err != nil {
		return err
	}

	// Add new post to database
	stmt, err := db.Prepare("INSERT INTO posts (title, body, userid) VALUES (?, ?, ?)")
	if err != nil {
		return err
	}

	stmt.Exec(r.FormValue("title"), r.FormValue("body"), userid)
	return nil

}
