# go-locale

[![Build Status](https://travis-ci.com/Xuanwo/go-locale.svg?branch=master)](https://travis-ci.com/Xuanwo/go-locale)
[![GoDoc](https://godoc.org/github.com/Xuanwo/go-locale?status.svg)](https://godoc.org/github.com/Xuanwo/go-locale)
[![Go Report Card](https://goreportcard.com/badge/github.com/Xuanwo/go-locale)](https://goreportcard.com/report/github.com/Xuanwo/go-locale)
[![codecov](https://codecov.io/gh/Xuanwo/go-locale/branch/master/graph/badge.svg)](https://codecov.io/gh/Xuanwo/go-locale)
[![License](https://img.shields.io/badge/license-apache%20v2-blue.svg)](https://github.com/Xuanwo/go-locale/blob/master/LICENSE)

`go-locale` is a Golang lib for cross platform locale detection.

## OS Support

Support all OS that Golang supported, except `android`:

- [aix: IBM AIX operating system](https://www.ibm.com/it-infrastructure/power/os/aix)
- android (*have on idea on this*)
- [darwin: Drawin, Mac OS X](https://opensource.apple.com/)
- [dragonfly: DragonFly BSD](https://www.dragonflybsd.org/)
- [freebsd: FreeBSD](https://www.freebsd.org/)
- [hurd: GNU Hurd](https://en.wikipedia.org/wiki/GNU_Hurd)
- [illumos](https://illumos.org/)
- [js: JavaScript runtime, WebAssembly](https://webassembly.org/)
- Linux: Ubuntu, CentOS, RHEL, Archlinux...
- [nacl: Native Client](https://developer.chrome.com/native-client)
- [netbsd: NetBSD](https://www.netbsd.org/)
- [openbsd: OpenBSD](https://www.openbsd.org/)
- [plan9: Plan 9 from Bell Labs](https://9p.io/plan9/)
- [solaris: Solaris](https://www.oracle.com/solaris)
- [windows: Windows](https://www.microsoft.com/en-us/windows/)
- [zos: z/OS](https://www.ibm.com/it-infrastructure/z/zos)

### POSIX Compatible Systems

- Lookup env `LANGUAGE`
- Lookup env `LC_ALL`
- Lookup env `LC_MESSAGES`
- Lookup env `LANG`
- Read file `$XDG_CONFIG_HOME/locale.conf`
- Read file `$HOME/.config/locale.conf`
- Read file `/etc/locale.conf`

### Js

- Lookup env `LANGUAGE`
- Lookup env `LC_ALL`

### Windows

- Lookup env `LANGUAGE`
- Lookup env `LC_ALL`
- Lookup env `LC_MESSAGES`
- Lookup env `LANG`
- [Windows Registry](https://renenyffenegger.ch/notes/Windows/registry/tree/HKEY_CURRENT_USER/Control-Panel/International/index)


### macOS X (darwin)

- Lookup env `LANGUAGE`
- Lookup env `LC_ALL`
- Lookup env `LC_MESSAGES`
- Lookup env `LANG`
- macOS X [User Defaults System](https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/UserDefaults/AboutPreferenceDomains/AboutPreferenceDomains.html)

## Usage

```go
import (
    "github.com/Xuanwo/go-locale"
)

func main() {
    tag, err := locale.Detect()
    if err != nil {
        log.Fatal(err)
    }
    // Have fun with language.Tag!

    tags, err := locale.DetectAll()
    if err != nil {
        log.Fatal(err)
    }
    // Get all available tags
}
```

## Acknowledgments

Inspired by [jibber_jabber](https://github.com/cloudfoundry-attic/jibber_jabber)
