//! Memory management functions

use std::alloc::{alloc, dealloc, realloc, Layout};

/// Allocate memory
#[no_mangle]
pub extern "C" fn haira_alloc(size: i64) -> *mut u8 {
    if size <= 0 {
        return std::ptr::null_mut();
    }
    unsafe {
        let layout = Layout::from_size_align_unchecked(size as usize, 8);
        alloc(layout)
    }
}

/// Reallocate memory
#[no_mangle]
pub extern "C" fn haira_realloc(ptr: *mut u8, new_size: i64) -> *mut u8 {
    if ptr.is_null() {
        return haira_alloc(new_size);
    }
    if new_size <= 0 {
        haira_free(ptr);
        return std::ptr::null_mut();
    }
    unsafe {
        // We don't know the original size, assume worst case
        let old_layout = Layout::from_size_align_unchecked(1, 8);
        let new_layout = Layout::from_size_align_unchecked(new_size as usize, 8);
        realloc(ptr, old_layout, new_layout.size())
    }
}

/// Free memory
#[no_mangle]
pub extern "C" fn haira_free(ptr: *mut u8) {
    if ptr.is_null() {
        return;
    }
    unsafe {
        // We use a minimal layout since we don't track sizes
        let layout = Layout::from_size_align_unchecked(1, 8);
        dealloc(ptr, layout);
    }
}
