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

			actualCount := 0
			ch := IntRange(bufferSize, elementCount)

			for value := range Map(ch, double) {
				assert.Equal(t, double(actualCount), value)
				actualCount++
			}

			assert.Equal(t, elementCount, actualCount)
		})
	}
}

func TestMapUntil(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("no cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			actualCount := 0
			ch := IntRange(bufferSize, elementCount)

			for value := range MapUntil(context.Background().Done(), ch, double) {
				assert.Equal(t, double(actualCount), value)
				actualCount++
			}

			assert.Equal(t, elementCount, actualCount)
		})

		t.Run(fmt.Sprintf("with cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			actualCount := 0
			cancelAtCount := rand.IntN(elementCount/4) + 1
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := IntRange(bufferSize, elementCount)

			for value := range MapUntil(ctx.Done(), ch, double) {
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

			values := make(map[int]int)
			ch := IntRange(bufferSize, elementCount)

			testChannel := Flatten(Map(ch, func(x int) <-chan int {
				return Repeat(0, x+1, x+1)
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

func TestFlattenUntil(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("no cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			values := make(map[int]int)
			ch := IntRange(bufferSize, elementCount)

			testChannel := FlattenUntil(context.Background().Done(), Map(ch, func(x int) <-chan int {
				return Repeat(0, x+1, x+1)
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

			actualCount := 0
			cancelAtCount := rand.IntN(elementCount/4) + 1
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := IntRange(bufferSize, elementCount)

			testChannel := FlattenUntil(ctx.Done(), Map(ch, Just))

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

			values := make(map[int]int)
			ch := IntRange(bufferSize, elementCount)

			testChannel := Bind(ch, func(x int) <-chan int {
				return Repeat(0, x+1, x+1)
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

func TestBindUntil(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("no cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {

			values := make(map[int]int)
			ch := IntRange(bufferSize, elementCount)

			testChannel := BindUntil(context.Background().Done(), ch, func(x int) <-chan int {
				return Repeat(0, x+1, x+1)
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

			actualCount := 0
			cancelAtCount := rand.IntN(elementCount/4) + 1
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := IntRange(bufferSize, elementCount)

			testChannel := BindUntil(ctx.Done(), ch, Just)

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

			actualCount := 0
			ch := IntRange(bufferSize, elementCount)

			zs := Bind(
				Bind(ch,
					func(x int) <-chan string {
						return Just(fmt.Sprint(x))
					}),
				func(y string) <-chan int {
					return Just(len(y))
				})

			for z := range zs {
				actualCount += z
			}

			assert.Equal(t, elementCount*2-10, actualCount)
		})
	}
}

func TestTake(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("producer with buffer size %d", bufferSize), func(t *testing.T) {
			actualCount := 0
			ch := Infinite(bufferSize, 1)

			for x := range Take(ch, elementCount) {
				actualCount += x
			}

			assert.Equal(t, elementCount, actualCount)
		})
	}
}

func TestTakeUntil(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("no cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {
			actualCount := 0
			ch := Repeat(bufferSize, 1, elementCount)

			for x := range TakeUntil(context.Background().Done(), ch) {
				actualCount += x
			}

			assert.Equal(t, elementCount, actualCount)
		})

		t.Run(fmt.Sprintf("with cancellation, producer with buffer size %d", bufferSize), func(t *testing.T) {
			actualCount := 0
			cancelAtCount := rand.IntN(elementCount-20) + 20
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			ch := Infinite(bufferSize, 1)

			for x := range TakeUntil(ctx.Done(), ch) {
				actualCount += x
				if actualCount == cancelAtCount {
					cancel()
				}
			}

			assert.GreaterOrEqual(t, actualCount, cancelAtCount)
			assert.Less(t, actualCount, elementCount)
		})
	}
}

func IntRange(channelSize int, count int) <-chan int {
	ch := make(chan int, channelSize)

	go func() {
		defer close(ch)

		for i := 0; i < count; i++ {
			ch <- i
		}
	}()

	return ch
}

func Infinite(channelSize int, value int) <-chan int {
	ch := make(chan int, channelSize)

	go func() {
		for {
			ch <- value
		}
	}()

	return ch
}

func Repeat(channelSize int, value int, count int) <-chan int {
	ch := make(chan int, channelSize)

	go func() {
		defer close(ch)

		for i := 0; i < count; i++ {
			ch <- value
		}
	}()

	return ch
}
