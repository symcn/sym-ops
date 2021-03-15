package context

import (
	"context"

	"github.com/symcn/sym-ops/pkg/types"
)

type wrapperCtx struct {
	context.Context

	builtin [types.ContextKeyEnd]interface{}
}

func (c *wrapperCtx) Value(key interface{}) interface{} {
	if contextKey, ok := key.(types.ContextKey); ok {
		return c.builtin[contextKey]
	}
	return c.Context.Value(key)
}

// WithValue add the given key-value pair into the existed value context
func WithValue(parent context.Context, key types.ContextKey, value interface{}) context.Context {
	if v, ok := parent.(*wrapperCtx); ok {
		v.builtin[key] = value
		return v
	}

	v := &wrapperCtx{Context: parent}
	v.builtin[key] = value
	return v
}

// GetValue returns result with key
func GetValue(ctx context.Context, key types.ContextKey) interface{} {
	if v, ok := ctx.(*wrapperCtx); ok {
		return v.builtin[key]
	}
	return ctx.Value(key)
}

// GetValueString returns string result with key
// if type is not string or not exist, return ""
func GetValueString(ctx context.Context, key types.ContextKey) string {
	var val interface{}
	if v, ok := ctx.(*wrapperCtx); ok {
		val = v.builtin[key]
	} else {
		val = ctx.Value(key)
	}

	if val == nil {
		return ""
	}
	if result, ok := val.(string); ok {
		return result
	}
	// means value type is not string
	return ""
}

// GetValueBool returns bool result with key
// if type is not bool or not exist, return false
func GetValueBool(ctx context.Context, key types.ContextKey) bool {
	var val interface{}
	if v, ok := ctx.(*wrapperCtx); ok {
		val = v.builtin[key]
	} else {
		val = ctx.Value(key)
	}

	if val == nil {
		return false
	}
	if b, ok := val.(bool); ok {
		return b
	}
	// means value type is not bool
	return false
}

// GetValueInt64 returns int64 result with key
// if type is not int64 or not exist, return 0
func GetValueInt64(ctx context.Context, key types.ContextKey) int64 {
	var val interface{}
	if v, ok := ctx.(*wrapperCtx); ok {
		val = v.builtin[key]
	} else {
		val = ctx.Value(key)
	}

	if val == nil {
		return 0
	}
	if result, ok := val.(int64); ok {
		return result
	}
	// means value type is not int64
	return 0
}
