//! String operations

use std::ptr;

/// HairaString - the runtime string representation
#[repr(C)]
pub struct HairaString {
    pub data: *mut u8,
    pub len: i64,
    pub cap: i64,
}

impl HairaString {
    pub fn new(s: &[u8]) -> *mut HairaString {
        let len = s.len() as i64;
        let cap = len + 1;
        let data = unsafe {
            let ptr = libc::malloc(cap as usize) as *mut u8;
            if !ptr.is_null() {
                ptr::copy_nonoverlapping(s.as_ptr(), ptr, s.len());
                *ptr.add(s.len()) = 0; // null terminator
            }
            ptr
        };

        let result = Box::new(HairaString { data, len, cap });
        Box::into_raw(result)
    }

    pub fn empty() -> *mut HairaString {
        let data = unsafe {
            let ptr = libc::malloc(1) as *mut u8;
            if !ptr.is_null() {
                *ptr = 0;
            }
            ptr
        };
        let result = Box::new(HairaString {
            data,
            len: 0,
            cap: 1,
        });
        Box::into_raw(result)
    }
}

/// Create a HairaString from a static string pointer and length
/// This wraps a static string in a HairaString struct for consistent handling
#[no_mangle]
pub extern "C" fn haira_string_from_static(ptr: *const u8, len: i64) -> *mut HairaString {
    if ptr.is_null() || len <= 0 {
        return HairaString::empty();
    }
    let slice = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    HairaString::new(slice)
}

/// String concatenation
#[no_mangle]
pub extern "C" fn haira_string_concat(
    a: *const u8,
    alen: i64,
    b: *const u8,
    blen: i64,
) -> *mut HairaString {
    let a_slice = if a.is_null() || alen <= 0 {
        &[]
    } else {
        unsafe { std::slice::from_raw_parts(a, alen as usize) }
    };

    let b_slice = if b.is_null() || blen <= 0 {
        &[]
    } else {
        unsafe { std::slice::from_raw_parts(b, blen as usize) }
    };

    let mut result = Vec::with_capacity(a_slice.len() + b_slice.len());
    result.extend_from_slice(a_slice);
    result.extend_from_slice(b_slice);

    HairaString::new(&result)
}

/// Integer to string
#[no_mangle]
pub extern "C" fn haira_int_to_string(value: i64) -> *mut HairaString {
    let s = value.to_string();
    HairaString::new(s.as_bytes())
}

/// Float to string
#[no_mangle]
pub extern "C" fn haira_float_to_string(value: f64) -> *mut HairaString {
    let s = value.to_string();
    HairaString::new(s.as_bytes())
}

/// Get string length
#[no_mangle]
pub extern "C" fn haira_string_len(_ptr: *const u8, len: i64) -> i64 {
    len
}

/// Check if string is empty
#[no_mangle]
pub extern "C" fn haira_string_is_empty(_ptr: *const u8, len: i64) -> i64 {
    if len == 0 {
        1
    } else {
        0
    }
}

/// Convert string to uppercase
#[no_mangle]
pub extern "C" fn haira_string_upper(ptr: *const u8, len: i64) -> *mut HairaString {
    if ptr.is_null() || len <= 0 {
        return HairaString::empty();
    }
    let slice = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let upper: Vec<u8> = slice.iter().map(|&c| c.to_ascii_uppercase()).collect();
    HairaString::new(&upper)
}

/// Convert string to lowercase
#[no_mangle]
pub extern "C" fn haira_string_lower(ptr: *const u8, len: i64) -> *mut HairaString {
    if ptr.is_null() || len <= 0 {
        return HairaString::empty();
    }
    let slice = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let lower: Vec<u8> = slice.iter().map(|&c| c.to_ascii_lowercase()).collect();
    HairaString::new(&lower)
}

/// Trim whitespace from both ends
#[no_mangle]
pub extern "C" fn haira_string_trim(ptr: *const u8, len: i64) -> *mut HairaString {
    if ptr.is_null() || len <= 0 {
        return HairaString::empty();
    }
    let slice = unsafe { std::slice::from_raw_parts(ptr, len as usize) };

    // Find start
    let start = slice
        .iter()
        .position(|&c| !c.is_ascii_whitespace())
        .unwrap_or(slice.len());
    // Find end
    let end = slice
        .iter()
        .rposition(|&c| !c.is_ascii_whitespace())
        .map(|i| i + 1)
        .unwrap_or(start);

    HairaString::new(&slice[start..end])
}

/// Get substring (start inclusive, end exclusive)
#[no_mangle]
pub extern "C" fn haira_string_slice(
    ptr: *const u8,
    len: i64,
    mut start: i64,
    mut end: i64,
) -> *mut HairaString {
    if ptr.is_null() || len <= 0 {
        return HairaString::empty();
    }

    // Handle negative indices
    if start < 0 {
        start = len + start;
    }
    if end < 0 {
        end = len + end;
    }

    // Clamp to valid range
    if start < 0 {
        start = 0;
    }
    if end > len {
        end = len;
    }
    if start > end {
        start = end;
    }

    let slice = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    HairaString::new(&slice[start as usize..end as usize])
}

