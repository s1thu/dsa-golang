# Go Worker Pool Pattern

## What is a Worker Pool?

A **Worker Pool** (also known as Thread Pool) is a concurrency pattern where a fixed number of workers (goroutines) process jobs from a shared queue. Instead of creating a new goroutine for every task, we reuse a limited set of workers to handle multiple jobs.

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   Jobs Queue          Workers              Results Queue    │
│   ┌─────────┐        ┌─────────┐          ┌─────────┐      │
│   │ Job 1   │───────▶│Worker 1 │─────────▶│Result 1 │      │
│   │ Job 2   │───────▶│Worker 2 │─────────▶│Result 2 │      │
│   │ Job 3   │───────▶│Worker 3 │─────────▶│Result 3 │      │
│   │ Job 4   │        └─────────┘          │Result 4 │      │
│   │ Job 5   │                             │Result 5 │      │
│   └─────────┘                             └─────────┘      │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

---

## Why Use Worker Pool?

| Problem Without Worker Pool                             | Solution With Worker Pool                   |
| ------------------------------------------------------- | ------------------------------------------- |
| Creating thousands of goroutines for thousands of tasks | Fixed number of goroutines handle all tasks |
| Memory explosion (each goroutine uses ~2KB stack)       | Controlled memory usage                     |
| CPU thrashing from context switching                    | Predictable resource consumption            |
| Uncontrolled concurrency                                | Bounded concurrency                         |

### Real-World Analogy 🏪

Think of a **fast food restaurant**:

- **Without Worker Pool**: Hire a new employee for every customer order, then fire them after serving. Chaos!
- **With Worker Pool**: Have 3-5 employees (workers) at the counter. Customers (jobs) line up in a queue. Each employee picks the next order when they're free.

---

## Code Breakdown - Step by Step

### Step 1: Define the Worker Function

```go
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        fmt.Printf("Worker %d started job %d\n", id, job)
        time.Sleep(time.Second)  // Simulate work
        fmt.Printf("Worker %d finished job %d\n", id, job)
        results <- job * 2
    }
}
```

**Why each part?**

| Code                    | Purpose                                                                |
| ----------------------- | ---------------------------------------------------------------------- |
| `id int`                | Identifies which worker is processing (useful for debugging/logging)   |
| `jobs <-chan int`       | **Receive-only** channel - worker can only READ jobs from this channel |
| `results chan<- int`    | **Send-only** channel - worker can only WRITE results to this channel  |
| `wg *sync.WaitGroup`    | Tracks when all workers are done                                       |
| `defer wg.Done()`       | Signals "I'm done" when function exits                                 |
| `for job := range jobs` | Keeps taking jobs until channel is closed                              |
| `results <- job * 2`    | Sends the processed result                                             |

**Why receive-only and send-only channels?**

- **Type Safety**: Prevents bugs - workers can't accidentally close the jobs channel
- **Clear Intent**: Makes code self-documenting about data flow direction
- **Compile-time Checks**: Go compiler enforces these restrictions

---

### Step 2: Set Up Configuration

```go
const numJobs = 5
const numWorkers = 3
```

**Why?**

- `numJobs`: Total work items to process
- `numWorkers`: How many goroutines run in parallel

**Rule of Thumb for `numWorkers`:**

- **CPU-bound tasks**: `runtime.NumCPU()` (number of CPU cores)
- **I/O-bound tasks**: Can be higher (10x-100x CPU cores) since they spend time waiting

---

### Step 3: Create Buffered Channels

```go
jobs := make(chan int, numJobs)
results := make(chan int, numJobs)
```

**Why buffered channels?**

| Unbuffered `make(chan int)`           | Buffered `make(chan int, 5)`                   |
| ------------------------------------- | ---------------------------------------------- |
| Sender blocks until receiver is ready | Sender can send up to N items without blocking |
| Tight synchronization                 | Decouples sender and receiver                  |
| Good for coordination                 | Good for throughput                            |

In worker pools, **buffered channels act as queues**, allowing the main goroutine to queue up all jobs without waiting.

---

### Step 4: Launch Workers

```go
var wg sync.WaitGroup

for i := 1; i <= numWorkers; i++ {
    wg.Add(1)
    go worker(i, jobs, results, &wg)
}
```

**Why this order matters:**

1. **`wg.Add(1)` BEFORE `go worker(...)`** - Must register before goroutine starts
2. **Pass `&wg` (pointer)** - All workers share the SAME WaitGroup instance

**What happens:**

```
Timeline:
─────────────────────────────────────────▶
    │
    ├── Worker 1 starts, waiting for jobs
    ├── Worker 2 starts, waiting for jobs
    └── Worker 3 starts, waiting for jobs
```

All 3 workers are now running and blocked on `for job := range jobs`, waiting for work.

---

### Step 5: Send Jobs

```go
for j := 1; j <= numJobs; j++ {
    jobs <- j
}
close(jobs)
```

**Why close the channel?**

- `close(jobs)` signals to workers: "No more jobs coming!"
- Without closing, workers would wait forever at `for job := range jobs`
- This makes the `range` loop exit cleanly

**Visual:**

```
Jobs Channel: [1][2][3][4][5] → CLOSED

Worker 1: Takes 1, processes...
Worker 2: Takes 2, processes...
Worker 3: Takes 3, processes...
Worker 1: (finished 1) Takes 4, processes...
Worker 2: (finished 2) Takes 5, processes...
Worker 3: (finished 3) No more jobs, range exits
```

---

### Step 6: Wait and Collect Results

