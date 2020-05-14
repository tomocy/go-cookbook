package authz

import (
	"context"
	"errors"
	"net/url"
	"time"
)

type ClientRepo interface {
	NewCreds(context.Context) (ClientID, string, error)
	NewAccessToken(context.Context) (string, error)
	NewCode(context.Context) (string, error)
	Find(context.Context, ClientID) (Client, bool, error)
	Save(context.Context, Client) error
	Delete(context.Context, Client) error
}

func NewClient(id ClientID, secret string, redirectURI string) (Client, error) {
	c := Client{
		codes: make(map[string]Code),
		toks:  make(map[string]AccessToken),
	}

	if err := c.setID(id); err != nil {
		return Client{}, err
	}
	if err := c.setSecret(secret); err != nil {
		return Client{}, err
	}
	if err := c.setRedirectURL(redirectURI); err != nil {
		return Client{}, err
	}

	return c, nil
}

type Client struct {
	id          ClientID
	secret      string
	redirectURI url.URL
	codes       map[string]Code
	toks        map[string]AccessToken
}

func (c Client) ID() ClientID {
	return c.id
}

func (c *Client) setID(id ClientID) error {
	if id == "" {
		return ErrInvalidArg("id should not be empty")
	}

	c.id = id

	return nil
}

func (c Client) Secret() string {
	return c.secret
}

func (c *Client) setSecret(secret string) error {
	if secret == "" {
		return ErrInvalidArg("secret should not be empty")
	}

	c.secret = secret

	return nil
}

func (c *Client) setRedirectURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil {
		return ErrInvalidArg(err.Error())
	}

	c.redirectURI = *parsed

	return nil
}

func (c Client) Code(code string) (Code, bool) {
	stored, ok := c.codes[code]
	return stored, ok
}

func (c *Client) StoreCode(code Code) error {
	return c.setCode(code)
}

func (c Client) RedirectURIWithAuthzCode(code Code) (string, error) {
	if code.IsExpired() {
		return "", ErrInvalidArg("code should not be expired")
	}

	queries := url.Values{
		"code": []string{code.Code()},
	}
	uri := c.redirectURI
	uri.RawQuery = queries.Encode()

	return uri.String(), nil
}

func (c *Client) setCode(code Code) error {
	if code.IsExpired() {
		return ErrInvalidArg("code should not be expired")
	}

	c.codes[code.code] = code

	return nil
}

func (c *Client) DeleteCode(code string) {
	delete(c.codes, code)
}

func (c Client) AccessToken(tok string) (AccessToken, bool) {
	stored, ok := c.toks[tok]
	return stored, ok
}

func (c *Client) StoreAccessToken(tok AccessToken) error {
	return c.setAccessToken(tok)
}

func (c *Client) setAccessToken(tok AccessToken) error {
	if tok.IsExpired() {
		return ErrInvalidArg("token should not be expired")
	}

	c.toks[tok.tok] = tok

	return nil
}

func (c *Client) DeleteAccessToken(tok string) {
	delete(c.toks, tok)
}

type ClientID string

func NewCode(code string) (Code, error) {
	var c Code
	if err := c.setCode(code); err != nil {
		return Code{}, err
	}
	if err := c.setCreatedAt(time.Now()); err != nil {
		return Code{}, err
	}

	return c, nil
}

type Code struct {
	code      string
	createdAt time.Time
}

func (c Code) Code() string {
	return c.code
}

func (c *Code) setCode(code string) error {
	if code == "" {
		return ErrInvalidArg("code should not be empty")
	}

	c.code = code

	return nil
}

func (c *Code) setCreatedAt(at time.Time) error {
	if at.IsZero() {
		return ErrInvalidArg("creation time should not be zero value")
	}

	c.createdAt = at

	return nil
}

func (c Code) IsExpired() bool {
	now := time.Now()
	diff := now.Sub(c.createdAt)

	return diff < 0 || 10*time.Minute < diff
}

func NewAccessToken(tok string) (AccessToken, error) {
	var t AccessToken
	if err := t.setToken(tok); err != nil {
		return AccessToken{}, err
	}
	if err := t.setCreatedAt(time.Now()); err != nil {
		return AccessToken{}, err
	}

	return t, nil
}

type AccessToken struct {
	tok       string
	createdAt time.Time
}

func (t AccessToken) Token() string {
	return t.tok
}

func (t *AccessToken) setToken(tok string) error {
	if tok == "" {
		return ErrInvalidArg("token should not be empty")
	}

	t.tok = tok

	return nil
}

func (t *AccessToken) setCreatedAt(at time.Time) error {
	if at.IsZero() {
		return ErrInvalidArg("creation time should not be zero value")
	}

	t.createdAt = at

	return nil
}

func (t AccessToken) IsExpired() bool {
	now := time.Now()
	diff := now.Sub(t.createdAt)
	return diff < 0 || 1*time.Hour < diff
}

type Introspection struct {
	Active   bool   `json:"active"`
	Username string `json:"username"`
}

func IsErrInternal(err error) bool {
	_, ok := err.(errInternal)
	return ok
}

type errInternal interface {
	error
	ErrInternal()
}

func IsErrInput(err error) bool {
	if err == nil {
		return false
	}
	if _, ok := err.(errInput); ok {
		return true
	}

	return IsErrInput(errors.Unwrap(err))
}

type errInput interface {
	error
	ErrInput()
}

func IsErrInvalidArg(err error) bool {
	_, ok := err.(ErrInvalidArg)
	return ok
}

type ErrInvalidArg string

func (e ErrInvalidArg) Error() string {
	return string(e)
}

func (ErrInvalidArg) ErrInput() {}
