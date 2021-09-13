// +build !integration_test

package locale

import (
	"bufio"
	"bytes"
	"os/exec"
	"strings"
)

var detectors = []detector{
	detectViaEnvLanguage,
	detectViaEnvLc,
	detectViaUserDefaultsSystem,
}

// detectViaUserDefaultsSystem will detect language via Apple User Defaults System
//
// ref: https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/UserDefaults/AboutPreferenceDomains/AboutPreferenceDomains.html
func detectViaUserDefaultsSystem() ([]string, error) {
	cmd := exec.Command("defaults", "read", "NSGlobalDomain", "AppleLanguages")

	var out bytes.Buffer
	cmd.Stdout = &out

	// Output should be like:
	//
	// (
	//    en,
	//    ja,
	//    fr,
	//    de,
	//    es,
	//    it,
	//    pt,
	//    "pt-PT",
	//    nl,
	//    sv,
	//    nb,
	//    da,
	//    fi,
	//    ru,
	//    pl,
	//    "zh-Hans",
	//    "zh-Hant",
	//    ko,
	//    ar,
	//    cs,
	//    hu,
	//    tr
	// )
	err := cmd.Run()
	if err != nil {
		return nil, &Error{"detect via user defaults system", err}
	}

	m := make([]string, 0)
	s := bufio.NewScanner(&out)
	for s.Scan() {
		text := s.Text()
		// Ignore "(" and ")"
		if !strings.HasPrefix(text, " ") {
			continue
		}
		// Trim all space, " and ,
		text = strings.Trim(text, " \",")
		m = append(m, text)
	}

	if len(m) == 0 {
		return nil, &Error{"detect via user defaults system", ErrNotDetected}
	}
	return m, nil
}
