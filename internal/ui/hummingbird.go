package ui

import "strings"

// HummingbirdKeys defines the 30-key Hummingbird layout organized in groups of 5:
// 1. Home row left hand: a, s, d, f, g
// 2. Home row right hand: h, j, k, l, ;
// 3. Bottom row left hand: z, x, c, v, b
// 4. Bottom row right hand: n, m, ,, ., /
// 5. Top row left hand: q, w, e, r, t
// 6. Top row right hand: y, u, i, o, p
var HummingbirdKeys = []string{
	// Group 1: Home row - Left hand
	"a", "s", "d", "f", "g",
	// Group 2: Home row - Right hand
	"h", "j", "k", "l", ";",
	// Group 3: Bottom row - Left hand
	"z", "x", "c", "v", "b",
	// Group 4: Bottom row - Right hand
	"n", "m", ",", ".", "/",
	// Group 5: Top row - Left hand
	"q", "w", "e", "r", "t",
	// Group 6: Top row - Right hand
	"y", "u", "i", "o", "p",
}

// KeyToIndex returns the 0-based item index corresponding to the given key,
// or -1 if the key is not in HummingbirdKeys.
func KeyToIndex(key string) int {
	normalized := strings.ToLower(key)
	for i, k := range HummingbirdKeys {
		if k == normalized {
			return i
		}
	}
	return -1
}

// IndexToKey returns the key corresponding to the given item index,
// or an empty string if the index is out of bounds.
func IndexToKey(index int) string {
	if index >= 0 && index < len(HummingbirdKeys) {
		return HummingbirdKeys[index]
	}
	return ""
}
