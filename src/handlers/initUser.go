package handlers

import (
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"

	"link/src/cache"
	"link/src/types"
	"link/src/utils"

	"github.com/gorilla/sessions"
)

var User = sessions.NewCookieStore([]byte("your-secret-key"))

func init() {
	gob.Register(utils.RelayList{})
	gob.Register([][]string{})
	gob.Register(types.NostrEvent{})
	gob.Register([]types.NostrEvent{})
}

func InitUser(w http.ResponseWriter, r *http.Request) {
	log.Println("🔑 InitUser called")

	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Failed to parse form: %v", err)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	publicKey := r.FormValue("publicKey")
	if publicKey == "" {
		log.Println("❌ Missing publicKey in form data")
		http.Error(w, "Missing publicKey", http.StatusBadRequest)
		return
	}
	log.Printf("✅ Received publicKey: %s", publicKey)

	initialRelays := []string{
		"wss://purplepag.es", "wss://relay.damus.io", "wss://nos.lol",
		"wss://relay.primal.net", "wss://relay.nostr.band", "wss://offchain.pub",
	}

	userRelays, err := utils.FetchUserRelays(publicKey, initialRelays)
	if err != nil {
		log.Printf("❌ Failed to fetch user relays: %v", err)
		http.Error(w, "Failed to fetch user relays", http.StatusInternalServerError)
		return
	}

	allRelays := append(userRelays.Read, userRelays.Write...)
	allRelays = append(allRelays, userRelays.Both...)
	log.Printf("✅ Fetched user relays: %+v", userRelays)

	userMetadataEvent, err := utils.FetchUserMetadata(publicKey, allRelays)
	if err != nil || userMetadataEvent == nil {
		log.Printf("❌ Failed to fetch user metadata: %v", err)
		http.Error(w, "Failed to fetch user metadata", http.StatusInternalServerError)
		return
	}

	relaysJSON, err := json.Marshal(userRelays)
	if err != nil {
		log.Printf("❌ Failed to serialize user relays: %v", err)
		http.Error(w, "Failed to process user relays", http.StatusInternalServerError)
		return
	}

	eventJSON, err := json.Marshal(userMetadataEvent)
	if err != nil {
		log.Printf("❌ Failed to serialize user metadata event: %v", err)
		http.Error(w, "Failed to process user metadata", http.StatusInternalServerError)
		return
	}

	cache.SetUserData(publicKey, string(eventJSON), string(relaysJSON))
	log.Println("✅ User data cached successfully")

	session, _ := User.Get(r, "session-name")
	session.Values["UserPublicKey"] = publicKey
	if err := session.Save(r, w); err != nil {
		log.Printf("❌ Failed to save session: %v", err)
		http.Error(w, "Failed to save session", http.StatusInternalServerError)
		return
	}

	npub, err := utils.EncodeNpub(publicKey)
	if err != nil {
		log.Printf("❌ Failed to encode publicKey to npub: %v", err)
		http.Error(w, "Failed to encode public key", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/"+npub, http.StatusSeeOther)
	log.Printf("🔄 Redirecting to /%s", npub)
}
