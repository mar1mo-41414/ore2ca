package trust

import "github.com/ore2ca/ore2ca/internal/store"

type Result struct {
	OS      bool
	Firefox bool
	OSErr   error
	FFErr   error
}

func Install(s *store.Store) (*Result, error) {
	return installPlatform(s)
}

func Uninstall(s *store.Store) error {
	return uninstallPlatform(s)
}
