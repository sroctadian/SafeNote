package crypto

// Argon2id tuning parameters. Chosen to keep derivation comfortably
// under the <300ms encryption budget on typical desktop hardware
// while remaining resistant to offline brute force.
const (
	ArgonTime    uint32 = 3
	ArgonMemory  uint32 = 64 * 1024 // 64 MB
	ArgonThreads uint8  = 4
	ArgonKeyLen  uint32 = 32 // 256-bit key

	SaltSize  = 16 // bytes
	NonceSize = 12 // bytes, standard for AES-GCM
	KeySize   = 32 // bytes, AES-256
)
