// In-memory-only cache of the most recently unlocked note's PIN, so
// moving from View -> Edit for the SAME note doesn't re-prompt for a
// PIN the user just entered.
//
// Security note: this is intentionally narrow and short-lived —
// - Only ever holds ONE note's PIN at a time (not a general vault).
// - Lives only as a JS variable in memory; never written to
//   localStorage/sessionStorage/disk, so it never survives a reload.
// - Cleared whenever the user returns to Home (see home.js), so a note
//   left unlocked in the background does not stay unlocked indefinitely
//   across unrelated navigation.

let unlockedId = null;
let unlockedPin = null;

export function rememberUnlock(id, pin) {
  unlockedId = id;
  unlockedPin = pin;
}

export function getRememberedPin(id) {
  return unlockedId === id ? unlockedPin : null;
}

export function forgetUnlock() {
  unlockedId = null;
  unlockedPin = null;
}
