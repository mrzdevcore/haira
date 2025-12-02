//! Testing framework for Haira
//!
//! Provides assertion functions and test result tracking for unit testing.

use std::cell::Cell;
use std::io::Write;
use std::sync::atomic::{AtomicI64, Ordering};

// Test statistics
static TESTS_RUN: AtomicI64 = AtomicI64::new(0);
static TESTS_PASSED: AtomicI64 = AtomicI64::new(0);
static TESTS_FAILED: AtomicI64 = AtomicI64::new(0);

thread_local! {
    static CURRENT_TEST_NAME: Cell<Option<String>> = const { Cell::new(None) };
}

/// Start a new test with the given name
#[no_mangle]
pub extern "C" fn haira_test_start(name_ptr: *const u8, name_len: i64) {
    let name = if name_ptr.is_null() || name_len <= 0 {
        "unnamed test".to_string()
    } else {
        let slice = unsafe { std::slice::from_raw_parts(name_ptr, name_len as usize) };
        String::from_utf8_lossy(slice).to_string()
    };

    TESTS_RUN.fetch_add(1, Ordering::SeqCst);
    CURRENT_TEST_NAME.with(|n| n.set(Some(name.clone())));
    print!("  test {} ... ", name);
    let _ = std::io::stdout().flush();
}

/// End the current test (called implicitly if test passes)
#[no_mangle]
pub extern "C" fn haira_test_pass() {
    TESTS_PASSED.fetch_add(1, Ordering::SeqCst);
    println!("\x1b[32mok\x1b[0m");
    CURRENT_TEST_NAME.with(|n| n.set(None));
}

/// Mark the current test as failed with a message
#[no_mangle]
pub extern "C" fn haira_test_fail(msg_ptr: *const u8, msg_len: i64) {
    TESTS_FAILED.fetch_add(1, Ordering::SeqCst);

    let msg = if msg_ptr.is_null() || msg_len <= 0 {
        "assertion failed".to_string()
    } else {
        let slice = unsafe { std::slice::from_raw_parts(msg_ptr, msg_len as usize) };
        String::from_utf8_lossy(slice).to_string()
    };

    println!("\x1b[31mFAILED\x1b[0m");
    eprintln!("    {}", msg);
    CURRENT_TEST_NAME.with(|n| n.set(None));
}

/// Assert that a condition is true
#[no_mangle]
pub extern "C" fn haira_assert(condition: i64) -> i64 {
    if condition != 0 {
        1 // success
    } else {
        haira_test_fail(
            "assertion failed: expected true".as_ptr(),
            "assertion failed: expected true".len() as i64,
        );
        0 // failure
    }
}

/// Assert that a condition is true with a custom message
#[no_mangle]
pub extern "C" fn haira_assert_msg(condition: i64, msg_ptr: *const u8, msg_len: i64) -> i64 {
    if condition != 0 {
        1 // success
    } else {
        haira_test_fail(msg_ptr, msg_len);
        0 // failure
    }
}

/// Assert that two integers are equal
#[no_mangle]
pub extern "C" fn haira_assert_eq(expected: i64, actual: i64) -> i64 {
    if expected == actual {
        1 // success
    } else {
        let msg = format!("assertion failed: expected {}, got {}", expected, actual);
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0 // failure
    }
}

/// Assert that two integers are not equal
#[no_mangle]
pub extern "C" fn haira_assert_ne(a: i64, b: i64) -> i64 {
    if a != b {
        1 // success
    } else {
        let msg = format!("assertion failed: {} should not equal {}", a, b);
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0 // failure
    }
}

/// Assert that two floats are approximately equal (within epsilon)
#[no_mangle]
pub extern "C" fn haira_assert_float_eq(expected: f64, actual: f64, epsilon: f64) -> i64 {
    if (expected - actual).abs() <= epsilon {
        1 // success
    } else {
        let msg = format!(
            "assertion failed: expected {} (±{}), got {}",
            expected, epsilon, actual
        );
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0 // failure
    }
}

