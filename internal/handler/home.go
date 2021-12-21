package handler

import (
	"database/sql"
	"forum/internal/env"
	"forum/internal/handler/getpost"
	"forum/internal/session"
	"forum/internal/tpl"
	"net/http"
)

type Post struct {
	ID           int
	Username     string
	Title        string
	Body         string
	CreationDate string
	Tags         []string
	LikeCount    int
	DislikeCount int
}

type HomePage struct {
	UserInfo session.User
	AllPosts []Post
	AllTags  []string
}

func Home(env *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.Error(w, "Page not found", 404)
			return
		}
		// Every time the user goes to home page it checks if he is logged in
		if _, err := session.Check(env.DB, w, r); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		posts, err := allPosts(env.DB)
		if err != nil { // If err is nil, we know we got all the posts
			http.Error(w, err.Error(), 500)
			return
		}

		tags, err := GetAllTags(env.DB) // function is in createpost.go (line 167)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		homePage := HomePage{
			UserInfo: session.UserInfo, // We need UserInfo for "base.html" template
			AllPosts: posts,
			AllTags:  tags,
		}

		tpl.RenderTemplates(w, "home.html", homePage, "./templates/base.html", "./templates/searchbar.html", "./templates/home.html")

	}
}

func allPosts(db *sql.DB) ([]Post, error) {

	rows, err := db.Query("SELECT * FROM posts")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var posts []Post
	for rows.Next() {
		var post Post
		var userid int

		if err := rows.Scan(&post.ID, &userid, &post.Title, &post.Body, &post.CreationDate); err != nil {
			return posts, err
		}

		tags, err := getpost.Tags(db, post.ID)
		if err != nil {
			return posts, err
		}

		username, err := GetUsername(db, userid)
		if err != nil {

			return posts, err
		}

		count, err := getpost.LikesDislike(db, post.ID)
		if err != nil {
			return posts, err
		}

		post.LikeCount = count.Likes
		post.DislikeCount = count.Dislikes
		post.Username = username
		post.Tags = tags

		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return posts, err
	}

	return posts, nil

}

func GetUsername(db *sql.DB, userid int) (string, error) {
	var username string
	if err := db.QueryRow("SELECT username FROM users WHERE id = ?", userid).Scan(&username); err != nil {
		return "", err
	}

	return username, nil
}
