package app

func NewUser(id UserID) (User, error) {
	var u User
	if err := u.setID(id); err != nil {
		return User{}, err
	}

	return u, nil
}

type User struct {
	id UserID
}

func (u User) ID() UserID {
	return u.id
}

func (u *User) setID(id UserID) error {
	if id == "" {
		return ErrInvalidArg("id should not be empty")
	}

	u.id = id

	return nil
}

type UserID string

type Provider struct {
	name string
	tok  string
}

type errInput interface {
	error
	ErrInput()
}

type ErrInvalidArg string

func (e ErrInvalidArg) Error() string {
	return string(e)
}

func (ErrInvalidArg) ErrInput() {}
