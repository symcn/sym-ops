package context

import (
	"context"
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
