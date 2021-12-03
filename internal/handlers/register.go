package handlers

import (
	"fmt"
	"html/template"
	"net/http"
)

func Register(w http.ResponseWriter, r *http.Request) {

	tpl, _ := template.ParseFiles("./templates/register.html")
	err := tpl.Execute(w, RegValidation)
	if err != nil {
		fmt.Println(err)

		http.Error(w, "500 Internal Server error", 500)
		return
	}
}
