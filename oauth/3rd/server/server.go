package server

type Provider struct {
	Name  string
	Token string
}

type Owner struct {
	Name     string
	Provider string
}
