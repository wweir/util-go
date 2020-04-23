package util

import (
	"errors"
	"strconv"
	"strings"
)

// FirstErr return the first err of err and wrapped error
func FirstErr(err error, wrappedErrs ...interface{}) error {
	if err != nil {
		return err
	}

	for i := range wrappedErrs {
		if wrapped, ok := wrappedErrs[i].(interface{ Error() error }); ok {
			return wrapped.Error()
		}
		if wrapped, ok := wrappedErrs[i].(interface{ Err() error }); ok {
			return wrapped.Err()
		}
	}
	return nil
}

// FirstErrs return the first err of a group of errors
func FirstErrs(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// MultiErr is a wrap of multi error
type MultiErr struct {
	Errs []error
}

func (m *MultiErr) Error() string {
	if len(m.Errs) == 1 {
		return m.Errs[0].Error()
	}

	summary := map[string]int{}
	for _, err := range m.Errs {
		summary[err.Error()]++
	}

	b := strings.Builder{}
	b.WriteString(strconv.Itoa(len(m.Errs)))
	b.WriteString(" errors in ")
	b.WriteString(strconv.Itoa(len(summary)))
	b.WriteString(" kinds, they are: [")

	for msg, count := range summary {
		if count != 1 {
			b.WriteString("(")
			b.WriteString(strconv.Itoa(count))
			b.WriteString("*) ")
		}

		b.WriteString(msg)
		b.WriteString(", ")
	}

	return strings.TrimSuffix(b.String(), ", ") + "]"
}

// MergeErr merge multi errors into one
func MergeErr(errs ...error) error {
	me := &MultiErr{Errs: make([]error, 0, len(errs))}
	for _, err := range errs {
		if err != nil {
			me.Errs = append(me.Errs, err)
		}
	}

	if len(me.Errs) == 0 {
		return nil
	}
	return me
}

// As return the error type same as target
func (m *MultiErr) As(target interface{}) error {
	for _, err := range m.Errs {
		if errors.As(err, target) {
			return err
		}
	}
	return nil
}