/// Check if string contains substring
#[no_mangle]
pub extern "C" fn haira_string_contains(
    ptr: *const u8,
    len: i64,
    needle: *const u8,
    needle_len: i64,
) -> i64 {
    if needle.is_null() || needle_len <= 0 {
        return 1; // Empty needle is always contained
    }
    if ptr.is_null() || len <= 0 || needle_len > len {
        return 0;
    }

    let haystack = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let needle = unsafe { std::slice::from_raw_parts(needle, needle_len as usize) };

    haystack.windows(needle.len()).any(|w| w == needle) as i64
}

/// Check if string starts with prefix
#[no_mangle]
pub extern "C" fn haira_string_starts_with(
    ptr: *const u8,
    len: i64,
    prefix: *const u8,
    prefix_len: i64,
) -> i64 {
    if prefix.is_null() || prefix_len <= 0 {
        return 1;
    }
    if ptr.is_null() || len < prefix_len {
        return 0;
    }

    let s = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let p = unsafe { std::slice::from_raw_parts(prefix, prefix_len as usize) };

    s.starts_with(p) as i64
}

/// Check if string ends with suffix
#[no_mangle]
pub extern "C" fn haira_string_ends_with(
    ptr: *const u8,
    len: i64,
    suffix: *const u8,
    suffix_len: i64,
) -> i64 {
    if suffix.is_null() || suffix_len <= 0 {
        return 1;
    }
    if ptr.is_null() || len < suffix_len {
        return 0;
    }

    let s = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let suf = unsafe { std::slice::from_raw_parts(suffix, suffix_len as usize) };

    s.ends_with(suf) as i64
}

/// Find index of substring (-1 if not found)
#[no_mangle]
pub extern "C" fn haira_string_index_of(
    ptr: *const u8,
    len: i64,
    needle: *const u8,
    needle_len: i64,
) -> i64 {
    if needle.is_null() || needle_len <= 0 {
        return 0;
    }
    if ptr.is_null() || len <= 0 || needle_len > len {
        return -1;
    }

    let haystack = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let needle = unsafe { std::slice::from_raw_parts(needle, needle_len as usize) };

    haystack
        .windows(needle.len())
        .position(|w| w == needle)
        .map(|i| i as i64)
        .unwrap_or(-1)
}

/// Replace all occurrences
#[no_mangle]
pub extern "C" fn haira_string_replace(
    ptr: *const u8,
    len: i64,
    old: *const u8,
    old_len: i64,
    new: *const u8,
    new_len: i64,
) -> *mut HairaString {
    if ptr.is_null() || len <= 0 {
        return HairaString::empty();
    }
    if old.is_null() || old_len <= 0 {
        // Return copy of original
        let s = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
        return HairaString::new(s);
    }

    let s = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let old_s = unsafe { std::slice::from_raw_parts(old, old_len as usize) };
    let new_s = if new.is_null() || new_len <= 0 {
        &[]
    } else {
        unsafe { std::slice::from_raw_parts(new, new_len as usize) }
    };

    // Simple replacement algorithm
    let mut result = Vec::new();
    let mut i = 0;
    while i < s.len() {
        if i + old_s.len() <= s.len() && &s[i..i + old_s.len()] == old_s {
            result.extend_from_slice(new_s);
            i += old_s.len();
        } else {
            result.push(s[i]);
            i += 1;
        }
    }

    HairaString::new(&result)
}

/// Repeat string n times
#[no_mangle]
pub extern "C" fn haira_string_repeat(ptr: *const u8, len: i64, n: i64) -> *mut HairaString {
    if ptr.is_null() || len <= 0 || n <= 0 {
        return HairaString::empty();
    }

    let s = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let mut result = Vec::with_capacity((len * n) as usize);
    for _ in 0..n {
        result.extend_from_slice(s);
    }

    HairaString::new(&result)
}

/// Reverse string
#[no_mangle]
pub extern "C" fn haira_string_reverse(ptr: *const u8, len: i64) -> *mut HairaString {
    if ptr.is_null() || len <= 0 {
        return HairaString::empty();
    }

    let s = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    let reversed: Vec<u8> = s.iter().rev().copied().collect();

    HairaString::new(&reversed)
}

/// Get character at index (-1 if out of bounds)
#[no_mangle]
pub extern "C" fn haira_string_char_at(ptr: *const u8, len: i64, mut index: i64) -> i64 {
    if ptr.is_null() || len <= 0 {
        return -1;
    }

    // Handle negative index
    if index < 0 {
        index = len + index;
    }
    if index < 0 || index >= len {
        return -1;
    }

    let s = unsafe { std::slice::from_raw_parts(ptr, len as usize) };
    s[index as usize] as i64
}
