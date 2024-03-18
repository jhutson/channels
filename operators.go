package channels

import (
	"sync"
)

// Map applies the given function to each element in the input channel
// and returns a new channel containing the results.
func Map[A any, B any](ch <-chan A, f func(A) B) <-chan B {
	out := make(chan B)
	go func() {
		defer close(out)
		for value := range ch {
			out <- f(value)
		}
	}()
	return out
}

// MapUntil applies the given function to each element in the input channel
// and returns a new channel containing the results.
// It is cancelled if the supplied done channel is closed before the operation has completed.
func MapUntil[A any, B any](done <-chan struct{}, ch <-chan A, f func(A) B) <-chan B {
	out := make(chan B)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case value, ok := <-ch:
				if !ok {
					return
				}
				out <- f(value)
			}
		}
	}()
	return out
}

// Flatten takes a channel of channels and merges it into a single channel.
func Flatten[A any](ch <-chan <-chan A) <-chan A {
	out := make(chan A)
	go func() {
		defer close(out)
		var wait sync.WaitGroup

		for inner := range ch {
			wait.Add(1)
			go func(ch2 <-chan A) {
				defer wait.Done()

				for value := range inner {
					out <- value
				}
			}(inner)

		}

		wait.Wait()
	}()
	return out
}

// FlattenUntil takes a channel of channels and merges it into a single channel.
// It is cancelled if the supplied done channel is closed before the operation has completed.
func FlattenUntil[A any](done <-chan struct{}, ch <-chan <-chan A) <-chan A {
	out := make(chan A)
	go func() {
		defer close(out)
		var wait sync.WaitGroup

	loop:
		for {
			select {
			case <-done:
				break loop
			case inner, ok := <-ch:
				if !ok {
					break loop
				}
				wait.Add(1)
				go func(ch2 <-chan A) {
					defer wait.Done()

					for {
						select {
						case <-done:
							return
						case value, ok := <-ch2:
							if !ok {
								return
							}
							out <- value
						}
					}
				}(inner)
			}
		}

		wait.Wait()
	}()
	return out
}

// Bind maps each element of the input channel to a new channel,
// then flattens the result into a single channel.
func Bind[A any, B any](ch <-chan A, f func(A) <-chan B) <-chan B {
	// Can be written as Flatten(Map(ch, f)).
	// Combined the two operators here to reduce the number of channels used.
	out := make(chan B)
	go func() {
		defer close(out)
		for value := range ch {
			inner := f(value)
			for innerVal := range inner {
				out <- innerVal
			}
		}
	}()
	return out
}

// BindUntil maps each element of the input channel to a new channel,
// then flattens the result into a single channel.
// It is cancelled if the supplied done channel is closed before the operation has completed.
func BindUntil[A any, B any](done <-chan struct{}, ch <-chan A, f func(A) <-chan B) <-chan B {
	// Can be written as Flatten(Map(ch, f)).
	// Combined the two operators here to reduce the number of channels used.
	out := make(chan B)
	go func() {
		defer close(out)
		var wait sync.WaitGroup

	loop:
		for {
			select {
			case <-done:
				break loop
			case value, ok := <-ch:
				if !ok {
					break loop
				}
				inner := f(value)

				wait.Add(1)
				go func(ch2 <-chan B) {
					defer wait.Done()

					for {
						select {
						case <-done:
							return
						case value, ok := <-ch2:
							if !ok {
								return
							}
							out <- value
						}
					}
				}(inner)
			}
		}

		wait.Wait()
	}()
	return out
}

// TODO: Take(n)
// TODO: TakeUntil(ch)?

// Just creates a channel that produces a single element.
func Just[A any](value A) <-chan A {
	out := make(chan A)
	go func() {
		defer close(out)
		out <- value
	}()
	return out
}

// Empty creates a closed channel with no elements.
func Empty[A any]() <-chan A {
	out := make(chan A)
	close(out)
	return out
}
