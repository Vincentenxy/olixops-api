package ctx

import "context"

type UserMetaKey struct{}

type UserMeta struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
	JWT      string `json:"jwt"`
}

func WithUserMeta(ctx context.Context, userMeta UserMeta) context.Context {
	return context.WithValue(ctx, UserMetaKey{}, userMeta)
}

func GetUserMeta(ctx context.Context) *UserMeta {
	val := ctx.Value(UserMetaKey{})
	if val == nil {
		return nil
	}

	m, ok := val.(UserMeta)
	if !ok {
		return nil
	}

	return &m
}
