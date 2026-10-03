package contextx

import "context"

func SetString(ctx context.Context, key string, value string) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetString(ctx context.Context, key string) string {
	v, err := ctx.Value(key).(string)
	if !err {
		return ""
	}
	return v
}

func SetInt(ctx context.Context, key string, value int) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetInt(ctx context.Context, key string) int {
	v, err := ctx.Value(key).(int)
	if !err {
		return 0
	}
	return v
}

func SetBool(ctx context.Context, key string, value bool) context.Context {
	return context.WithValue(ctx, key, value)
}

func GetBool(ctx context.Context, key string) bool {
	v, err := ctx.Value(key).(bool)
	if !err {
		return false
	}
	return v
}
