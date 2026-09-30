//go:build !windows

package keychain

// restrictToOwner is a no-op off Windows: the 0o700/0o600 modes already make the
// session file and its directory owner-only.
func restrictToOwner(string) error { return nil }
