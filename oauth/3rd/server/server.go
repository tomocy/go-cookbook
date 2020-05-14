package server

type User struct {
	ID        string
	Providers map[string]Provider
}

type Provider struct {
	Name  string
	Token string
}

type Owner struct {
	Name     string
	Provider string
}
