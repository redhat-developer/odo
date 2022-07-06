//go:build aix || dragonfly || freebsd || hurd || illumos || linux || nacl || netbsd || openbsd || plan9 || solaris || zos
// +build aix dragonfly freebsd hurd illumos linux nacl netbsd openbsd plan9 solaris zos

package locale

import (
	"bufio"
	"os"
	"path"
	"strings"
)

var detectors = []detector{
	detectViaEnvLanguage,
	detectViaEnvLc,
	detectViaLocaleConf,
}

func detectViaLocaleConf() (_ []string, err error) {
	defer func() {
		if err != nil {
			err = &Error{"detect via locale conf", err}
		}
	}()

	fp := getLocaleConfPath()
	if fp == "" {
		return nil, ErrNotDetected
	}

	f, err := os.Open(fp)
	if err != nil {
		return nil, err
	}

	// Output should be like:
	//
	// LANG=en_US.UTF-8
	// LC_CTYPE="en_US.UTF-8"
	// LC_NUMERIC="en_US.UTF-8"
	// LC_TIME="en_US.UTF-8"
	// LC_COLLATE="en_US.UTF-8"
	// LC_MONETARY="en_US.UTF-8"
	// LC_MESSAGES=
	// LC_PAPER="en_US.UTF-8"
	// LC_NAME="en_US.UTF-8"
	// LC_ADDRESS="en_US.UTF-8"
	// LC_TELEPHONE="en_US.UTF-8"
	// LC_MEASUREMENT="en_US.UTF-8"
	// LC_IDENTIFICATION="en_US.UTF-8"
	// LC_ALL=
	m := make(map[string]string)
	s := bufio.NewScanner(f)
	for s.Scan() {
		value := strings.Split(s.Text(), "=")
		// Ignore not set locale value.
		if len(value) != 2 || value[1] == "" {
			continue
		}
		m[value[0]] = strings.Trim(value[1], "\"")
	}

	for _, v := range envs {
		x, ok := m[v]
		if ok {
			return []string{parseEnvLc(x)}, nil
		}
	}
	return nil, ErrNotDetected
}

// getLocaleConfPath will try to get correct locale conf path.
//
// Following path could be returned:
//   - "$XDG_CONFIG_HOME/locale.conf" (follow XDG Base Directory specification)
//   - "$HOME/.config/locale.conf" (user level locale config)
//   - "/etc/locale.conf" (system level locale config)
//   - "" (empty means no valid path found, caller need to handle this.)
//
// ref:
//   - POSIX Locale: https://pubs.opengroup.org/onlinepubs/9699919799/
//   - XDG Base Directory: https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
func getLocaleConfPath() string {
	// Try to loading from $XDG_CONFIG_HOME/locale.conf
	xdg, ok := os.LookupEnv("XDG_CONFIG_HOME")
	if ok {
		fp := path.Join(xdg, "locale.conf")
		_, err := os.Stat(fp)
		if err == nil {
			return fp
		}
	}

	// Try to loading from $HOME/.config/locale.conf
	home, ok := os.LookupEnv("HOME")
	if ok {
		fp := path.Join(home, ".config", "locale.conf")
		_, err := os.Stat(fp)
		if err == nil {
			return fp
		}
	}

	// Try to loading from /etc/locale.conf
	fp := "/etc/locale.conf"
	_, err := os.Stat(fp)
	if err == nil {
		return fp
	}

	return ""
}
