//! Concurrency primitives - threads and channels

use std::collections::VecDeque;
use std::sync::{Condvar, Mutex};
use std::thread;

/// Channel implementation
#[repr(C)]
pub struct HairaChannel {
    inner: *mut ChannelInner,
}

struct ChannelInner {
    buffer: Mutex<ChannelBuffer>,
    not_empty: Condvar,
    not_full: Condvar,
}

struct ChannelBuffer {
    queue: VecDeque<i64>,
    capacity: usize,
    closed: bool,
}

/// Create a new channel with given capacity
#[no_mangle]
pub extern "C" fn haira_channel_new(capacity: i64) -> *mut HairaChannel {
    let cap = if capacity <= 0 { 1 } else { capacity as usize };

    let inner = Box::new(ChannelInner {
        buffer: Mutex::new(ChannelBuffer {
            queue: VecDeque::with_capacity(cap),
            capacity: cap,
            closed: false,
        }),
        not_empty: Condvar::new(),
        not_full: Condvar::new(),
    });

    let channel = Box::new(HairaChannel {
        inner: Box::into_raw(inner),
    });

    Box::into_raw(channel)
}

/// Send a value to the channel (blocks if full)
#[no_mangle]
pub extern "C" fn haira_channel_send(ch: *mut HairaChannel, value: i64) {
    if ch.is_null() {
        return;
    }

    let channel = unsafe { &*ch };
    let inner = unsafe { &*channel.inner };

    let mut buffer = inner.buffer.lock().unwrap();

    // Wait until there's room or channel is closed
    while buffer.queue.len() >= buffer.capacity && !buffer.closed {
        buffer = inner.not_full.wait(buffer).unwrap();
    }

    if !buffer.closed {
        buffer.queue.push_back(value);
        inner.not_empty.notify_one();
    }
}

/// Receive a value from the channel (blocks if empty)
#[no_mangle]
pub extern "C" fn haira_channel_receive(ch: *mut HairaChannel) -> i64 {
    if ch.is_null() {
        return 0;
    }

    let channel = unsafe { &*ch };
    let inner = unsafe { &*channel.inner };

    let mut buffer = inner.buffer.lock().unwrap();

    // Wait until there's data or channel is closed
    while buffer.queue.is_empty() && !buffer.closed {
        buffer = inner.not_empty.wait(buffer).unwrap();
    }

    if let Some(value) = buffer.queue.pop_front() {
        inner.not_full.notify_one();
        value
    } else {
        0 // Channel closed and empty
    }
}

/// Close the channel
#[no_mangle]
pub extern "C" fn haira_channel_close(ch: *mut HairaChannel) {
    if ch.is_null() {
        return;
    }

    let channel = unsafe { &*ch };
    let inner = unsafe { &*channel.inner };

    let mut buffer = inner.buffer.lock().unwrap();
    buffer.closed = true;

    // Wake up all waiting threads
    inner.not_empty.notify_all();
    inner.not_full.notify_all();
}

/// Check if channel has data available (non-blocking)
#[no_mangle]
pub extern "C" fn haira_channel_has_data(ch: *mut HairaChannel) -> i64 {
    if ch.is_null() {
        return 0;
    }

    let channel = unsafe { &*ch };
    let inner = unsafe { &*channel.inner };

    let buffer = inner.buffer.lock().unwrap();
    (!buffer.queue.is_empty()) as i64
}

/// Check if channel is closed
#[no_mangle]
pub extern "C" fn haira_channel_is_closed(ch: *mut HairaChannel) -> i64 {
    if ch.is_null() {
        return 1;
    }

    let channel = unsafe { &*ch };
    let inner = unsafe { &*channel.inner };

    let buffer = inner.buffer.lock().unwrap();
    buffer.closed as i64
}

// Thread functions

/// Spawn a new thread running the given function (fire-and-forget)
#[no_mangle]
pub extern "C" fn haira_spawn(func: extern "C" fn()) -> i64 {
    let handle = thread::spawn(move || {
        func();
    });

    // Detach - we don't track this handle
    drop(handle);
    1 // Success
}

/// Spawn a new thread that can be joined
#[no_mangle]
pub extern "C" fn haira_spawn_joinable(func: extern "C" fn()) -> i64 {
    let handle = thread::spawn(move || {
        func();
    });

    // Store the handle
    let boxed = Box::new(handle);
    Box::into_raw(boxed) as i64
}

/// Wait for a joinable thread to complete
#[no_mangle]
pub extern "C" fn haira_thread_join(handle: i64) {
    if handle == 0 {
        return;
    }

    let boxed: Box<thread::JoinHandle<()>> =
        unsafe { Box::from_raw(handle as *mut thread::JoinHandle<()>) };

    let _ = boxed.join();
}
