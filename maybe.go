package channels

// Maybe represents either the value of a successful operation or the error for a failed operation.
type Maybe[A any] struct {
	value A
	error error
}

func MaybeValue[A any](value A) Maybe[A] {
	return Maybe[A]{
		value: value,
	}
}

func MaybeError[A any](err error) Maybe[A] {
	return Maybe[A]{
		error: err,
	}
}

// Result returns the underlying value and error stored in the Maybe container.
// If the Maybe container is in a successful state, it returns the stored value and a nil error.
// If the Maybe container is in a failure state, it returns the stored value (which could be the zero value of type A)
// and the error that occurred.
//
// Returns:
//
//	A - The underlying value stored in the Maybe container.
//	error - The error associated with the Maybe container. It is nil if the operation was successful.
//
// Example:
//
//	m := Maybe[int]{value: 42, error: nil}
//	value, err := m.Result()
//	if err != nil {
//	    fmt.Println("Error:", err)
//	} else {
//	    fmt.Println("Value:", value)
//	}
//	// Output:
//	// Value: 42
//
//	m := Maybe[int]{value: 0, error: errors.New("value not found")}
//	value, err := m.Result()
//	if err != nil {
//	    fmt.Println("Error:", err)
//	} else {
//	    fmt.Println("Value:", value)
//	}
//	// Output:
//	// Error: value not found
func (m Maybe[A]) Result() (A, error) {
	return m.value, m.error
}
