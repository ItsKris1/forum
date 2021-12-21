package handler

import (
	"database/sql"
	"forum/internal/env"
	"forum/internal/handler/auth"
	"forum/internal/handler/check"
	"forum/internal/handler/getpost"
	"forum/internal/session"
	"forum/internal/tpl"
	"net/http"
)

type UserPage struct {
	UserInfo     session.User
	LikedPosts   []Post
	CreatedPosts []Post
}

func UserDetails(env *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Only GET request allowed", 400)
			return
		}

		isLogged, err := session.Check(env.DB, w, r)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		if !isLogged {
			http.Redirect(w, r, "/login", 302)
			auth.LoginMsgs.LoginRequired = true
			return
		}
		db := env.DB
		userid := r.URL.Query().Get("id")
		if err := check.URLQuery(db, "SELECT id FROM users WHERE id = ?", userid); err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		likedPosts, err := userLikedPosts(db, userid)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), 500)
			return
		}

		createdPosts, err := userCreatedPosts(db, userid)
		if err != nil && err != sql.ErrNoRows {
			http.Error(w, err.Error(), 500)
			return
		}

		userPage := UserPage{
			UserInfo:     session.UserInfo,
			LikedPosts:   likedPosts,
			CreatedPosts: createdPosts,
		}

		tpl.RenderTemplates(w, "userdetails.html", userPage, "./templates/base.html", "./templates/userdetails.html")
	}
}

func userLikedPosts(db *sql.DB, userid string) ([]Post, error) {
	rows, err := db.Query("SELECT postid FROM postlikes WHERE userid = ? AND like = 1", userid)
	if err != nil {
		return nil, err
	}

	var likedPosts []Post

	for rows.Next() {
		var postid int
		var post Post

		if err := rows.Scan(&postid); err != nil {
			return likedPosts, err
		}

		var userid int
		if err := db.QueryRow("SELECT userid FROM posts WHERE postid = ?", postid).Scan(&userid); err != nil {
			return likedPosts, err
		}

		row := db.QueryRow("SELECT postid, title, body, creation_date FROM posts WHERE postid = ? AND userid = ?", postid, userid)
		if err := row.Scan(&post.ID, &post.Title, &post.Body, &post.CreationDate); err != nil {
			return likedPosts, err
		}

		tags, err := getpost.Tags(db, post.ID)
		if err != nil {
			return likedPosts, err
		}

		count, err := getpost.LikesDislike(db, post.ID)
		if err != nil {
			return likedPosts, err
		}

		post.LikeCount = count.Likes
		post.DislikeCount = count.Dislikes
		post.Tags = tags

		likedPosts = append(likedPosts, post)
	}

	if err := rows.Err(); err != nil {
		return likedPosts, err
	}

	return likedPosts, nil
}

func userCreatedPosts(db *sql.DB, userid string) ([]Post, error) {
	rows, err := db.Query("SELECT postid, title, body, creation_date FROM posts WHERE userid = ?", userid)
	if err != nil {
		return nil, err
	}

	var createdPosts []Post

	for rows.Next() {

		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Body, &post.CreationDate); err != nil {
			return createdPosts, err
		}

		tags, err := getpost.Tags(db, post.ID)
		if err != nil {
			return createdPosts, err
		}

		count, err := getpost.LikesDislike(db, post.ID)
		if err != nil {
			return createdPosts, err
		}

		post.LikeCount = count.Likes
		post.DislikeCount = count.Dislikes
		post.Tags = tags
		createdPosts = append(createdPosts, post)
	}

	if err := rows.Err(); err != nil {
		return createdPosts, err
	}

	return createdPosts, nil

}
