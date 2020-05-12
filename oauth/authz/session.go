package authz

import "context"

type SessionRepo interface {
	NextID(context.Context) (SessionID, error)
	Find(context.Context, SessionID) (Session, bool, error)
	FindByUserID(context.Context, UserID) (Session, bool, error)
	Save(context.Context, Session) error
	Delete(context.Context, Session) error
}

func NewSession(id SessionID, userID UserID) (Session, error) {
	var s Session
	if err := s.setID(id); err != nil {
		return Session{}, err
	}
	if err := s.setUserID(userID); err != nil {
		return Session{}, err
	}

	return s, nil
}

type Session struct {
	id     SessionID
	userID UserID
}

func (s Session) ID() SessionID {
	return s.id
}

func (s *Session) setID(id SessionID) error {
	if id == "" {
		return ErrInvalidArg("id should not be empty")
	}

	s.id = id

	return nil
}

func (s Session) UserID() UserID {
	return s.userID
}

func (s *Session) setUserID(userID UserID) error {
	if userID == "" {
		return ErrInvalidArg("user id should not be empty")
	}

	s.userID = userID

	return nil
}

type SessionID string
