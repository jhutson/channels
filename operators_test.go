package channels

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
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
			ch := IntRange(bufferSize, 0, elementCount)

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
			ch := IntRange(bufferSize, 0, elementCount)

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

			ch := IntRange(bufferSize, 0, elementCount)

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
			ch := IntRange(bufferSize, 0, elementCount)

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
			ch := IntRange(bufferSize, 0, elementCount)

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

			ch := IntRange(bufferSize, 0, elementCount)

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
			ch := IntRange(bufferSize, 0, elementCount)

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
			ch := IntRange(bufferSize, 0, elementCount)

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

			ch := IntRange(bufferSize, 0, elementCount)

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
			ch := IntRange(bufferSize, 0, elementCount)

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

func TestAggregate(t *testing.T) {
	for _, bufferSize := range []int{0, 1, elementCount} {
		t.Run(fmt.Sprintf("producer with buffer size %d", bufferSize), func(t *testing.T) {
			ch := StringIntRange(bufferSize, 1, elementCount)

			addLength := func(x string, sum int) int {
				n, _ := strconv.Atoi(x)
				return sum + n
			}

			expectedTotal := elementCount * (elementCount + 1) / 2 //  n(n + 1) ∕ 2
			total := 0
			count := 0
			for result := range Aggregate(ch, 0, addLength) {
				total = result
				count++
			}

			assert.Equal(t, 1, count)
			assert.Equal(t, expectedTotal, total)
		})
	}
}

func IntRange(channelSize int, start int, count int) <-chan int {
	ch := make(chan int, channelSize)

	go func() {
		defer close(ch)

		endCount := start + count
		for i := start; i < endCount; i++ {
			ch <- i
		}
	}()

	return ch
}

func StringIntRange(channelSize int, start int, count int) <-chan string {
	ch := make(chan string, channelSize)

	go func() {
		defer close(ch)

		endCount := start + count
		for i := start; i < endCount; i++ {
			ch <- fmt.Sprint(i)
		}
	}()

	return ch
}

func Infinite[A any](channelSize int, value A) <-chan A {
	ch := make(chan A, channelSize)

	go func() {
		for {
			ch <- value
		}
	}()

	return ch
}

func Repeat[A any](channelSize int, value A, count int) <-chan A {
	ch := make(chan A, channelSize)

	go func() {
		defer close(ch)

		for i := 0; i < count; i++ {
			ch <- value
		}
	}()

	return ch
}
