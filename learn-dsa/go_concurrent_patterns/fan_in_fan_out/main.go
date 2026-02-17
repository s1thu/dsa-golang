package main

import (
	"fmt"
	"sync"
	"time"
)

// generator produces data and sends it to a channel
func generator(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

// fanOut: Takes one channel and spawns multiple goroutines to process data concurrently
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			fmt.Printf("Processing %d\n", n)
			time.Sleep(time.Second) // Simulate work
			out <- n * n
		}
		close(out)
	}()
	return out
}

// fanIn: Merges multiple channels into one channel
func fanIn(channels ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup

	// Start a goroutine for each input channel
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for val := range c {
				out <- val
			}
		}(ch)
	}

	// Close output channel when all inputs are done
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func main() {
	// Stage 1: Generate data
	input := generator(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

	// Stage 2: Fan-Out - distribute work to multiple workers
	numWorkers := 3
	workers := make([]<-chan int, numWorkers)

	// Create channels for each worker to read from
	// We need to distribute input to multiple workers
	inputChannels := make([]chan int, numWorkers)
	for i := 0; i < numWorkers; i++ {
		inputChannels[i] = make(chan int)
		workers[i] = square(inputChannels[i])
	}

	// Distribute input data to workers in round-robin fashion
	go func() {
		i := 0
		for n := range input {
			inputChannels[i%numWorkers] <- n
			i++
		}
		// Close all input channels when done
		for _, ch := range inputChannels {
			close(ch)
		}
	}()

	// Stage 3: Fan-In - merge all results into one channel
	results := fanIn(workers...)

	// Collect and print results
	for result := range results {
		fmt.Println("Result:", result)
	}
}
