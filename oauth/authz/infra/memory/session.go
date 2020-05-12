package memory

import (
	"context"

	"github.com/tomocy/go-cookbook/oauth/authz"
	"github.com/tomocy/go-cookbook/oauth/authz/infra/rand"
)

func NewSessionRepo() *SessionRepo {
	return &SessionRepo{
		sessions: make(map[authz.SessionID]authz.Session),
	}
}

type SessionRepo struct {
	sessions map[authz.SessionID]authz.Session
}

func (r SessionRepo) NextID(context.Context) (authz.SessionID, error) {
	return authz.SessionID(rand.GenerateString(35)), nil
}

func (r SessionRepo) Find(_ context.Context, id authz.SessionID) (authz.Session, bool, error) {
	for _, stored := range r.sessions {
		if stored.ID() == id {
			return stored, true, nil
		}
	}

	return authz.Session{}, false, nil
}

func (r SessionRepo) FindByUserID(_ context.Context, userID authz.UserID) (authz.Session, bool, error) {
	for _, stored := range r.sessions {
		if stored.UserID() == userID {
			return stored, true, nil
		}
	}

	return authz.Session{}, false, nil
}

func (r *SessionRepo) Save(_ context.Context, sess authz.Session) error {
	r.sessions[sess.ID()] = sess
	return nil
}

func (r *SessionRepo) Delete(_ context.Context, sess authz.Session) error {
	delete(r.sessions, sess.ID())
	return nil
}
