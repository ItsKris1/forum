### Forum authentication

Give user an option to login & register through Google or Github.

### Implementation

One handler which will handle responses either from Google or Github.<br>
Exchanging recieved authorization code from response for an access token.<br>
Send a request with an access token to an API endpoint.<br>
Finish user authorization with data from the request.

### How to run the project

Navigate to project root directory **/forum-authentication**<br>
Run ```go run cmd/forum/main.go```<br>
Go to **http://localhost:8000**

### Technologies
Frontend
- JQuery Select2
- Bootstrap 5
- HTML & CSS

Backend
- Golang
- SQLite

### Links
Google OAuth - https://developers.google.com/identity/protocols/oauth2/web-server<br>
Github OAuth - https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps
### Author
Kristofer Kangro(itskris)