/// Assert that two strings are equal
#[no_mangle]
pub extern "C" fn haira_assert_str_eq(
    expected_ptr: *const u8,
    expected_len: i64,
    actual_ptr: *const u8,
    actual_len: i64,
) -> i64 {
    let expected = if expected_ptr.is_null() || expected_len <= 0 {
        ""
    } else {
        unsafe {
            std::str::from_utf8_unchecked(std::slice::from_raw_parts(
                expected_ptr,
                expected_len as usize,
            ))
        }
    };

    let actual = if actual_ptr.is_null() || actual_len <= 0 {
        ""
    } else {
        unsafe {
            std::str::from_utf8_unchecked(std::slice::from_raw_parts(
                actual_ptr,
                actual_len as usize,
            ))
        }
    };

    if expected == actual {
        1 // success
    } else {
        let msg = format!(
            "assertion failed: expected \"{}\", got \"{}\"",
            expected, actual
        );
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0 // failure
    }
}

/// Assert that a value is greater than another
#[no_mangle]
pub extern "C" fn haira_assert_gt(a: i64, b: i64) -> i64 {
    if a > b {
        1
    } else {
        let msg = format!("assertion failed: {} is not greater than {}", a, b);
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0
    }
}

/// Assert that a value is greater than or equal to another
#[no_mangle]
pub extern "C" fn haira_assert_ge(a: i64, b: i64) -> i64 {
    if a >= b {
        1
    } else {
        let msg = format!("assertion failed: {} is not >= {}", a, b);
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0
    }
}

/// Assert that a value is less than another
#[no_mangle]
pub extern "C" fn haira_assert_lt(a: i64, b: i64) -> i64 {
    if a < b {
        1
    } else {
        let msg = format!("assertion failed: {} is not less than {}", a, b);
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0
    }
}

/// Assert that a value is less than or equal to another
#[no_mangle]
pub extern "C" fn haira_assert_le(a: i64, b: i64) -> i64 {
    if a <= b {
        1
    } else {
        let msg = format!("assertion failed: {} is not <= {}", a, b);
        haira_test_fail(msg.as_ptr(), msg.len() as i64);
        0
    }
}

/// Print test summary and return exit code (0 if all passed, 1 if any failed)
#[no_mangle]
pub extern "C" fn haira_test_summary() -> i64 {
    let run = TESTS_RUN.load(Ordering::SeqCst);
    let passed = TESTS_PASSED.load(Ordering::SeqCst);
    let failed = TESTS_FAILED.load(Ordering::SeqCst);

    println!();
    if failed == 0 {
        println!(
            "\x1b[32mtest result: ok\x1b[0m. {} passed; {} failed; {} total",
            passed, failed, run
        );
        0
    } else {
        println!(
            "\x1b[31mtest result: FAILED\x1b[0m. {} passed; {} failed; {} total",
            passed, failed, run
        );
        1
    }
}

/// Reset test counters (useful for running multiple test suites)
#[no_mangle]
pub extern "C" fn haira_test_reset() {
    TESTS_RUN.store(0, Ordering::SeqCst);
    TESTS_PASSED.store(0, Ordering::SeqCst);
    TESTS_FAILED.store(0, Ordering::SeqCst);
}

/// Get the number of tests run
#[no_mangle]
pub extern "C" fn haira_test_count() -> i64 {
    TESTS_RUN.load(Ordering::SeqCst)
}

/// Get the number of tests passed
#[no_mangle]
pub extern "C" fn haira_test_passed() -> i64 {
    TESTS_PASSED.load(Ordering::SeqCst)
}

/// Get the number of tests failed
#[no_mangle]
pub extern "C" fn haira_test_failed() -> i64 {
    TESTS_FAILED.load(Ordering::SeqCst)
}

/// Print a test section header
#[no_mangle]
pub extern "C" fn haira_test_section(name_ptr: *const u8, name_len: i64) {
    let name = if name_ptr.is_null() || name_len <= 0 {
        "Tests".to_string()
    } else {
        let slice = unsafe { std::slice::from_raw_parts(name_ptr, name_len as usize) };
        String::from_utf8_lossy(slice).to_string()
    };

    println!();
    println!("running {} tests", name);
}
