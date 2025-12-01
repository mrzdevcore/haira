//! Error handling with thread-local state

use std::cell::Cell;

thread_local! {
    static CURRENT_ERROR: Cell<i64> = const { Cell::new(0) };
}

/// Set current error
#[no_mangle]
pub extern "C" fn haira_set_error(error: i64) {
    CURRENT_ERROR.with(|e| e.set(error));
}

/// Get and clear current error
#[no_mangle]
pub extern "C" fn haira_get_error() -> i64 {
    CURRENT_ERROR.with(|e| {
        let err = e.get();
        e.set(0);
        err
    })
}

/// Check if there's an error
#[no_mangle]
pub extern "C" fn haira_has_error() -> i64 {
    CURRENT_ERROR.with(|e| if e.get() != 0 { 1 } else { 0 })
}

/// Clear error
#[no_mangle]
pub extern "C" fn haira_clear_error() {
    CURRENT_ERROR.with(|e| e.set(0));
}
