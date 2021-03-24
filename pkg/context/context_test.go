package context

import (
	"context"
	"math"
	"testing"

	"github.com/symcn/sym-ops/pkg/types"
)

func TestWithValue(t *testing.T) {
	ctx := context.TODO()

	WithValue(ctx, types.ContextKeyAppsetStatus, "1")
	if GetValueString(ctx, types.ContextKeyAppsetStatus) == "1" {
		t.Error("value not save, not equal")
	}

	ctx = WithValue(ctx, types.ContextKeyAppsetStatus, "1")
	if GetValueString(ctx, types.ContextKeyAppsetStatus) != "1" {
		t.Error("value not save")
	}
}

func TestGetValue(t *testing.T) {
	ctx := context.TODO()
	ctx = WithValue(ctx, types.ContextKeyStepStop, "end")

	t.Run("get value", func(t *testing.T) {
		value := struct{}{}
		WithValue(ctx, types.ContextKeyStepStop, value)
		if _, ok := GetValue(ctx, types.ContextKeyStepStop).(struct{}); !ok {
			t.Errorf("save 'struct{}' but got %v", GetValue(ctx, types.ContextKeyStepStop))
		}

		if _, ok := GetValue(context.TODO(), types.ContextKeyStepStop).(struct{}); ok {
			t.Errorf("not save native context but got %v", GetValue(context.TODO(), types.ContextKeyStepStop))
		}
	})

	t.Run("get string", func(t *testing.T) {
		WithValue(ctx, types.ContextKeyStepStop, "end")
		if GetValueString(ctx, types.ContextKeyStepStop) != "end" {
			t.Errorf("save 'end' but got %s", GetValueString(ctx, types.ContextKeyStepStop))
		}

		WithValue(ctx, types.ContextKeyStepStop, nil)
		if GetValueString(ctx, types.ContextKeyStepStop) != "" {
			t.Errorf("save nil but got %s", GetValueString(ctx, types.ContextKeyStepStop))
		}

		WithValue(ctx, types.ContextKeyStepStop, true)
		if GetValueString(ctx, types.ContextKeyStepStop) != "" {
			t.Errorf("save bool but got %s", GetValueString(ctx, types.ContextKeyStepStop))
		}

		if GetValue(ctx, types.ContextKeyEnd) != nil {
			t.Errorf("get ContextKeyEnd is nil but got %s", GetValue(ctx, types.ContextKeyStepStop))
		}
	})

	t.Run("get bool", func(t *testing.T) {
		WithValue(ctx, types.ContextKeyStepStop, true)
		if !GetValueBool(ctx, types.ContextKeyStepStop) {
			t.Error("save 'true' but got false")
		}

		WithValue(ctx, types.ContextKeyStepStop, nil)
		if GetValueBool(ctx, types.ContextKeyStepStop) {
			t.Error("save nil but got true")
		}

		WithValue(ctx, types.ContextKeyStepStop, "end")
		if GetValueBool(ctx, types.ContextKeyStepStop) {
			t.Errorf("save string but got %t", GetValueBool(ctx, types.ContextKeyStepStop))
		}

		if GetValueBool(ctx, types.ContextKeyEnd) {
			t.Error("get ContextKeyEnd is nil but got true")
		}
	})

	t.Run("get int64", func(t *testing.T) {
		WithValue(ctx, types.ContextKeyStepStop, int64(100))
		if GetValueInt64(ctx, types.ContextKeyStepStop) != 100 {
			t.Errorf("save 100 but got %d", GetValueInt64(ctx, types.ContextKeyStepStop))
		}

		WithValue(ctx, types.ContextKeyStepStop, nil)
		if GetValueInt64(ctx, types.ContextKeyStepStop) != 0 {
			t.Errorf("save nil but got %d", GetValueInt64(ctx, types.ContextKeyStepStop))
		}

		WithValue(ctx, types.ContextKeyStepStop, true)
		if GetValueInt64(ctx, types.ContextKeyStepStop) != 0 {
			t.Errorf("save bool but got %d", GetValueInt64(ctx, types.ContextKeyStepStop))
		}

		if GetValueInt64(ctx, types.ContextKeyEnd) != 0 {
			t.Errorf("get ContextKeyEnd is nil but got %d", GetValueInt64(ctx, types.ContextKeyStepStop))
		}
	})

	t.Run("native value", func(t *testing.T) {
		value := 100
		WithValue(ctx, types.ContextKeyStepStop, value)

		if ctx.Value(math.MaxInt32) != nil {
			t.Errorf("get max value must is nil but got %v", ctx.Value(math.MaxInt32))
		}
		if ctx.Value(types.ContextKeyStepStop) != value {
			t.Errorf("save 100 but got %v", ctx.Value(types.ContextKeyStepStop))
		}
		if ctx.Value(nil) != nil {
			t.Errorf("get nil value must is nil but got %v", ctx.Value(nil))
		}
	})
}
