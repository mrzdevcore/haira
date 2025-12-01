//! Environment functions

use crate::strings::HairaString;
use std::ptr;

/// Get environment variable (returns NULL if not set)
#[no_mangle]
pub extern "C" fn haira_env_get(name: *const u8, name_len: i64) -> *mut HairaString {
    if name.is_null() || name_len <= 0 {
        return ptr::null_mut();
    }

    let name_slice = unsafe { std::slice::from_raw_parts(name, name_len as usize) };
    let name_str = match std::str::from_utf8(name_slice) {
        Ok(s) => s,
        Err(_) => return ptr::null_mut(),
    };

    match std::env::var(name_str) {
        Ok(value) => {
            let bytes = value.as_bytes();
            let len = bytes.len() as i64;
            let cap = len + 1;

            let data = unsafe {
                let ptr = libc::malloc(cap as usize) as *mut u8;
                if !ptr.is_null() {
                    ptr::copy_nonoverlapping(bytes.as_ptr(), ptr, bytes.len());
                    *ptr.add(bytes.len()) = 0;
                }
                ptr
            };

            let result = Box::new(HairaString { data, len, cap });
            Box::into_raw(result)
        }
        Err(_) => ptr::null_mut(),
    }
}

/// Exit program with code
#[no_mangle]
pub extern "C" fn haira_exit(code: i64) {
    std::process::exit(code as i32);
}

// File I/O functions

use crate::error::haira_set_error;
use std::fs::{self, File, OpenOptions};
use std::io::Write;

/// Read entire file to string (returns NULL on error)
#[no_mangle]
pub extern "C" fn haira_file_read(path: *const u8, path_len: i64) -> *mut HairaString {
    if path.is_null() || path_len <= 0 {
        haira_set_error(1);
        return ptr::null_mut();
    }

    let path_slice = unsafe { std::slice::from_raw_parts(path, path_len as usize) };
    let path_str = match std::str::from_utf8(path_slice) {
        Ok(s) => s,
        Err(_) => {
            haira_set_error(1);
            return ptr::null_mut();
        }
    };

    match fs::read(path_str) {
        Ok(contents) => {
            let len = contents.len() as i64;
            let cap = len + 1;

            let data = unsafe {
                let ptr = libc::malloc(cap as usize) as *mut u8;
                if !ptr.is_null() {
                    ptr::copy_nonoverlapping(contents.as_ptr(), ptr, contents.len());
                    *ptr.add(contents.len()) = 0;
                }
                ptr
            };

            let result = Box::new(HairaString { data, len, cap });
            Box::into_raw(result)
        }
        Err(_) => {
            haira_set_error(1);
            ptr::null_mut()
        }
    }
}

/// Write string to file (returns 0 on success, non-zero on error)
#[no_mangle]
pub extern "C" fn haira_file_write(
    path: *const u8,
    path_len: i64,
    content: *const u8,
    content_len: i64,
) -> i64 {
    if path.is_null() || path_len <= 0 {
        return 1;
    }

    let path_slice = unsafe { std::slice::from_raw_parts(path, path_len as usize) };
    let path_str = match std::str::from_utf8(path_slice) {
        Ok(s) => s,
        Err(_) => return 1,
    };

    let content_slice = if content.is_null() || content_len <= 0 {
        &[]
    } else {
        unsafe { std::slice::from_raw_parts(content, content_len as usize) }
    };

    match File::create(path_str) {
        Ok(mut file) => match file.write_all(content_slice) {
            Ok(_) => 0,
            Err(_) => 2,
        },
        Err(_) => 1,
    }
}

/// Append string to file (returns 0 on success, non-zero on error)
#[no_mangle]
pub extern "C" fn haira_file_append(
    path: *const u8,
    path_len: i64,
    content: *const u8,
    content_len: i64,
) -> i64 {
    if path.is_null() || path_len <= 0 {
        return 1;
    }

    let path_slice = unsafe { std::slice::from_raw_parts(path, path_len as usize) };
    let path_str = match std::str::from_utf8(path_slice) {
        Ok(s) => s,
        Err(_) => return 1,
    };

    let content_slice = if content.is_null() || content_len <= 0 {
        &[]
    } else {
        unsafe { std::slice::from_raw_parts(content, content_len as usize) }
    };

    match OpenOptions::new().append(true).create(true).open(path_str) {
        Ok(mut file) => match file.write_all(content_slice) {
            Ok(_) => 0,
            Err(_) => 2,
        },
        Err(_) => 1,
    }
}

/// Check if file exists
#[no_mangle]
pub extern "C" fn haira_file_exists(path: *const u8, path_len: i64) -> i64 {
    if path.is_null() || path_len <= 0 {
        return 0;
    }

    let path_slice = unsafe { std::slice::from_raw_parts(path, path_len as usize) };
    let path_str = match std::str::from_utf8(path_slice) {
        Ok(s) => s,
        Err(_) => return 0,
    };

    std::path::Path::new(path_str).exists() as i64
}
