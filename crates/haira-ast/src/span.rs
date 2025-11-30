//! Source location tracking for AST nodes.

/// A span represents a range in the source code.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash, Default)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct Span {
    /// Start byte offset (inclusive)
    pub start: u32,
    /// End byte offset (exclusive)
    pub end: u32,
}

impl Span {
    /// Create a new span from start and end offsets.
    pub fn new(start: u32, end: u32) -> Self {
        Self { start, end }
    }

    /// Create an empty span at a position.
    pub fn empty(pos: u32) -> Self {
        Self {
            start: pos,
            end: pos,
        }
    }

    /// Create a span covering two spans.
    pub fn merge(self, other: Span) -> Self {
        Self {
            start: self.start.min(other.start),
            end: self.end.max(other.end),
        }
    }

    /// Check if this span contains a byte offset.
    pub fn contains(&self, offset: u32) -> bool {
        self.start <= offset && offset < self.end
    }

    /// Get the length of this span in bytes.
    pub fn len(&self) -> u32 {
        self.end - self.start
    }

    /// Check if this span is empty.
    pub fn is_empty(&self) -> bool {
        self.start == self.end
    }
}

/// A value with an associated source span.
#[derive(Debug, Clone, PartialEq, Eq, Hash)]
#[cfg_attr(feature = "serde", derive(serde::Serialize, serde::Deserialize))]
pub struct Spanned<T> {
    pub node: T,
    pub span: Span,
}

impl<T> Spanned<T> {
    /// Create a new spanned value.
    pub fn new(node: T, span: Span) -> Self {
        Self { node, span }
    }

    /// Map the inner value.
    pub fn map<U>(self, f: impl FnOnce(T) -> U) -> Spanned<U> {
        Spanned {
            node: f(self.node),
            span: self.span,
        }
    }

    /// Get a reference to the inner value.
    pub fn as_ref(&self) -> Spanned<&T> {
        Spanned {
            node: &self.node,
            span: self.span,
        }
    }
}

impl<T> std::ops::Deref for Spanned<T> {
    type Target = T;

    fn deref(&self) -> &Self::Target {
        &self.node
    }
}

impl<T> std::ops::DerefMut for Spanned<T> {
    fn deref_mut(&mut self) -> &mut Self::Target {
        &mut self.node
    }
}
