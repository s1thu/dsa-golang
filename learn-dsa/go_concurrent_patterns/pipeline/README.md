# Go Pipeline Pattern

## What is a Pipeline?

A **Pipeline** is a concurrency pattern where data flows through a series of stages, each connected by channels. Each stage is a group of goroutines running the same function that:

1. Receives values from an **input channel**
2. Performs some **transformation** on the data
3. Sends results to an **output channel**

```
┌──────────────────────────────────────────────────────────────────────────────────┐
│                                                                                  │
│                              Pipeline Pattern                                    │
│                                                                                  │
│   ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐     ┌─────────┐   │
│   │ Generate│────▶│ Stage 1 │────▶│ Stage 2 │────▶│ Stage 3 │────▶│ Consumer│   │
│   │  (1,2,3)│     │ Square  │     │ Add 10  │     │ Double  │     │ (Print) │   │
│   └─────────┘     └─────────┘     └─────────┘     └─────────┘     └─────────┘   │
│                                                                                  │
│   Data Flow:  1 ─▶ 1  ─▶ 11 ─▶ 22                                               │
│               2 ─▶ 4  ─▶ 14 ─▶ 28                                               │
│               3 ─▶ 9  ─▶ 19 ─▶ 38                                               │
│                                                                                  │
└──────────────────────────────────────────────────────────────────────────────────┘
```

---

## Why Use Pipelines?

| Problem Without Pipelines           | Solution With Pipelines               |
| ----------------------------------- | ------------------------------------- |
| Tightly coupled processing logic    | Each stage is independent and modular |
| Sequential bottleneck               | Concurrent stages process in parallel |
| Hard to add/remove processing steps | Easy to chain or modify stages        |
| Difficult to test individual steps  | Each stage can be tested in isolation |
| Memory-heavy batch processing       | Streaming data with bounded memory    |

### Real-World Analogy 🏭

Think of an **assembly line** in a factory:

- **Stage 1**: Raw materials enter → Cut into shape
- **Stage 2**: Shaped materials → Paint applied
- **Stage 3**: Painted items → Quality check
- **Stage 4**: Approved items → Packaging

Each station works **concurrently** - while Station 1 cuts the next item, Station 2 paints the previous one.

Another example - **Data Processing**:

- **Stage 1**: Read raw data from file
- **Stage 2**: Parse and validate data
- **Stage 3**: Transform/enrich data
- **Stage 4**: Write to database

---

## Code Breakdown - Step by Step

### Step 1: Generator Function (Source Stage)

```go
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
```

**Why each part?**

| Code          | Purpose                                             |
| ------------- | --------------------------------------------------- |
| `nums ...int` | Variadic parameter - accepts any number of integers |
| `<-chan int`  | Returns a **receive-only** channel                  |
| `go func()`   | Sends data concurrently without blocking            |
| `close(out)`  | Signals that no more data will be sent              |

---

### Step 2: Processing Stages (Transform Stages)

```go
func square(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * n
        }
        close(out)
    }()
    return out
}

func addTen(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n + 10
        }
        close(out)
    }()
    return out
}
```

**Why each part?**

| Code               | Purpose                                         |
| ------------------ | ----------------------------------------------- |
| `in <-chan int`    | Receives from **upstream** stage (receive-only) |
| `out := make(...)` | Creates output channel for **downstream**       |
| `for n := range`   | Iterates until input channel is closed          |
| `out <- n * n`     | Sends transformed data downstream               |
| `close(out)`       | Propagates completion signal downstream         |

---

### Step 3: Chaining Stages Together

```go
func main() {
    // Create pipeline by chaining stages
    numbers := generate(1, 2, 3, 4, 5)
    squared := square(numbers)
    addedTen := addTen(squared)
    doubled := double(addedTen)

    // Consume the final output
    for result := range doubled {
        fmt.Println("Result:", result)
    }
}
```

**Key Insight**: Each function call returns immediately! The actual processing happens concurrently in goroutines.

---

### Step 4: Generic Transform Helper (Advanced)

```go
func transform(in <-chan int, fn func(int) int, name string) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- fn(n)
        }
        close(out)
    }()
    return out
}
```

This allows you to create pipeline stages dynamically:

```go
stage1 := transform(input, func(n int) int { return n * 3 }, "triple")
stage2 := transform(stage1, func(n int) int { return n - 5 }, "subtract")
```

---

## Pipeline vs Other Patterns

### Pipeline vs Fan-Out/Fan-In

| Aspect        | Pipeline                   | Fan-Out/Fan-In                |
| ------------- | -------------------------- | ----------------------------- |
| **Flow**      | Linear, sequential stages  | Parallel workers, then merge  |
| **Use Case**  | Multi-step transformations | CPU-bound parallel processing |
| **Structure** | A → B → C → D              | A → [B1, B2, B3] → C          |
| **Focus**     | Data transformation chain  | Distributing workload         |

### When to Use Each

- **Pipeline**: When data needs sequential transformations (parse → validate → transform → store)
- **Fan-Out/Fan-In**: When same operation needs parallel execution (process many files concurrently)

---

## Key Properties of Pipelines

### 1. Bounded Memory

Data is processed **one item at a time** (streaming), not all at once (batch):

```go
// ❌ Batch - loads everything into memory
allData := loadAllData()
transformed := transformAll(allData)

// ✅ Pipeline - bounded memory usage
for item := range pipeline(dataSource) {
    process(item)
}
```

### 2. Backpressure

Slow consumers automatically slow down producers (unbuffered channels block):

```go
// If consumer is slow, producer blocks on send
producer -> [channel] -> slowConsumer
```

### 3. Graceful Shutdown

Closing the source channel propagates through all stages:

```go
close(source) → stage1 closes → stage2 closes → consumer exits
```

---

## Running the Example

```bash
go run main.go
```

**Expected Output:**

```
=== Simple Pipeline Demo ===
Pipeline: generate -> square -> addTen -> double

Squaring 1
Adding 10 to 1
Squaring 2
Doubling 11
Final Result: 22
Adding 10 to 4
Squaring 3
Doubling 14
Final Result: 28
...
```

Notice how stages interleave - while one number is being doubled, another is being squared!

---

## Common Pitfalls

### 1. Forgetting to Close Channels

```go
// ❌ Consumer will block forever
func badStage(in <-chan int) <-chan int {
    out := make(chan int)
    go func() {
        for n := range in {
            out <- n * 2
        }
        // Missing: close(out)
    }()
    return out
}
```

### 2. Channel Direction

```go
// ✅ Correct - restricts channel usage
func stage(in <-chan int) <-chan int  // receive-only input, receive-only output

// ❌ Avoid - allows unintended writes
func stage(in chan int) chan int
```

### 3. Blocking Main Goroutine

```go
// ❌ May exit before pipeline completes
go func() {
    for result := range pipeline {
        fmt.Println(result)
    }
}()
// main exits immediately

// ✅ Block until pipeline completes
for result := range pipeline {
    fmt.Println(result)
}
```

---

## Summary

| Concept          | Description                                        |
| ---------------- | -------------------------------------------------- |
| **Pipeline**     | Chain of stages connected by channels              |
| **Stage**        | Goroutine that transforms data and passes it along |
| **Backpressure** | Slow consumers slow down the entire pipeline       |
| **Bounded**      | Fixed memory usage regardless of data volume       |
| **Composable**   | Stages can be easily added, removed, or reordered  |

The Pipeline pattern is fundamental to Go's concurrency model and is used extensively in production systems for:

- ETL (Extract, Transform, Load) processes
- Image/video processing
- Log processing and aggregation
- Real-time data streaming
