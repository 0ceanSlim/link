package routes

import (
	"link/src/utils"
	"net/http"
)

func Login(w http.ResponseWriter, r *http.Request) {
	data := utils.PageData{
		Title: "Login",
	}
	utils.RenderTemplate(w, data, "login.html", true)
}