```go
wg.Wait()
close(results)

for result := range results {
    fmt.Println("Result:", result)
}
```

**Why this sequence?**

1. `wg.Wait()` - Block until ALL workers finish ALL jobs
2. `close(results)` - Signal that no more results will be sent
3. `range results` - Read all results from the buffered channel

**Important:** We close `results` AFTER `wg.Wait()` because:

- Workers might still be writing to `results`
- Writing to a closed channel causes **panic**!

---

## Execution Flow Visualization

```
Time ──────────────────────────────────────────────────────────────▶

Main:    [Create channels] [Start workers] [Send jobs] [Wait...] [Read results]
                                │              │           │
Worker1: ──────────────────[Job1: 1s]─────[Job4: 1s]──────Done
Worker2: ──────────────────[Job2: 1s]─────[Job5: 1s]──────Done
Worker3: ──────────────────[Job3: 1s]─────────────────────Done

Total time: ~2 seconds (not 5 seconds!)
```

Without worker pool: 5 seconds (sequential)
With 3 workers: ~2 seconds (parallel)

---

## Real-World Use Cases

### 1. 🖼️ Image Processing Service

```go
// Process user-uploaded images (resize, compress, add watermark)
func imageWorker(jobs <-chan Image, results chan<- ProcessedImage, wg *sync.WaitGroup) {
    defer wg.Done()
    for img := range jobs {
        resized := resize(img, 800, 600)
        compressed := compress(resized, 80)
        watermarked := addWatermark(compressed)
        results <- watermarked
    }
}
```

**Why worker pool?** Limit concurrent image processing to avoid running out of memory.

---

### 2. 🌐 Web Scraper

```go
// Fetch data from multiple URLs
func scrapeWorker(urls <-chan string, results chan<- PageData, wg *sync.WaitGroup) {
    defer wg.Done()
    client := &http.Client{Timeout: 10 * time.Second}
    for url := range urls {
        resp, _ := client.Get(url)
        data := parseHTML(resp.Body)
        results <- data
    }
}
```

**Why worker pool?** Respect rate limits, avoid overwhelming servers, control memory for parsed data.

---

### 3. 📧 Email Sender

```go
// Send bulk marketing emails
func emailWorker(emails <-chan Email, status chan<- SendStatus, wg *sync.WaitGroup) {
    defer wg.Done()
    for email := range emails {
        err := smtpClient.Send(email)
        status <- SendStatus{Email: email.To, Success: err == nil}
    }
}
```

**Why worker pool?** SMTP servers have connection limits, prevent spam detection.

---

### 4. 🗄️ Database Batch Operations

```go
// Insert records into database
func dbWorker(records <-chan Record, wg *sync.WaitGroup) {
    defer wg.Done()
    db := connectDB()
    defer db.Close()

    for record := range records {
        db.Exec("INSERT INTO users VALUES (?)", record)
    }
}
```

**Why worker pool?** Database connection pools are limited. Reuse connections efficiently.

---

### 5. 📁 File Processing Pipeline

```go
// Process log files
func logWorker(files <-chan string, results chan<- LogSummary, wg *sync.WaitGroup) {
    defer wg.Done()
    for filepath := range files {
        content := readFile(filepath)
        errors := countErrors(content)
        warnings := countWarnings(content)
        results <- LogSummary{File: filepath, Errors: errors, Warnings: warnings}
    }
}
```

**Why worker pool?** Disk I/O is slow; limit concurrent file reads to prevent thrashing.

---

## Common Patterns & Variations

### Pattern 1: Dynamic Job Generation

```go
// Producer goroutine generates jobs dynamically
go func() {
    for i := 0; ; i++ {
        if shouldStop() {
            break
        }
        jobs <- i
    }
    close(jobs)
}()
```

### Pattern 2: Error Handling

```go
type Result struct {
    Value int
    Err   error
}

func worker(jobs <-chan int, results chan<- Result, wg *sync.WaitGroup) {
    defer wg.Done()
    for job := range jobs {
        value, err := process(job)
        results <- Result{Value: value, Err: err}
    }
}
```

### Pattern 3: Graceful Shutdown with Context

```go
func worker(ctx context.Context, jobs <-chan int, wg *sync.WaitGroup) {
    defer wg.Done()
    for {
        select {
        case <-ctx.Done():
            return  // Shutdown signal received
        case job, ok := <-jobs:
            if !ok {
                return  // Channel closed
            }
            process(job)
        }
    }
}
```

---

## Best Practices

| Do ✅                                   | Don't ❌                                  |
| --------------------------------------- | ----------------------------------------- |
| Use buffered channels for jobs/results  | Use unbuffered channels (causes blocking) |
| Close channels from sender side         | Close channels from receiver side         |
| Use `sync.WaitGroup` for coordination   | Use `time.Sleep` to wait for completion   |
| Size worker pool based on workload type | Create unlimited goroutines               |
| Use `context.Context` for cancellation  | Kill goroutines forcefully                |

---

## Summary

The Worker Pool pattern gives you:

1. **Controlled Concurrency** - Fixed number of goroutines
2. **Resource Efficiency** - Reuse workers instead of creating new ones
3. **Backpressure** - Buffered channels queue work when workers are busy
4. **Clean Shutdown** - Close channel to signal "no more work"
5. **Easy Scaling** - Change `numWorkers` to tune performance

```go
// The essence of worker pool in 4 lines:
workers := startWorkers(n)    // 1. Start fixed workers
sendJobs(jobs)                // 2. Feed jobs to channel
close(jobs)                   // 3. Signal completion
waitForResults()              // 4. Collect results
```
