package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"link/src/cache"
	"link/src/handlers"
	"link/src/utils"
)

func Settings(w http.ResponseWriter, r *http.Request) {
	session, err := handlers.User.Get(r, "session-name")
	if err != nil {
		log.Printf("Error getting session: %v\n", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	UserPublicKey, ok := session.Values["UserPublicKey"].(string)
	if !ok || UserPublicKey == "" {
		log.Println("No publicKey found in session")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Fetch the relay list from the cache
	cachedData, found := cache.GetUserData(UserPublicKey)
	var relays utils.RelayList
	if found {
		if err := json.Unmarshal([]byte(cachedData.Relays), &relays); err != nil {
			log.Printf("❌ Failed to parse cached relays JSON: %v", err)
		}
	} else {
		log.Println("No relay list found in cache for Settings view")
	}

	data := utils.PageData{
		Title:     "Settings",
		UserPublicKey: UserPublicKey,
		Relays:    relays,
	}

	utils.RenderTemplate(w, data, "settings.html", false)
}
