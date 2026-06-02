package trust

import "github.com/mar1mo-41414/ore2ca/internal/store"

type Result struct {
	OS      bool
	Firefox bool
	OSErr   error
	FFErr   error
}

type UninstallResult struct {
	OS      bool
	Firefox bool
	OSErr   error
	FFErr   error
}

func Install(s *store.Store) (*Result, error) {
	return installPlatform(s)
}

func Uninstall(s *store.Store) (*UninstallResult, error) {
	r := &UninstallResult{}
	if err := uninstallPlatform(s); err != nil {
		r.OSErr = err
	} else {
		r.OS = true
	}
	r.Firefox, r.FFErr = uninstallNSS()
	return r, nil
}
