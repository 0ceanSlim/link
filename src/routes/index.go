package routes

import (
	"link/src/handlers"
	"link/src/types"
	"link/src/utils"
	"log"
	"net/http"
)

func Index(w http.ResponseWriter, r *http.Request) {
	session, _ := handlers.User.Get(r, "session-name")

	publicKey, ok := session.Values["publicKey"].(string)
	if !ok || publicKey == "" {
		log.Println("⚠️ No publicKey found in session. Redirecting to login.")
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	log.Printf("✅ Session data found: publicKey=%s", publicKey)

	displayName, _ := session.Values["displayName"].(string)
	picture, _ := session.Values["picture"].(string)
	about, _ := session.Values["about"].(string)

	log.Printf("✅ DisplayName=%s, Picture=%s, About=%s", displayName, picture, about)

	relays, ok := session.Values["relays"].(utils.RelayList)
	if !ok {
		log.Println("⚠️ No relay list found in session")
		relays = utils.RelayList{}
	}

	relayURLs := relays.ToStringSlice()

	// Fetch last 10 kind 1 notes
	notes, err := utils.FetchLast10Kind1Notes(publicKey, relayURLs)
	if err != nil {
		log.Printf("⚠️ Failed to fetch notes: %v", err)
		notes = []types.NostrEvent{}
	}

	// ✅ Get donation addresses from session
	donationTags, ok := session.Values["donationTags"].([][]string)
	if !ok {
		log.Println("⚠️ No donation tags found in session")
		donationTags = [][]string{} // Default to empty
	} else {
		log.Printf("✅ DonationTags: %+v", donationTags)
	}

	// Construct page data
	data := utils.PageData{
		Title:        "Dashboard",
		DisplayName:  displayName,
		Picture:      picture,
		PublicKey:    publicKey,
		About:        about,
		Relays:       relays,
		Notes:        notes,
		DonationTags: donationTags, // ✅ Pass donation addresses to template
	}

	utils.RenderTemplate(w, data, "index.html", false)
}
