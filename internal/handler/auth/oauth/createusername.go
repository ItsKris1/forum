package oauth

import (
	"forum/internal/handler/structs"
	"forum/internal/session"
	"forum/internal/tpl"
	"net/http"
)

type CreateUsernamePage struct {
	ErrorMsg string // Message to display if user entered an existing username
	UserInfo structs.User
}

func CreateUsername() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// This page is only accessible if user is asked to provide username AFTER authentication
		if UserEmail == "" {
			http.Error(w, "Bad request", 400)
			return
		}

		pageData := CreateUsernamePage{
			UserInfo: session.UserInfo,
			ErrorMsg: ErrorMessage,
		}

		tpl.RenderTemplates(w, "createusername.html", pageData, "./templates/base.html", "./templates/auth/createusername.html")
	}
}
