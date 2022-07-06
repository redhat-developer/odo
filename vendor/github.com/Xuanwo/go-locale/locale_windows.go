package locale

import (
	"golang.org/x/sys/windows/registry"
)

var detectors = []detector{
	detectViaEnvLanguage,
	detectViaEnvLc,
	detectViaRegistry,
}

// detectViaRegistry will detect language via Windows Registry
//
// ref: https://renenyffenegger.ch/notes/Windows/registry/tree/HKEY_CURRENT_USER/Control-Panel/International/index
func detectViaRegistry() (langs []string, err error) {
	defer func() {
		if err != nil {
			err = &Error{"detect via registry", err}
		}
	}()

	key, err := registry.OpenKey(registry.CURRENT_USER, `Control Panel\International`, registry.QUERY_VALUE)
	if err != nil {
		return nil, err
	}
	defer key.Close()

	lang, _, err := key.GetStringValue("LocaleName")
	if err != nil {
		return nil, err
	}

	return []string{lang}, nil
}
