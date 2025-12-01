//! Math functions

/// Absolute value (int)
#[no_mangle]
pub extern "C" fn haira_abs(x: i64) -> i64 {
    x.abs()
}

/// Absolute value (float)
#[no_mangle]
pub extern "C" fn haira_fabs(x: f64) -> f64 {
    x.abs()
}

/// Minimum of two ints
#[no_mangle]
pub extern "C" fn haira_min(a: i64, b: i64) -> i64 {
    a.min(b)
}

/// Maximum of two ints
#[no_mangle]
pub extern "C" fn haira_max(a: i64, b: i64) -> i64 {
    a.max(b)
}

/// Minimum of two floats
#[no_mangle]
pub extern "C" fn haira_fmin(a: f64, b: f64) -> f64 {
    a.min(b)
}

/// Maximum of two floats
#[no_mangle]
pub extern "C" fn haira_fmax(a: f64, b: f64) -> f64 {
    a.max(b)
}

/// Clamp int to range
#[no_mangle]
pub extern "C" fn haira_clamp(x: i64, min: i64, max: i64) -> i64 {
    x.clamp(min, max)
}

/// Floor
#[no_mangle]
pub extern "C" fn haira_floor(x: f64) -> f64 {
    x.floor()
}

/// Ceiling
#[no_mangle]
pub extern "C" fn haira_ceil(x: f64) -> f64 {
    x.ceil()
}

/// Round
#[no_mangle]
pub extern "C" fn haira_round(x: f64) -> f64 {
    x.round()
}

/// Truncate (towards zero)
#[no_mangle]
pub extern "C" fn haira_trunc(x: f64) -> f64 {
    x.trunc()
}

/// Power
#[no_mangle]
pub extern "C" fn haira_pow(base: f64, exp: f64) -> f64 {
    base.powf(exp)
}

/// Square root
#[no_mangle]
pub extern "C" fn haira_sqrt(x: f64) -> f64 {
    x.sqrt()
}

/// Natural log
#[no_mangle]
pub extern "C" fn haira_log(x: f64) -> f64 {
    x.ln()
}

/// Log base 10
#[no_mangle]
pub extern "C" fn haira_log10(x: f64) -> f64 {
    x.log10()
}

/// Exponential (e^x)
#[no_mangle]
pub extern "C" fn haira_exp(x: f64) -> f64 {
    x.exp()
}

/// Sine
#[no_mangle]
pub extern "C" fn haira_sin(x: f64) -> f64 {
    x.sin()
}

/// Cosine
#[no_mangle]
pub extern "C" fn haira_cos(x: f64) -> f64 {
    x.cos()
}

/// Tangent
#[no_mangle]
pub extern "C" fn haira_tan(x: f64) -> f64 {
    x.tan()
}

/// Arc sine
#[no_mangle]
pub extern "C" fn haira_asin(x: f64) -> f64 {
    x.asin()
}

/// Arc cosine
#[no_mangle]
pub extern "C" fn haira_acos(x: f64) -> f64 {
    x.acos()
}

/// Arc tangent
#[no_mangle]
pub extern "C" fn haira_atan(x: f64) -> f64 {
    x.atan()
}

/// Arc tangent of y/x
#[no_mangle]
pub extern "C" fn haira_atan2(y: f64, x: f64) -> f64 {
    y.atan2(x)
}

/// Random integer in range [0, max)
#[no_mangle]
pub extern "C" fn haira_random_int(max: i64) -> i64 {
    if max <= 0 {
        return 0;
    }
    // Simple LCG random - good enough for basic use
    use std::sync::atomic::{AtomicU64, Ordering};
    static SEED: AtomicU64 = AtomicU64::new(0);

    let mut seed = SEED.load(Ordering::Relaxed);
    if seed == 0 {
        // Initialize with time-based seed
        seed = std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .unwrap_or_default()
            .as_nanos() as u64;
    }

    // LCG parameters (same as glibc)
    seed = seed.wrapping_mul(1103515245).wrapping_add(12345);
    SEED.store(seed, Ordering::Relaxed);

    ((seed >> 16) as i64).abs() % max
}

/// Random float in range [0, 1)
#[no_mangle]
pub extern "C" fn haira_random_float() -> f64 {
    haira_random_int(i64::MAX) as f64 / i64::MAX as f64
}

/// Seed random number generator
#[no_mangle]
pub extern "C" fn haira_random_seed(seed: i64) {
    use std::sync::atomic::{AtomicU64, Ordering};
    static SEED: AtomicU64 = AtomicU64::new(0);
    SEED.store(seed as u64, Ordering::Relaxed);
}
