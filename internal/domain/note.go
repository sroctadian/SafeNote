package domain

import "time"

// Note is the persisted representation of a note. Title is stored as
// plaintext to support fast substring search (per spec section 3); all
// other sensitive content is stored only in encrypted form. Content is
// never populated on this struct except transiently in memory right
// after decryption — it must never be marshalled back to disk.
type Note struct {
	ID               string    `json:"id"`
	Title            string    `json:"title"`
	EncryptedContent []byte    `json:"-"`
	Salt             []byte    `json:"-"`
	Nonce            []byte    `json:"-"`
	Tags             []string  `json:"tags"`
	Favorite         bool      `json:"favorite"`
	Pinned           bool      `json:"pinned"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// NoteCard is the safe, non-sensitive projection of a Note used for
// list/grid views (spec section 2: only Title, CreatedAt, UpdatedAt).
type NoteCard struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Tags      []string  `json:"tags"`
	Favorite  bool      `json:"favorite"`
	Pinned    bool      `json:"pinned"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// DecryptedNote is the transient, in-memory-only representation shown
// after a successful PIN + decrypt flow. It must never be persisted.
type DecryptedNote struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
}

// SortMode enumerates supported home-page sort orders.
type SortMode string

const (
	SortNewest    SortMode = "newest"
	SortOldest    SortMode = "oldest"
	SortAlphabet  SortMode = "alphabet"
)

// ListQuery captures Home Page search/sort/pagination parameters.
type ListQuery struct {
	Search   string
	Sort     SortMode
	Page     int
	PageSize int
	OnlyFav  bool
}
