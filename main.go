package main

import (
	"link/src/handlers"
	"link/src/routes"
	"link/src/utils"

	"fmt"
	"net/http"
)

func main() {
	// Load Configurations
	cfg, err := utils.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	mux := http.NewServeMux()
	// Login / Logout
	mux.HandleFunc("/login", routes.Login) // Login route
	mux.HandleFunc("/init-user", handlers.InitUser)
	mux.HandleFunc("/logout", handlers.LogoutHandler) // Logout process
	mux.HandleFunc("/encode-key", handlers.EncodeKeyHandler)
	mux.HandleFunc("/get-session", handlers.GetSessionHandler)
	mux.HandleFunc("/get-cache", handlers.GetCacheHandler)
	mux.HandleFunc("/fetch_current_kind0", handlers.FetchCurrentKind0FromCache)

	// Initialize Routes
	//mux.HandleFunc("/", routes.Index)
	mux.HandleFunc("/settings", routes.Settings)
	mux.HandleFunc("/", routes.ProfileHandler)

	// Function Handlers
	mux.HandleFunc("/save_donation_addresses", handlers.SaveDonationAddresses)
	mux.HandleFunc("/fetch_user_metadata", handlers.FetchUserMetadataHandler)

	// Serve Web Files
	// Serve specific files from the root directory
	mux.HandleFunc("/favicon.svg", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/favicon.svg")
	})
	// Serve static files from the /web/static directory at /static/
	staticDir := "web/static"
	mux.Handle("/static/", http.StripPrefix("/static", http.FileServer(http.Dir(staticDir))))

	// Serve CSS files from the /web/style directory at /style/
	styleDir := "web/style"
	mux.Handle("/style/", http.StripPrefix("/style", http.FileServer(http.Dir(styleDir))))

	fmt.Printf("Server is running on http://localhost:%d\n", cfg.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), mux)
}
