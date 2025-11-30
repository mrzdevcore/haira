// Haira Runtime Library
// This provides basic I/O and memory operations for compiled Haira programs.

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>

// Print a string (pointer + length)
void haira_print(const char* str, int64_t len) {
    fwrite(str, 1, (size_t)len, stdout);
}

// Print an integer
void haira_print_int(int64_t value) {
    printf("%lld", (long long)value);
}

// Print a float
void haira_print_float(double value) {
    printf("%g", value);
}

// Print a boolean
void haira_print_bool(int8_t value) {
    printf("%s", value ? "true" : "false");
}

// Print a newline
void haira_println(void) {
    printf("\n");
    fflush(stdout);
}

// Memory allocation
void* haira_alloc(int64_t size) {
    return malloc((size_t)size);
}

// Memory reallocation
void* haira_realloc(void* ptr, int64_t size) {
    return realloc(ptr, (size_t)size);
}

// Memory deallocation
void haira_free(void* ptr) {
    free(ptr);
}

// String concatenation - returns new string
typedef struct {
    char* data;
    int64_t len;
    int64_t cap;
} HairaString;

HairaString* haira_string_concat(const char* a, int64_t alen, const char* b, int64_t blen) {
    HairaString* result = (HairaString*)malloc(sizeof(HairaString));
    result->len = alen + blen;
    result->cap = result->len + 1;
    result->data = (char*)malloc((size_t)result->cap);
    memcpy(result->data, a, (size_t)alen);
    memcpy(result->data + alen, b, (size_t)blen);
    result->data[result->len] = '\0';
    return result;
}

// Integer to string
HairaString* haira_int_to_string(int64_t value) {
    HairaString* result = (HairaString*)malloc(sizeof(HairaString));
    result->data = (char*)malloc(32);
    result->len = snprintf(result->data, 32, "%lld", (long long)value);
    result->cap = 32;
    return result;
}

// Float to string
HairaString* haira_float_to_string(double value) {
    HairaString* result = (HairaString*)malloc(sizeof(HairaString));
    result->data = (char*)malloc(64);
    result->len = snprintf(result->data, 64, "%g", value);
    result->cap = 64;
    return result;
}

// Panic/abort
void haira_panic(const char* msg, int64_t len) {
    fprintf(stderr, "panic: ");
    fwrite(msg, 1, (size_t)len, stderr);
    fprintf(stderr, "\n");
    exit(1);
}

// Error handling
// Thread-local current error (0 = no error, non-zero = error value)
static __thread int64_t haira_current_error = 0;

// Set current error
void haira_set_error(int64_t error) {
    haira_current_error = error;
}

// Get and clear current error
int64_t haira_get_error(void) {
    int64_t err = haira_current_error;
    haira_current_error = 0;
    return err;
}

// Check if there's an error
int64_t haira_has_error(void) {
    return haira_current_error != 0 ? 1 : 0;
}

// Clear error
void haira_clear_error(void) {
    haira_current_error = 0;
}

// ============================================================================
// Concurrency - Spawn/Threads
// ============================================================================

#include <pthread.h>
#include <unistd.h>

// Thread wrapper for spawn
typedef struct {
    void (*func)(void);
} SpawnArgs;

static void* spawn_thread_wrapper(void* arg) {
    SpawnArgs* args = (SpawnArgs*)arg;
    args->func();
    free(args);
    return NULL;
}

// Spawn a new thread running the given function (fire-and-forget)
// Returns thread handle (pthread_t cast to int64_t)
int64_t haira_spawn(void (*func)(void)) {
    pthread_t thread;
    SpawnArgs* args = (SpawnArgs*)malloc(sizeof(SpawnArgs));
    args->func = func;
    
    if (pthread_create(&thread, NULL, spawn_thread_wrapper, args) != 0) {
        free(args);
        return 0; // Error
    }
    
    // Detach the thread so it cleans up automatically
    pthread_detach(thread);
    
    return (int64_t)thread;
}

