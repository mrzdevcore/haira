//! I/O functions

use std::io::{self, Write};

/// Print a string (pointer + length)
#[no_mangle]
pub extern "C" fn haira_print(ptr: *const u8, len: i64) {
    if ptr.is_null() || len <= 0 {
        return;
    }
    let slice = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let _ = io::stdout().write_all(slice);
}

/// Print an integer
#[no_mangle]
pub extern "C" fn haira_print_int(value: i64) {
    print!("{}", value);
}

/// Print a float
#[no_mangle]
pub extern "C" fn haira_print_float(value: f64) {
    print!("{}", value);
}

/// Print a boolean
#[no_mangle]
pub extern "C" fn haira_print_bool(value: i8) {
    print!("{}", if value != 0 { "true" } else { "false" });
}

/// Print a newline and flush
#[no_mangle]
pub extern "C" fn haira_println() {
    println!();
    let _ = io::stdout().flush();
}

/// Panic with message
#[no_mangle]
pub extern "C" fn haira_panic(msg: *const u8, len: i64) {
    let message = if msg.is_null() || len <= 0 {
        "unknown error".to_string()
    } else {
        let slice = unsafe { std::slice::from_raw_parts(msg, len as usize) };
        String::from_utf8_lossy(slice).to_string()
    };
    eprintln!("panic: {}", message);
    std::process::exit(1);
}
