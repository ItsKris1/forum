package oauth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"forum/internal/env"
	"forum/internal/session"
	"net/http"
)

type OAuth struct {
	clientID     string
	clientSecret string
	redirectURI  string

	tokenURL string
	APIURL   string

	code string
}

var UserEmail string // Saving user email for later use

// Authentication by using OAuth
func AuthenticateUser(env *env.Env) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var authData OAuth

		// authHeader is used to authorize API requests
		// authHeader consists of a NAME and ACCESS TOKEN
		// ex: token 897safasf987
		var authHeader string

		// OAuth provider name(Google, Github) from URL
		provider := r.URL.Path[7:]
		switch provider {

		case "google":
			authData = OAuth{
				clientID:     "a",
				clientSecret: "a",
				redirectURI:  "http://localhost:8000/oauth/google",

				tokenURL: "https://oauth2.googleapis.com/token",
				APIURL:   "https://www.googleapis.com/oauth2/v2/userinfo",
			}
			authHeader = "Bearer "

		case "github":
			authData = OAuth{
				clientID:     "f",
				clientSecret: "f",
				redirectURI:  "http://localhost:8000/oauth/github",

				tokenURL: "https://github.com/login/oauth/access_token",
				APIURL:   "https://api.github.com/user/emails",
			}
			authHeader = "token "

		default:
			http.Error(w, "OAuth: Invalid redirect URL", 400)
			return
		}

		// Authorization code from URL
		authData.code = r.FormValue("code")

		// Get access token by using the data from the authData
		// authData will have the URL and URL parameters to make the request and get access token
		accessToken, err := getAccessToken(authData, provider)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		// Add accessToken to authHeader to authorize requests
		authHeader += accessToken

		// Sends a request to the API and authorizes it by setting HTTP header "Authorization" to authHeader value
		UserEmail, err = getUserEmail(authData.APIURL, authHeader, provider)
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		// Create a database connection
		db := env.DB

		// Check if the authenticated user is registered
		row := db.QueryRow("SELECT id FROM users WHERE email = ?", UserEmail)

		var userid int
		switch err := row.Scan(&userid); err {

		case nil: // If user is registered we create a session
			session.Create(userid, w, r, db)
			http.Redirect(w, r, "/", 302)
			return

		case sql.ErrNoRows: // If user is not registered
			http.Redirect(w, r, "/createusername", 302) // Navigate to page, where user can create the username and finish sign in
			return

		default: // If something unexpected happened
			http.Error(w, err.Error(), 500)
			return
		}
	}
}

// Get access token by using the data from the authData
// authData will have the URL and URL parameters to make the request and get access token
func getAccessToken(authData OAuth, provider string) (string, error) {
	type authResponse struct {
		AccessToken string `json:"access_token"`
	}

	var authResp authResponse
	var client http.Client
	var url string

	// Construct URLs based on the provider
	switch provider {
	case "google":
		url = fmt.Sprintf("https://oauth2.googleapis.com/token?client_id=%s&client_secret=%s&code=%s&grant_type=authorization_code&redirect_uri=http://localhost:8000/oauth/google", authData.clientID, authData.clientSecret, authData.code)
	case "github":
		url = fmt.Sprintf("https://github.com/login/oauth/access_token?client_id=%s&client_secret=%s&code=%s", authData.clientID, authData.clientSecret, authData.code)
	default:
		return "", errors.New("No provider!")
	}

	// Create a POST request to URL
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", err
	}

	// Give back the data as JSON
	req.Header.Set("accept", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Internet connection or client policy error")
		return "", err
	}

	err = json.NewDecoder(resp.Body).Decode(&authResp)
	if err != nil {
		fmt.Println("Error occured during decoding access token response")
		return "", err
	}
	defer resp.Body.Close()

	return authResp.AccessToken, nil
}

// Sends a request to the API and
// authorizes it by setting HTTP header "Authorization" to authHeader value
func getUserEmail(endpoint, authHeader, provider string) (string, error) {
	var email string       // Store user email here
	var client http.Client // Create client so we can modify request headers

	// Create a GET request to the endpoint
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return "", err
	}

	// Authorize the request by using the HTTP header
	req.Header.Add("Authorization", authHeader)

	// Give the data back as JSON
	req.Header.Add("accept", "application/json")

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Internet connection or client policy error")
		return "", err
	}
	defer resp.Body.Close()

	var response interface{}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", err
	}

	// If you are sure that the structure of the content of the response,
	// given its type, is always what you expect it to be, you can use a
	// quick-and-dirty type switch/assertion.
	switch v := response.(type) {
	case []interface{}:
		email = v[0].(map[string]interface{})["email"].(string)
	case map[string]interface{}:
		email = v["email"].(string)
	}



	return email, nil
}
