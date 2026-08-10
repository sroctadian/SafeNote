# Architecture Decision Record — SafeNote

## ADR-001: Per-note key derivation (Secret Key + PIN via Argon2id)

**Decision**: Each note's encryption key is derived independently via
`Argon2id(PIN + SecretKey, per-note random salt)`, rather than using one
global key for all notes.

**Rationale**: A compromised single note's key (e.g. a guessed weak PIN)
does not expose any other note, since every note has its own salt and
therefore its own derived key even under an identical PIN. This also lets
each note carry a different PIN.

**Consequence**: Opening N notes at once requires N separate Argon2id
derivations. At the tuned parameters (64MB memory, time=3, threads=4)
this stays within the <300ms budget on typical desktop hardware for a
single note; bulk operations (e.g. search) never derive keys since search
operates over the plaintext title only.

## ADR-002: Title stored as plaintext

**Decision**: Note titles are stored unencrypted to support instant
substring search (<100ms for 100k notes) without deriving a key per
keystroke.

**Rationale**: Full-content encryption with encrypted titles would
require either (a) deriving keys and decrypting every note on every
search keystroke, which cannot meet the latency budget, or (b) a
searchable-encryption scheme, which is significantly more complex and
still leaks access patterns. Storing only the title in plaintext is an
explicit, documented trade-off called out in the spec itself ("Title may
remain plaintext only if required for searching").

**Mitigation**: Users who need titles protected too can avoid putting
sensitive information in the title field and rely on note content
(which is always encrypted) for anything sensitive.

## ADR-003: Local vault key wraps the Secret Key at rest

**Decision**: The Secret Key is not stored in plaintext or hashed alone;
it is encrypted with a separate, randomly generated 256-bit "vault key"
stored in its own file (`vault.key`, 0600 permissions) outside the SQLite
database.

**Rationale**: This adds defense in depth — copying the `.db` file alone
does not expose the Secret Key. An attacker needs both files.

**Limitation**: This is not full OS-keychain integration (Windows
Credential Manager / macOS Keychain / Secret Service). A future revision
could swap `internal/crypto/keystore.go` for an OS-keychain-backed
implementation without changing any calling code, since `SettingsService`
only depends on receiving 32 bytes of key material.

## ADR-004: Changing the Secret Key does not retroactively re-encrypt notes

**Decision**: `ChangeSecretKey` re-wraps the Secret Key value itself but
does **not** walk every note and re-derive/re-encrypt its content.

**Rationale**: Existing notes were encrypted with a key derived from the
*old* Secret Key + their PIN. Re-encrypting all notes transactionally on
every Secret Key change is a larger, higher-risk operation (must decrypt
every note, which requires either caching all PINs — never done — or
prompting for every PIN individually).

**Consequence**: After changing the Secret Key, existing notes must still
be opened with their original PIN *and* the Secret Key that was active
when they were last saved. A future revision should implement an
explicit "Re-encrypt all notes" maintenance flow in Settings that opens
each note (prompting for its PIN) and re-saves it under the new Secret
Key, batching the work with progress feedback.

## ADR-005: Generic error on any unlock failure

**Decision**: `NoteService.Open` collapses every possible failure — wrong
PIN, wrong Secret Key, corrupted ciphertext, missing note — into a single
`ErrInvalidPin` returned to the frontend.

**Rationale**: Spec section 4 explicitly requires that a wrong PIN "never
reveal reason". Distinguishing "note not found" from "wrong PIN" would
let an attacker enumerate valid note IDs.

## ADR-006: In-memory-only lockout tracking

**Decision**: Failed-attempt counters and cooldown timers live in a
process-local `LockoutTracker`, not in the database.

**Rationale**: SafeNote is a single-user offline desktop app; the
cooldown only needs to survive within a running session, and keeping it
out of the database avoids adding another place brute-force attempt
metadata could leak. Restarting the app resets cooldowns — an accepted
trade-off given the also-required Argon2id cost per guess.
