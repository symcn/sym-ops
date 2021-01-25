package types

// StepNext step should need next
type StepNext bool

// Continue continue do next step
// Stop stop now
const (
	Continue StepNext = true
	Stop     StepNext = false
)

// NeedRequeue need requeue
type NeedRequeue bool

// Requeue requeue last
// Done mark this step don't need requeue
const (
	Requeue NeedRequeue = true
	Done    NeedRequeue = false
)

// Step step interface
type Step interface {
	Do() (isContinue StepNext, isRedo NeedRequeue, err error)
}