// Spawn a new thread that can be joined (for async blocks)
// Returns thread handle (pthread_t cast to int64_t)
int64_t haira_spawn_joinable(void (*func)(void)) {
    pthread_t* thread = (pthread_t*)malloc(sizeof(pthread_t));
    SpawnArgs* args = (SpawnArgs*)malloc(sizeof(SpawnArgs));
    args->func = func;
    
    if (pthread_create(thread, NULL, spawn_thread_wrapper, args) != 0) {
        free(args);
        free(thread);
        return 0; // Error
    }
    
    // Return pointer to thread handle (so we can join later)
    return (int64_t)thread;
}

// Wait for a joinable thread to complete
void haira_thread_join(int64_t thread_handle) {
    if (thread_handle == 0) return;
    pthread_t* thread = (pthread_t*)thread_handle;
    pthread_join(*thread, NULL);
    free(thread);
}

// Sleep for milliseconds
void haira_sleep(int64_t ms) {
    usleep((useconds_t)(ms * 1000));
}

// ============================================================================
// Channels
// ============================================================================

typedef struct {
    int64_t* buffer;
    int64_t capacity;
    int64_t count;
    int64_t read_pos;
    int64_t write_pos;
    pthread_mutex_t mutex;
    pthread_cond_t not_empty;
    pthread_cond_t not_full;
    int closed;
} HairaChannel;

// Create a new channel with given capacity (0 = unbuffered/sync)
HairaChannel* haira_channel_new(int64_t capacity) {
    HairaChannel* ch = (HairaChannel*)malloc(sizeof(HairaChannel));
    ch->capacity = capacity > 0 ? capacity : 1;
    ch->buffer = (int64_t*)malloc(sizeof(int64_t) * (size_t)ch->capacity);
    ch->count = 0;
    ch->read_pos = 0;
    ch->write_pos = 0;
    ch->closed = 0;
    pthread_mutex_init(&ch->mutex, NULL);
    pthread_cond_init(&ch->not_empty, NULL);
    pthread_cond_init(&ch->not_full, NULL);
    return ch;
}

// Send a value to the channel (blocks if full)
void haira_channel_send(HairaChannel* ch, int64_t value) {
    pthread_mutex_lock(&ch->mutex);
    
    while (ch->count == ch->capacity && !ch->closed) {
        pthread_cond_wait(&ch->not_full, &ch->mutex);
    }
    
    if (!ch->closed) {
        ch->buffer[ch->write_pos] = value;
        ch->write_pos = (ch->write_pos + 1) % ch->capacity;
        ch->count++;
        pthread_cond_signal(&ch->not_empty);
    }
    
    pthread_mutex_unlock(&ch->mutex);
}

// Receive a value from the channel (blocks if empty)
// Returns the value, or 0 if channel is closed and empty
int64_t haira_channel_receive(HairaChannel* ch) {
    pthread_mutex_lock(&ch->mutex);
    
    while (ch->count == 0 && !ch->closed) {
        pthread_cond_wait(&ch->not_empty, &ch->mutex);
    }
    
    int64_t value = 0;
    if (ch->count > 0) {
        value = ch->buffer[ch->read_pos];
        ch->read_pos = (ch->read_pos + 1) % ch->capacity;
        ch->count--;
        pthread_cond_signal(&ch->not_full);
    }
    
    pthread_mutex_unlock(&ch->mutex);
    return value;
}

// Close the channel
void haira_channel_close(HairaChannel* ch) {
    pthread_mutex_lock(&ch->mutex);
    ch->closed = 1;
    pthread_cond_broadcast(&ch->not_empty);
    pthread_cond_broadcast(&ch->not_full);
    pthread_mutex_unlock(&ch->mutex);
}

// Check if channel has data available (non-blocking)
int64_t haira_channel_has_data(HairaChannel* ch) {
    pthread_mutex_lock(&ch->mutex);
    int64_t has = ch->count > 0 ? 1 : 0;
    pthread_mutex_unlock(&ch->mutex);
    return has;
}

// Check if channel is closed
int64_t haira_channel_is_closed(HairaChannel* ch) {
    pthread_mutex_lock(&ch->mutex);
    int64_t closed = ch->closed ? 1 : 0;
    pthread_mutex_unlock(&ch->mutex);
    return closed;
}
