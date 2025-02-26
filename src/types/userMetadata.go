package types

type UserMetadata struct {
	DisplayName string     `json:"display_name"`
	Picture     string     `json:"picture"`
	About       string     `json:"about"`
	Tags        [][]string `json:"tags"` // New field to store metadata tags
}
