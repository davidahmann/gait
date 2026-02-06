package errors

import "errors"

type Category string

const (
	CategoryInvalidInput      Category = "invalid_input"
	CategoryVerification      Category = "verification_failed"
	CategoryPolicyBlocked     Category = "policy_blocked"
	CategoryApprovalRequired  Category = "approval_required"
	CategoryDependencyMissing Category = "dependency_missing"
	CategoryIOFailure         Category = "io_failure"
	CategoryStateContention   Category = "state_contention"
	CategoryNetworkTransient  Category = "network_transient"
	CategoryNetworkPermanent  Category = "network_permanent"
	CategoryInternalFailure   Category = "internal_failure"
)

type classifiedError struct {
	category  Category
	code      string
	hint      string
	retryable bool
	cause     error
}

func (e *classifiedError) Error() string {
	if e.cause == nil {
		return "unknown error"
	}
	return e.cause.Error()
}

func (e *classifiedError) Unwrap() error {
	return e.cause
}

func (e *classifiedError) Category() Category {
	return e.category
}

func (e *classifiedError) Code() string {
	return e.code
}

func (e *classifiedError) Hint() string {
	return e.hint
}

func (e *classifiedError) Retryable() bool {
	return e.retryable
}

func Wrap(cause error, category Category, code, hint string, retryable bool) error {
	if cause == nil {
		return nil
	}
	return &classifiedError{
		category:  category,
		code:      code,
		hint:      hint,
		retryable: retryable,
		cause:     cause,
	}
}

func CategoryOf(err error) Category {
	var classified *classifiedError
	if errors.As(err, &classified) {
		return classified.category
	}
	return ""
}

func CodeOf(err error) string {
	var classified *classifiedError
	if errors.As(err, &classified) {
		return classified.code
	}
	return ""
}

func HintOf(err error) string {
	var classified *classifiedError
	if errors.As(err, &classified) {
		return classified.hint
	}
	return ""
}

func RetryableOf(err error) bool {
	var classified *classifiedError
	if errors.As(err, &classified) {
		return classified.retryable
	}
	return false
}
