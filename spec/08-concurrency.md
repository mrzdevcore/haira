# 8. Concurrency

## 8.1 Philosophy

Haira provides simple, safe concurrency primitives. No manual threads, no locks, no shared mutable state. Communication happens through message passing.

## 8.2 Async Blocks

Run operations concurrently and wait for all to complete:

```haira
async {
    users = get_all_users()
    posts = get_all_posts()
    metrics = get_metrics()
}

// All three run concurrently
// Block completes when all finish
// Results available after block

print("Users: {users.count}")
print("Posts: {posts.count}")
```

### Async with Dependencies

```haira
// Independent operations run in parallel
async {
    users = fetch_users()       // These three run
    posts = fetch_posts()       // concurrently
    config = load_config()
}

// Dependent operations must be sequential
async {
    user = fetch_user(id)
}
profile = build_profile(user)   // Needs user first
```

## 8.3 Spawn (Fire and Forget)

Start background work without waiting:

```haira
spawn {
    process_large_file(path)
    send_notification("done")
}

// Continues immediately
print("Processing started in background")
```

### Spawn with Handle

```haira
task = spawn {
    heavy_computation()
}

// Do other work...

// Wait when needed
result = task.wait()
```

## 8.4 Channels

Typed communication between concurrent operations:

### Creating Channels

```haira
// Unbuffered channel (synchronous)
messages = channel()

// Buffered channel
messages = channel(capacity: 100)
```

### Sending and Receiving

```haira
// Producer
spawn {
    for i in 0..10 {
        messages.send(i)
    }
    messages.close()
}

// Consumer
for msg in messages {
    print("Received: {msg}")
}
```

### Channel Operations

```haira
ch = channel()

ch.send(value)              // Send (blocks if full)
value = ch.receive()        // Receive (blocks if empty)
ch.close()                  // Close channel

// Non-blocking variants
sent = ch.try_send(value)   // Returns bool
value = ch.try_receive()    // Returns Option
```

## 8.5 Select

Wait on multiple channels:

```haira
messages = channel()
errors = channel()
timeout = after(5000)       // 5 seconds

select {
    msg from messages => {
        print("Got message: {msg}")
    }
    err from errors => {
        print("Got error: {err}")
    }
    _ from timeout => {
        print("Timed out")
    }
}
```

### Select with Default

```haira
select {
    msg from messages => handle(msg)
    default => {
        // No message available, do something else
        print("Nothing yet")
    }
}
```

### Select in Loop

```haira
while true {
    select {
        msg from messages => process(msg)
        cmd from commands => execute(cmd)
        _ from quit => break
    }
}
```

## 8.6 Parallel Operations

### Parallel Map

Process items concurrently:

```haira
// Process all users in parallel
results = users | parallel(u => process_user(u))

// With concurrency limit
results = users | parallel(u => fetch_data(u), limit: 10)
```

### Parallel Each

Side effects in parallel:

```haira
users | parallel_each(u => send_email(u))
```

### Collect Results

```haira
results = items
    | parallel(process)
    | collect              // Wait for all and gather results
```

## 8.7 Workers

Process queue with worker pool:

```haira
// Create worker queue
queue = worker_queue(process_job, workers: 4)

// Add jobs
for job in jobs {
    queue.add(job)
}

// Wait for completion
queue.wait()
```

### Worker with Results

```haira
queue = worker_queue(workers: 4)

for item in items {
    queue.add(item, process_item)
}

results = queue.results()   // Collect all results
```

## 8.8 Timeouts

### Operation Timeout

```haira
result = with_timeout(5000) {
    slow_operation()
}

if result.timed_out {
    print("Operation took too long")
} else {
    print("Result: {result.value}")
}
```

### Timer

```haira
timer = after(1000)         // 1 second

select {
    _ from timer => print("Time's up!")
}
```

### Interval

```haira
ticker = every(1000)        // Every second

spawn {
    for _ in ticker {
        print("Tick")
    }
}
```

## 8.9 Synchronization Primitives

### WaitGroup

```haira
wg = wait_group()

for item in items {
    wg.add(1)
    spawn {
        process(item)
        wg.done()
    }
}

wg.wait()                   // Wait for all to complete
```

### Once

```haira
init = once()

// Only first call executes
init.do {
    expensive_initialization()
}
```

## 8.10 Thread Safety

Haira encourages message passing over shared state. When shared state is needed:

### Atomic Values

```haira
counter = atomic(0)

spawn {
    counter.add(1)
}

value = counter.load()
```

### Mutex (when necessary)

```haira
data = mutex({ count: 0 })

spawn {
    data.lock {
        it.count = it.count + 1
    }
}
```

## 8.11 Best Practices

### Do

```haira
// Use channels for communication
results = channel()
spawn { results.send(compute()) }
value = results.receive()

// Use parallel for bulk operations
processed = items | parallel(transform)

// Set concurrency limits
results = urls | parallel(fetch, limit: 10)
```

### Don't

```haira
// Don't share mutable state
shared = { count: 0 }
spawn { shared.count = 1 }   // Bad!

// Don't forget to close channels
ch = channel()
spawn {
    for item in items {
        ch.send(item)
    }
    // ch.close() - Don't forget!
}
```
