package channels

import (
	"context"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
)

const elementCount = 100

func double(x int) int {
	return x * 2
}

func TestMap(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			actualCount := 0

			go PublishIntRange(ch, elementCount)

			for value := range Map(ch, double) {
				assert.Equal(t, double(actualCount), value)
				actualCount++
			}

			assert.Equal(t, elementCount, actualCount)
		})
	}
}

func TestMapC(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("no cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			actualCount := 0

			go PublishIntRange(ch, elementCount)

			for value := range MapC(context.Background().Done(), ch, double) {
				assert.Equal(t, double(actualCount), value)
				actualCount++
			}

			assert.Equal(t, elementCount, actualCount)
		})

		t.Run(fmt.Sprintf("with cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			actualCount := 0
			cancelAtCount := rand.IntN(elementCount/4) + 1
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go PublishIntRange(ch, elementCount)

			for value := range MapC(ctx.Done(), ch, double) {
				assert.Equal(t, double(actualCount), value)
				actualCount++
				if actualCount == cancelAtCount {
					cancel()
				}
			}

			assert.GreaterOrEqual(t, actualCount, cancelAtCount)
			assert.Less(t, actualCount, elementCount)
		})
	}
}

func TestFlatten(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			values := make(map[int]int)

			go PublishIntRange(ch, elementCount)

			testChannel := Flatten(Map(ch, func(x int) <-chan int {
				return Repeat(x+1, x+1)
			}))

			for value := range testChannel {
				values[value] += 1
			}

			assert.Len(t, values, elementCount)
			for i := 1; i <= elementCount; i++ {
				assert.Equal(t, i, values[i])
			}
		})
	}
}

func TestFlattenC(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("no cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			values := make(map[int]int)

			go PublishIntRange(ch, elementCount)

			testChannel := FlattenC(context.Background().Done(), Map(ch, func(x int) <-chan int {
				return Repeat(x+1, x+1)
			}))

			for value := range testChannel {
				values[value] += 1
			}

			assert.Len(t, values, elementCount)
			for i := 1; i <= elementCount; i++ {
				assert.Equal(t, i, values[i])
			}
		})

		t.Run(fmt.Sprintf("with cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			actualCount := 0
			cancelAtCount := rand.IntN(elementCount/4) + 1
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go PublishIntRange(ch, elementCount)

			testChannel := FlattenC(ctx.Done(), Map(ch, Just))

			for range testChannel {
				actualCount++
				if actualCount == cancelAtCount {
					cancel()
				}
			}

			assert.GreaterOrEqual(t, actualCount, cancelAtCount)
			assert.Less(t, actualCount, elementCount)
		})

	}
}

func TestBind(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			values := make(map[int]int)

			go PublishIntRange(ch, elementCount)

			testChannel := Bind(ch, func(x int) <-chan int {
				return Repeat(x+1, x+1)
			})

			for value := range testChannel {
				values[value] += 1
			}

			assert.Len(t, values, elementCount)
			for i := 1; i <= elementCount; i++ {
				assert.Equal(t, i, values[i])
			}
		})
	}
}

func TestBindC(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("no cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			values := make(map[int]int)

			go PublishIntRange(ch, elementCount)

			testChannel := BindC(context.Background().Done(), ch, func(x int) <-chan int {
				return Repeat(x+1, x+1)
			})

			for value := range testChannel {
				values[value] += 1
			}

			assert.Len(t, values, elementCount)
			for i := 1; i <= elementCount; i++ {
				assert.Equal(t, i, values[i])
			}
		})

		t.Run(fmt.Sprintf("with cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			ch := make(chan int, bufferSize)
			actualCount := 0
			cancelAtCount := rand.IntN(elementCount/4) + 1
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go PublishIntRange(ch, elementCount)

			testChannel := BindC(ctx.Done(), ch, Just)

			for range testChannel {
				actualCount++
				if actualCount == cancelAtCount {
					cancel()
				}
			}

			assert.GreaterOrEqual(t, actualCount, cancelAtCount)
			assert.Less(t, actualCount, elementCount)
		})

		t.Run(fmt.Sprintf("composition, producer with buffer size %d", bufferSize), func(t *testing.T) {
			ch := make(chan int, bufferSize)

			zs := Bind(
				Bind(ch,
					func(x int) <-chan string {
						return Just(fmt.Sprint(x))
					}),
				func(y string) <-chan int {
					return Just(len(y))
				})

			go PublishIntRange(ch, elementCount)

			actualCount := 0

			for z := range zs {
				actualCount += z
			}

			assert.Equal(t, elementCount*2-10, actualCount)
		})
	}
}

func PublishIntRange(ch chan<- int, count int) {
	defer close(ch)

	for i := 0; i < count; i++ {
		ch <- i
	}
}

func Repeat(value int, count int) <-chan int {
	ch := make(chan int)

	go func() {
		defer close(ch)

		for i := 0; i < count; i++ {
			ch <- value
		}
	}()

	return ch
}
