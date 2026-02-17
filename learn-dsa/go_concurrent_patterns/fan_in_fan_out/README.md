# Go Fan-In Fan-Out Pattern

## What is Fan-In Fan-Out?

**Fan-Out** and **Fan-In** are concurrency patterns used to parallelize work and then aggregate results.

- **Fan-Out**: Multiple goroutines read from the same channel, distributing work among them
- **Fan-In**: A single goroutine reads from multiple input channels and merges them into one output channel

```
┌──────────────────────────────────────────────────────────────────────────┐
│                                                                          │
│   Input              Fan-Out                 Fan-In           Output     │
│                                                                          │
│                     ┌─────────┐                                          │
│                  ┌─▶│Worker 1 │──┐                                       │
│                  │  └─────────┘  │                                       │
│   ┌─────────┐    │  ┌─────────┐  │         ┌─────────┐    ┌─────────┐   │
│   │  Data   │────┼─▶│Worker 2 │──┼────────▶│  Merge  │───▶│ Results │   │
│   └─────────┘    │  └─────────┘  │         └─────────┘    └─────────┘   │
│                  │  ┌─────────┐  │                                       │
│                  └─▶│Worker 3 │──┘                                       │
│                     └─────────┘                                          │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Why Use Fan-In Fan-Out?

| Problem Without Fan-In Fan-Out          | Solution With Fan-In Fan-Out       |
| --------------------------------------- | ---------------------------------- |
| Sequential processing of data           | Parallel processing across workers |
| Single point of bottleneck              | Multiple workers share the load    |
| Wasted CPU cycles waiting               | Better CPU utilization             |
| Scattered results from multiple sources | Unified output stream              |

### Real-World Analogy 🏭

Think of a **factory assembly line**:

- **Fan-Out**: Raw materials arrive and are distributed to multiple workstations. Each station processes items concurrently.
- **Fan-In**: Products from all workstations converge onto a single conveyor belt for packaging.

Another example - **Customer Support**:

- **Fan-Out**: Incoming support tickets are distributed among available agents
- **Fan-In**: All resolved tickets are merged into a single report/dashboard

---

## Code Breakdown - Step by Step

### Step 1: Generator Function (Data Source)

```go
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
```

**Why each part?**

| Code          | Purpose                                             |
| ------------- | --------------------------------------------------- |
| `nums ...int` | Variadic parameter - accepts any number of integers |
| `<-chan int`  | Returns a **receive-only** channel                  |
| `go func()`   | Sends data concurrently without blocking            |
| `close(out)`  | Signals that no more data will be sent              |

---

### Step 2: Worker Function (Fan-Out Stage)

```go
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
```

**Why each part?**

| Code                    | Purpose                                     |
| ----------------------- | ------------------------------------------- |
| `in <-chan int`         | **Receive-only** input channel              |
| `out := make(chan int)` | Creates output channel for results          |
| `for n := range in`     | Keeps reading until input channel closes    |
| `out <- n * n`          | Sends processed result                      |
| `close(out)`            | Signals downstream that this worker is done |

**Key Insight**: Multiple `square` workers can run in parallel, each processing different data items.

---

### Step 3: Fan-In Function (Merge Results)

```go
func fanIn(channels ...<-chan int) <-chan int {
    out := make(chan int)
    var wg sync.WaitGroup

    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                out <- val
            }
        }(ch)
    }

    go func() {
        wg.Wait()
        close(out)
    }()

    return out
}
```

**Why each part?**

| Code                     | Purpose                                     |
| ------------------------ | ------------------------------------------- |
| `channels ...<-chan int` | Accepts multiple input channels             |
| `var wg sync.WaitGroup`  | Tracks when all channels are drained        |
| `go func(c <-chan int)`  | Spawns goroutine for each input channel     |
| `defer wg.Done()`        | Signals when channel is fully read          |
| `wg.Wait()`              | Waits for all input channels to complete    |
| `close(out)`             | Closes output only when ALL inputs are done |

**Critical Point**: The `WaitGroup` ensures we don't close the output channel prematurely.

---

### Step 4: Main Function (Orchestration)

```go
func main() {
    // Stage 1: Generate data
    input := generator(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)

    // Stage 2: Fan-Out - distribute work to multiple workers
    numWorkers := 3
    workers := make([]<-chan int, numWorkers)

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
```

---

## Visual Flow

```
Input Numbers: [1, 2, 3, 4, 5, 6, 7, 8, 9, 10]
                        │
                        ▼
              ┌─────────────────┐
              │    Generator    │
              └─────────────────┘
                        │
         ┌──────────────┼──────────────┐
         │              │              │
         ▼              ▼              ▼
   ┌──────────┐  ┌──────────┐  ┌──────────┐
   │ Worker 1 │  │ Worker 2 │  │ Worker 3 │
   │ (Square) │  │ (Square) │  │ (Square) │
   │ 1,4,7,10 │  │  2,5,8   │  │  3,6,9   │
   └──────────┘  └──────────┘  └──────────┘
         │              │              │
         │       1,16,49,100    4,25,64    9,36,81
         │              │              │
         └──────────────┼──────────────┘
                        │
                        ▼
              ┌─────────────────┐
              │     Fan-In      │
              │  (Merge All)    │
              └─────────────────┘
                        │
                        ▼
          Results: [1, 4, 9, 16, 25, 36, 49, 64, 81, 100]
          (Order may vary due to concurrent execution)
```

---

## Key Differences: Fan-In Fan-Out vs Worker Pool

| Aspect            | Worker Pool                             | Fan-In Fan-Out                             |
| ----------------- | --------------------------------------- | ------------------------------------------ |
| **Structure**     | Fixed workers pulling from shared queue | Pipeline stages with distributed workers   |
| **Data Flow**     | Jobs → Workers → Results                | Source → Fan-Out → Workers → Fan-In → Sink |
| **Use Case**      | Task queue processing                   | Data pipeline processing                   |
| **Flexibility**   | Workers do same task                    | Each stage can do different operations     |
| **Composability** | Single stage                            | Multiple stages can be chained             |

---

## When to Use Fan-In Fan-Out?

✅ **Good for:**

- Data pipelines with multiple transformation stages
- Aggregating data from multiple sources
- Parallel processing with result consolidation
- ETL (Extract, Transform, Load) operations

❌ **Not ideal for:**

- Simple task queues (use Worker Pool instead)
- When order must be strictly preserved
- Very simple transformations

---

## Common Pitfalls

1. **Forgetting to close channels**: Results in goroutine leaks
2. **Closing output channel too early**: Causes panic when workers try to send
3. **Not using WaitGroup in Fan-In**: Output channel closes before all data is received
4. **Race conditions**: Ensure proper synchronization

---

## Run the Code

```bash
go run main.go
```

**Sample Output:**

```
Processing 1
Processing 2
Processing 3
Result: 1
Processing 4
Result: 4
Processing 5
Result: 9
Processing 6
Result: 16
Processing 7
Result: 25
Processing 8
Result: 36
Processing 9
Result: 49
Processing 10
Result: 64
Result: 81
Result: 100
```

_Note: Output order may vary due to concurrent execution_
