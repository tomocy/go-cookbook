package ouath

import "context"

type AccessTokenIntrospector struct {
	clientID string
}

func (i AccessTokenIntrospector) Introspect(ctx context.Context, tok string) (bool, error) {

}
