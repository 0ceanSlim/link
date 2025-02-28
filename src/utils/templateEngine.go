package utils

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type PageData struct {
	Title         string
	Theme         string
	UserPublicKey string
	PublicKey     string
	DisplayName   string
	Picture       string
	About         string
	Relays        RelayList
	SuccessRelays []string
	FailedRelays  []string
	DonationTags  [][]string
	IsOwnProfile  bool // ✅ Added field to check if logged-in user is viewing their own profile
	Npub          string
}

// fileExists checks if a file exists in the static directory
func fileExists(assetType, filename string) bool {
	path := filepath.Join("web/static/img", assetType, filename) // Adjusted path
	_, err := os.Stat(path)
	return err == nil
}

// Define the base directories for views and templates
const (
	viewsDir     = "web/views/"
	templatesDir = "web/views/templates/"
)

// Define the common layout templates filenames
var templateFiles = []string{
	"layout.html",
	"header.html",
	"footer.html",
}

// Initialize the common templates with full paths
var layout = PrependDir(templatesDir, templateFiles)

var loginLayout = PrependDir(templatesDir, []string{"login-layout.html", "footer.html"})

func splitString(input, sep string, index int) string {
	parts := strings.Split(input, sep)
	if index >= 0 && index < len(parts) {
		return parts[index]
	}
	return ""
}

func RenderTemplate(w http.ResponseWriter, data PageData, view string, useLoginLayout bool) {
	viewTemplate := filepath.Join(viewsDir, view)
	componentPattern := filepath.Join(viewsDir, "components", "*.html")
	componentTemplates, err := filepath.Glob(componentPattern)
	if err != nil {
		http.Error(w, "Error loading component templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var templates []string
	if useLoginLayout {
		templates = append(loginLayout, viewTemplate)
	} else {
		templates = append(layout, viewTemplate)
	}
	templates = append(templates, componentTemplates...)

	tmpl, err := template.New("").Funcs(template.FuncMap{
		"formatTimestamp":   formatTimestamp,
		"splitString":       splitString,
		"fileExists":      fileExists, // ✅ Added fileExists function
	}).ParseFiles(templates...)

	if err != nil {
		http.Error(w, "Error parsing templates: "+err.Error(), http.StatusInternalServerError)
		return
	}

	layoutName := "layout"
	if useLoginLayout {
		layoutName = "login-layout"
	}

	err = tmpl.ExecuteTemplate(w, layoutName, data)
	if err != nil {
		log.Printf("❌ Error executing template: %v", err)
	}
}
