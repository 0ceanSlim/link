package routes

import (
	"encoding/json"
	"link/src/cache"
	"link/src/handlers"
	"link/src/types"
	"link/src/utils"
	"log"
	"net/http"
	"strings"
)

// ProfileHandler serves both the index ("/") and profile pages ("/p/<npub>")
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	session, _ := handlers.User.Get(r, "session-name")
	userPublicKey, _ := session.Values["UserPublicKey"].(string)

	// If user is not logged in and is at "/", redirect to login
	if path == "/" {
		if userPublicKey == "" {
			log.Println("⚠️ No UserPublicKey found in session. Redirecting to login.")
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Convert hex public key to npub and redirect to /p/<npub>
		npub, err := utils.EncodeNpub(userPublicKey)
		if err != nil {
			log.Printf("❌ Failed to encode UserPublicKey to npub: %v", err)
			http.Error(w, "Failed to encode public key", http.StatusInternalServerError)
			return
		}

		log.Printf("🔄 Redirecting logged-in user to /%s", npub)
		http.Redirect(w, r, "/"+npub, http.StatusSeeOther)
		return
	}

	// If the path starts with "/", treat it as a profile view
	if strings.HasPrefix(path, "/") {
		npub := strings.TrimPrefix(path, "/")

		// Convert npub to hex pubkey
		pubKey, err := utils.DecodeNpub(npub)
		if err != nil {
			http.Error(w, "Invalid npub", http.StatusBadRequest)
			return
		}

		// Determine if the logged-in user is viewing their own profile
		isOwnProfile := userPublicKey == pubKey

		var metadata types.UserMetadata
		var donationTags [][]string
		var relays utils.RelayList

		if isOwnProfile {
			// Fetch data from cache for own profile
			cachedData, found := cache.GetUserData(pubKey)
			if !found {
				http.Error(w, "No cached data found", http.StatusNotFound)
				return
			}

			var rawEvent types.NostrEvent
			if err := json.Unmarshal([]byte(cachedData.Metadata), &rawEvent); err != nil {
				log.Printf("❌ Failed to parse cached raw event JSON: %v", err)
				http.Error(w, "Failed to parse raw event", http.StatusInternalServerError)
				return
			}

			if err := json.Unmarshal([]byte(cachedData.Relays), &relays); err != nil {
				log.Printf("❌ Failed to parse cached relays JSON: %v", err)
				http.Error(w, "Failed to parse relays", http.StatusInternalServerError)
				return
			}

			donationTags = extractDonationTags(rawEvent.Tags)
		} else {
			// Fetch relays for other users
			relaysPtr, err := utils.FetchUserRelays(pubKey, []string{
				"wss://purplepag.es", "wss://relay.damus.io", "wss://nos.lol",
				"wss://relay.primal.net", "wss://relay.nostr.band", "wss://offchain.pub",
			})
			if err != nil {
				log.Printf("Failed to fetch user relays: %v", err)
			}
			if relaysPtr != nil {
				relays = *relaysPtr
			}

			// Get user metadata (raw NostrEvent)
			userEvent, err := utils.FetchUserMetadata(pubKey, relays.ToStringSlice())
			if err != nil {
				http.Error(w, "Failed to fetch user metadata", http.StatusInternalServerError)
				return
			}
			if userEvent == nil {
				http.Error(w, "User metadata not found", http.StatusNotFound)
				return
			}

			donationTags = extractDonationTags(userEvent.Tags)
		}

		// Prepare page data
		data := utils.PageData{
			Title:         metadata.DisplayName + "'s Profile",
			DisplayName:   metadata.DisplayName,
			Picture:       metadata.Picture,
			PublicKey:     pubKey,
			UserPublicKey: userPublicKey,
			About:         metadata.About,
			Relays:        relays,
			DonationTags:  donationTags,
			IsOwnProfile:  isOwnProfile,
		}

		// Render profile template
		utils.RenderTemplate(w, data, "index.html", false)
		return
	}

	// If the route does not match known paths, return 404
	http.NotFound(w, r)
}

// extractDonationTags filters donation tags from the event tags
func extractDonationTags(tags [][]string) [][]string {
	var donationTags [][]string
	for _, tag := range tags {
		if len(tag) > 0 && tag[0] == "w" {
			donationTags = append(donationTags, tag)
		}
	}
	return donationTags
}
