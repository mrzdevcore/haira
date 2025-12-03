//! Regular expression functions

use crate::strings::HairaString;
use regex::Regex;
use std::collections::HashMap;
use std::sync::Mutex;

// Cache compiled regexes for performance
lazy_static::lazy_static! {
    static ref REGEX_CACHE: Mutex<HashMap<String, Regex>> = Mutex::new(HashMap::new());
}

fn get_or_compile_regex(pattern: &str) -> Option<Regex> {
    let mut cache = REGEX_CACHE.lock().ok()?;
    if let Some(re) = cache.get(pattern) {
        return Some(re.clone());
    }
    if let Ok(re) = Regex::new(pattern) {
        cache.insert(pattern.to_string(), re.clone());
        Some(re)
    } else {
        None
    }
}

/// Check if string matches regex pattern
/// Returns 1 if matches, 0 if not
#[no_mangle]
pub extern "C" fn haira_regex_match(
    str_ptr: *const u8,
    str_len: i64,
    pattern_ptr: *const u8,
    pattern_len: i64,
) -> i64 {
    if str_ptr.is_null() || pattern_ptr.is_null() || str_len <= 0 || pattern_len <= 0 {
        return 0;
    }

    let str_slice = unsafe { std::slice::from_raw_parts(str_ptr, str_len as usize) };
    let pattern_slice = unsafe { std::slice::from_raw_parts(pattern_ptr, pattern_len as usize) };

    let s = match std::str::from_utf8(str_slice) {
        Ok(s) => s,
        Err(_) => return 0,
    };
    let pattern = match std::str::from_utf8(pattern_slice) {
        Ok(p) => p,
        Err(_) => return 0,
    };

    match get_or_compile_regex(pattern) {
        Some(re) => re.is_match(s) as i64,
        None => 0,
    }
}

/// Find first match of regex in string
/// Returns HairaString* of the match, or empty string if no match
#[no_mangle]
pub extern "C" fn haira_regex_find(
    str_ptr: *const u8,
    str_len: i64,
    pattern_ptr: *const u8,
    pattern_len: i64,
) -> *mut HairaString {
    if str_ptr.is_null() || pattern_ptr.is_null() || str_len <= 0 || pattern_len <= 0 {
        return HairaString::empty();
    }

    let str_slice = unsafe { std::slice::from_raw_parts(str_ptr, str_len as usize) };
    let pattern_slice = unsafe { std::slice::from_raw_parts(pattern_ptr, pattern_len as usize) };

    let s = match std::str::from_utf8(str_slice) {
        Ok(s) => s,
        Err(_) => return HairaString::empty(),
    };
    let pattern = match std::str::from_utf8(pattern_slice) {
        Ok(p) => p,
        Err(_) => return HairaString::empty(),
    };

    match get_or_compile_regex(pattern) {
        Some(re) => match re.find(s) {
            Some(m) => HairaString::new(m.as_str().as_bytes()),
            None => HairaString::empty(),
        },
        None => HairaString::empty(),
    }
}

/// Replace first occurrence of regex match with replacement
#[no_mangle]
pub extern "C" fn haira_regex_replace(
    str_ptr: *const u8,
    str_len: i64,
    pattern_ptr: *const u8,
    pattern_len: i64,
    replacement_ptr: *const u8,
    replacement_len: i64,
) -> *mut HairaString {
    if str_ptr.is_null() || pattern_ptr.is_null() || str_len <= 0 || pattern_len <= 0 {
        return HairaString::empty();
    }

    let str_slice = unsafe { std::slice::from_raw_parts(str_ptr, str_len as usize) };
    let pattern_slice = unsafe { std::slice::from_raw_parts(pattern_ptr, pattern_len as usize) };
    let replacement_slice = if replacement_ptr.is_null() || replacement_len <= 0 {
        &[]
    } else {
        unsafe { std::slice::from_raw_parts(replacement_ptr, replacement_len as usize) }
    };

    let s = match std::str::from_utf8(str_slice) {
        Ok(s) => s,
        Err(_) => return HairaString::empty(),
    };
    let pattern = match std::str::from_utf8(pattern_slice) {
        Ok(p) => p,
        Err(_) => return HairaString::empty(),
    };
    let replacement = match std::str::from_utf8(replacement_slice) {
        Ok(r) => r,
        Err(_) => "",
    };

    match get_or_compile_regex(pattern) {
        Some(re) => {
            let result = re.replace(s, replacement);
            HairaString::new(result.as_bytes())
        }
        None => HairaString::new(str_slice),
    }
}

/// Replace all occurrences of regex match with replacement
#[no_mangle]
pub extern "C" fn haira_regex_replace_all(
    str_ptr: *const u8,
    str_len: i64,
    pattern_ptr: *const u8,
    pattern_len: i64,
    replacement_ptr: *const u8,
    replacement_len: i64,
) -> *mut HairaString {
    if str_ptr.is_null() || pattern_ptr.is_null() || str_len <= 0 || pattern_len <= 0 {
        return HairaString::empty();
    }

    let str_slice = unsafe { std::slice::from_raw_parts(str_ptr, str_len as usize) };
    let pattern_slice = unsafe { std::slice::from_raw_parts(pattern_ptr, pattern_len as usize) };
    let replacement_slice = if replacement_ptr.is_null() || replacement_len <= 0 {
        &[]
    } else {
        unsafe { std::slice::from_raw_parts(replacement_ptr, replacement_len as usize) }
    };

    let s = match std::str::from_utf8(str_slice) {
        Ok(s) => s,
        Err(_) => return HairaString::empty(),
    };
    let pattern = match std::str::from_utf8(pattern_slice) {
        Ok(p) => p,
        Err(_) => return HairaString::empty(),
    };
    let replacement = match std::str::from_utf8(replacement_slice) {
        Ok(r) => r,
        Err(_) => "",
    };

    match get_or_compile_regex(pattern) {
        Some(re) => {
            let result = re.replace_all(s, replacement);
            HairaString::new(result.as_bytes())
        }
        None => HairaString::new(str_slice),
    }
}

/// Count number of regex matches in string
#[no_mangle]
pub extern "C" fn haira_regex_count(
    str_ptr: *const u8,
    str_len: i64,
    pattern_ptr: *const u8,
    pattern_len: i64,
) -> i64 {
    if str_ptr.is_null() || pattern_ptr.is_null() || str_len <= 0 || pattern_len <= 0 {
        return 0;
    }

    let str_slice = unsafe { std::slice::from_raw_parts(str_ptr, str_len as usize) };
    let pattern_slice = unsafe { std::slice::from_raw_parts(pattern_ptr, pattern_len as usize) };

    let s = match std::str::from_utf8(str_slice) {
        Ok(s) => s,
        Err(_) => return 0,
    };
    let pattern = match std::str::from_utf8(pattern_slice) {
        Ok(p) => p,
        Err(_) => return 0,
    };

    match get_or_compile_regex(pattern) {
        Some(re) => re.find_iter(s).count() as i64,
        None => 0,
    }
}
