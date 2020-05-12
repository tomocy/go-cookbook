package usecase

import (
	"context"
	"fmt"

	"github.com/tomocy/go-cookbook/oauth/authz"
)

func NewCreateUser(userRepo authz.UserRepo, sessRepo authz.SessionRepo) CreateUser {
	return CreateUser{
		userRepo: userRepo,
		sessRepo: sessRepo,
	}
}

type CreateUser struct {
	userRepo authz.UserRepo
	sessRepo authz.SessionRepo
}

func (u CreateUser) Do(email, pass string) (authz.User, authz.Session, error) {
	ctx := context.TODO()

	_, found, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to find user by email: %w", err)
	}
	if found {
		return authz.User{}, authz.Session{}, fmt.Errorf("duplicated email address")
	}

	id, err := u.userRepo.NextID(ctx)
	if err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to generate user id: %w", err)
	}
	hashed, err := authz.HashPassword(pass)
	if err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to hash password: %w", err)
	}
	user, err := authz.NewUser(id, email, hashed)
	if err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to create user: %w", err)
	}
	if err := u.userRepo.Save(ctx, user); err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to save user: %w", err)
	}

	sessID, err := u.sessRepo.NextID(ctx)
	if err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to generate session id: %w", err)
	}
	sess, err := authz.NewSession(sessID, user.ID())
	if err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to create session: %w", err)
	}
	if err := u.sessRepo.Save(ctx, sess); err != nil {
		return authz.User{}, authz.Session{}, fmt.Errorf("failed to save session: %w", err)
	}

	return user, sess, nil
}

func NewAuthenticateUser(userRepo authz.UserRepo, sessRepo authz.SessionRepo) AuthenticateUser {
	return AuthenticateUser{
		userRepo: userRepo,
		sessRepo: sessRepo,
	}
}

type AuthenticateUser struct {
	userRepo authz.UserRepo
	sessRepo authz.SessionRepo
}

func (u AuthenticateUser) Do(email, pass string) (authz.User, authz.Session, bool, error) {
	ctx := context.TODO()

	user, found, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return authz.User{}, authz.Session{}, false, err
	}
	if !found {
		return authz.User{}, authz.Session{}, false, nil
	}
	if !user.Password().IsSame(pass) {
		return authz.User{}, authz.Session{}, false, nil
	}

	sessID, err := u.sessRepo.NextID(ctx)
	if err != nil {
		return authz.User{}, authz.Session{}, false, fmt.Errorf("failed to generate session id: %w", err)
	}
	sess, err := authz.NewSession(sessID, user.ID())
	if err != nil {
		return authz.User{}, authz.Session{}, false, fmt.Errorf("failed to create session: %w", err)
	}
	if err := u.sessRepo.Save(ctx, sess); err != nil {
		return authz.User{}, authz.Session{}, false, fmt.Errorf("faild to save session: %w", err)
	}

	return user, sess, true, nil
}

func NewCheckUserAuthenticated(repo authz.SessionRepo) CheckUserAuthenticated {
	return CheckUserAuthenticated{
		repo: repo,
	}
}

type CheckUserAuthenticated struct {
	repo authz.SessionRepo
}

func (u CheckUserAuthenticated) Do(userID authz.UserID) (authz.Session, bool, error) {
	ctx := context.TODO()

	sess, found, err := u.repo.FindByUserID(ctx, userID)
	if err != nil {
		return authz.Session{}, false, fmt.Errorf("failed to find session by user id: %w", err)
	}

	return sess, found, nil
}
