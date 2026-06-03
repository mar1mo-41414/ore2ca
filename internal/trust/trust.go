package trust

import "github.com/mar1mo-41414/ore2ca/internal/store"

// Options controls which browsers to register.
// If both Firefox and Chrome are false, auto-detect all installed browsers.
type Options struct {
	Firefox bool
	Chrome  bool
}

// explicit returns true when at least one browser flag was explicitly set.
func (o Options) explicit() bool { return o.Firefox || o.Chrome }

// BrowserResult holds the outcome for a single browser.
// Skipped=true means the browser was not installed or not requested — not an error.
type BrowserResult struct {
	Registered bool
	Skipped    bool
	Err        error
}

type Result struct {
	OS      bool
	OSErr   error
	Firefox BrowserResult
	Chrome  BrowserResult
}

type UninstallResult struct {
	OS      bool
	OSErr   error
	Firefox BrowserResult
	Chrome  BrowserResult
}

func Install(s *store.Store, opts Options) (*Result, error) {
	return installPlatform(s, opts)
}

func Uninstall(s *store.Store, opts Options) (*UninstallResult, error) {
	return uninstallPlatform(s, opts)
}
