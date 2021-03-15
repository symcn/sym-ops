package types

// ContextKey type
type ContextKey int

// Context key types(built-in)
const (
	ContextKeyStepStop ContextKey = iota
	ContextKeyNeedRequeue
	ContextKeyRequeueAfter
	ContextKeyAppsetStatus
	ContextKeyAdvdeploymentOwnerRes
	ContextKeyAdvdeploymentAggreStatus
	ContextKeyAdvdeploymentGenerationEqual
	ContextKeyEnd
)
