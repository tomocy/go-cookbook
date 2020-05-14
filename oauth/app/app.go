package app

type User struct {
	id UserID
}

type UserID string

type Provider struct {
	name   string
	userID UserID
	tok    string
}
