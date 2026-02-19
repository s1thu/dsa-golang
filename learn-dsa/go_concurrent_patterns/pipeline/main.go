package main

import (
	"fmt"
	"time"
)

// Stage 1: Generate numbers
func generate(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		for _, n := range nums {
			out <- n
		}
		close(out)
	}()
	return out
}

// Stage 2: Square each number
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			fmt.Printf("Squaring %d\n", n)
			time.Sleep(500 * time.Millisecond) // Simulate work
			out <- n * n
		}
		close(out)
	}()
	return out
}

// Stage 3: Add 10 to each number
func addTen(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			fmt.Printf("Adding 10 to %d\n", n)
			time.Sleep(300 * time.Millisecond) // Simulate work
			out <- n + 10
		}
		close(out)
	}()
	return out
}

// Stage 4: Double each number
func double(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			fmt.Printf("Doubling %d\n", n)
			time.Sleep(200 * time.Millisecond) // Simulate work
			out <- n * 2
		}
		close(out)
	}()
	return out
}

// Generic pipeline stage helper - applies any function to values
func transform(in <-chan int, fn func(int) int, name string) <-chan int {
	out := make(chan int)
	go func() {
		for n := range in {
			fmt.Printf("[%s] Processing %d\n", name, n)
			out <- fn(n)
		}
		close(out)
	}()
	return out
}

func main() {
	fmt.Println("=== Simple Pipeline Demo ===")
	fmt.Println("Pipeline: generate -> square -> addTen -> double")
	fmt.Println()

	// Create the pipeline by chaining stages
	// Data flows: generate -> square -> addTen -> double -> consumer
	numbers := generate(1, 2, 3, 4, 5)
	squared := square(numbers)
	addedTen := addTen(squared)
	doubled := double(addedTen)

	// Consume the final output
	for result := range doubled {
		fmt.Println("Final Result:", result)
	}

	fmt.Println()
	fmt.Println("=== Generic Transform Pipeline Demo ===")
	fmt.Println()

	// Alternative: Using the generic transform helper
	// Pipeline: generate -> triple -> subtract5 -> square
	input := generate(1, 2, 3)
	stage1 := transform(input, func(n int) int { return n * 3 }, "triple")
	stage2 := transform(stage1, func(n int) int { return n - 5 }, "subtract5")
	stage3 := transform(stage2, func(n int) int { return n * n }, "square")

	for result := range stage3 {
		fmt.Println("Generic Pipeline Result:", result)
	}
}
