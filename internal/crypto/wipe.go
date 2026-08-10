package crypto

// Wipe overwrites a byte slice with zeroes in place. Go's garbage
// collector and compiler optimizations mean this is best-effort, not a
// hard guarantee (the runtime may have copied the backing array), but it
// meaningfully shrinks the window a sensitive value like a derived key,
// PIN, or plaintext lives in memory.
func Wipe(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

// WipeString cannot zero a Go string in place (strings are immutable),
// so callers should avoid holding sensitive data in string form for any
// longer than necessary and prefer []byte where wiping matters. This
// helper exists to make that intent explicit at call sites and is a
// no-op by design.
func WipeString(_ *string) {
	// Intentionally not implemented: Go strings are immutable and
	// interned; a caller that needs guaranteed wiping must use []byte.
}
