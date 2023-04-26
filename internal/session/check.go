package session

import (
	"database/sql"
	"forum/internal/handler/query"
	"forum/internal/handler/structs"
	"net/http"
	"time"
)

var UserInfo structs.User

func Check(db *sql.DB, w http.ResponseWriter, r *http.Request) (bool, error) {
	cookie, err := r.Cookie("session")

	if err != nil {
		if err == http.ErrNoCookie { // If there isnt an existing cookie, there isnt an ongoing session
			UserInfo.ID = 0
			return false, nil
		}
		UserInfo.ID = 0
		return false, err

		// If cookie exists, get the UserID of the cookie and update the UserInfo.ID(which tracks, what is the logged in user ID)
	} else {
		// Check if that cookie belongs to user
		row := db.QueryRow("SELECT userid FROM sessions WHERE uuid = ?", cookie.Value)

		if err := row.Scan(&UserInfo.ID); err != nil {
			// If it wont find who the cookie belongs to - it deletes it
			if err == sql.ErrNoRows {
				UserInfo.ID = 0

				cookie.Expires = time.Unix(0, 0)
				http.SetCookie(w, cookie)
				return false, nil // Return nil because the error is handled
			}

			UserInfo.ID = 0
			return false, err
		}

		// Get the logged in user's Username
		username, err := query.GetUsername(db, UserInfo.ID)
		if err != nil {
			return false, err
		}
		UserInfo.Username = username

		return true, err
	}
}
