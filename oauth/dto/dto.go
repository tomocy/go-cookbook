package dto

type Introspection struct {
	Active   bool   `json:"active"`
	Username string `json:"username"`
}

type AccessToken struct {
	AccessToken string `json:"access_token"`
}
