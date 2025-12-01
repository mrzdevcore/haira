//! Time functions

use std::sync::OnceLock;
use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

static START_TIME: OnceLock<Instant> = OnceLock::new();

fn get_start_time() -> &'static Instant {
    START_TIME.get_or_init(Instant::now)
}

/// Get current time in milliseconds since epoch
#[no_mangle]
pub extern "C" fn haira_time_now() -> i64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .unwrap_or(Duration::ZERO)
        .as_millis() as i64
}

/// Get monotonic time in nanoseconds (for measuring durations)
#[no_mangle]
pub extern "C" fn haira_time_monotonic() -> i64 {
    get_start_time().elapsed().as_nanos() as i64
}

/// Sleep for milliseconds
#[no_mangle]
pub extern "C" fn haira_sleep(ms: i64) {
    if ms > 0 {
        std::thread::sleep(Duration::from_millis(ms as u64));
    }
}
