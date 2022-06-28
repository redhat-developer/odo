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
	detectViaDefaultsSystem,
}

// detectViaUserDefaultsSystem will detect language via Apple User Defaults System
//
// We will read AppleLocale and AppleLanguages in this order:
//   - user AppleLocale
//   - user AppleLanguages
//   - global AppleLocale
//   - global AppleLanguages
//
// ref:
//  - Apple Developer Guide: https://developer.apple.com/library/archive/documentation/Cocoa/Conceptual/UserDefaults/AboutPreferenceDomains/AboutPreferenceDomains.html
//  - Homebrew: https://github.com/Homebrew/brew/pull/7940
func detectViaDefaultsSystem() ([]string, error) {
	// Read user's apple locale setting.
	m, err := parseDefaultsSystemAppleLocale("-g")
	if err == nil {
		return m, nil
	}
	// Read user's apple languages setting.
	m, err = parseDefaultsSystemAppleLanguages("-g")
	if err == nil {
		return m, nil
	}
	// Read global locale preferences.
	m, err = parseDefaultsSystemAppleLocale("/Library/Preferences/.GlobalPreferences")
	if err == nil {
		return m, nil
	}
	// Read global language preferences.
	m, err = parseDefaultsSystemAppleLanguages("/Library/Preferences/.GlobalPreferences")
	if err == nil {
		return m, nil
	}

	return nil, &Error{"detect via defaults system", ErrNotDetected}
}

// parseDefaultsSystemAppleLocale will parse the AppleLocale output.
func parseDefaultsSystemAppleLocale(domain string) ([]string, error) {
	cmd := exec.Command("defaults", "read", domain, "AppleLocale")

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return nil, &Error{"detect via user defaults system", err}
	}

	content := strings.TrimSpace(out.String())
	if len(content) == 0 {
		return nil, &Error{"detect via defaults system", ErrNotDetected}
	}
	return []string{content}, nil
}

// parseDefaultsSystemAppleLanguages will parse the AppleLanguages output.
//
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
func parseDefaultsSystemAppleLanguages(domain string) ([]string, error) {
	cmd := exec.Command("defaults", "read", domain, "AppleLanguages")

	var out bytes.Buffer
	cmd.Stdout = &out

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
		// Doing canonicalize
		if value, ok := oldAppleLocaleToCanonical[text]; ok {
			text = value
		}
		m = append(m, text)
	}

	if len(m) == 0 {
		return nil, &Error{"detect via user defaults system", ErrNotDetected}
	}
	return m, nil
}

// oldAppleLocaleToCanonical is borrowed from swift-corelibs-foundation's CFLocaleIdentifier.c
//
// Old Apple devices could return "English" instead of "en-US", this map will make them canonical
//
// refs:
//   - CFLocaleIdentifier.c: https://github.com/apple/swift-corelibs-foundation/blob/main/CoreFoundation/Locale.subproj/CFLocaleIdentifier.c
var oldAppleLocaleToCanonical = map[string]string{
	"Afrikaans":             "af",         //                      # __CFBundleLanguageNamesArray
	"Albanian":              "sq",         //                      # __CFBundleLanguageNamesArray
	"Amharic":               "am",         //                      # __CFBundleLanguageNamesArray
	"Arabic":                "ar",         //                      # __CFBundleLanguageNamesArray
	"Armenian":              "hy",         //                      # __CFBundleLanguageNamesArray
	"Assamese":              "as",         //                      # __CFBundleLanguageNamesArray
	"Aymara":                "ay",         //                      # __CFBundleLanguageNamesArray
	"Azerbaijani":           "az",         // -Arab,-Cyrl,-Latn?   # __CFBundleLanguageNamesArray (had 3 entries "Azerbaijani" for "az-Arab", "az-Cyrl", "az-Latn")
	"Basque":                "eu",         //                      # __CFBundleLanguageNamesArray
	"Belarusian":            "be",         //                      # handle other names
	"Belorussian":           "be",         //                      # handle other names
	"Bengali":               "bn",         //                      # __CFBundleLanguageNamesArray
	"Brazilian Portugese":   "pt-BR",      //                      # from Installer.app Info.plist IFLanguages key, misspelled
	"Brazilian Portuguese":  "pt-BR",      //                      # correct spelling for above
	"Breton":                "br",         //                      # __CFBundleLanguageNamesArray
	"Bulgarian":             "bg",         //                      # __CFBundleLanguageNamesArray
	"Burmese":               "my",         //                      # __CFBundleLanguageNamesArray
	"Byelorussian":          "be",         //                      # __CFBundleLanguageNamesArray
	"Catalan":               "ca",         //                      # __CFBundleLanguageNamesArray
	"Chewa":                 "ny",         //                      # handle other names
	"Chichewa":              "ny",         //                      # handle other names
	"Chinese":               "zh",         // -Hans,-Hant?         # __CFBundleLanguageNamesArray (had 2 entries "Chinese" for "zh-Hant", "zh-Hans")
	"Chinese, Simplified":   "zh-Hans",    //                      # from Installer.app Info.plist IFLanguages key
	"Chinese, Traditional":  "zh-Hant",    //                      # correct spelling for below
	"Chinese, Tradtional":   "zh-Hant",    //                      # from Installer.app Info.plist IFLanguages key, misspelled
	"Croatian":              "hr",         //                      # __CFBundleLanguageNamesArray
	"Czech":                 "cs",         //                      # __CFBundleLanguageNamesArray
	"Danish":                "da",         //                      # __CFBundleLanguageNamesArray
	"Dutch":                 "nl",         //                      # __CFBundleLanguageNamesArray (had 2 entries "Dutch" for "nl", "nl-BE")
	"Dzongkha":              "dz",         //                      # __CFBundleLanguageNamesArray
	"English":               "en",         //                      # __CFBundleLanguageNamesArray
	"Esperanto":             "eo",         //                      # __CFBundleLanguageNamesArray
	"Estonian":              "et",         //                      # __CFBundleLanguageNamesArray
	"Faroese":               "fo",         //                      # __CFBundleLanguageNamesArray
	"Farsi":                 "fa",         //                      # __CFBundleLanguageNamesArray
	"Finnish":               "fi",         //                      # __CFBundleLanguageNamesArray
	"Flemish":               "nl-BE",      //                      # handle other names
	"French":                "fr",         //                      # __CFBundleLanguageNamesArray
	"Galician":              "gl",         //                      # __CFBundleLanguageNamesArray
	"Gallegan":              "gl",         //                      # handle other names
	"Georgian":              "ka",         //                      # __CFBundleLanguageNamesArray
	"German":                "de",         //                      # __CFBundleLanguageNamesArray
	"Greek":                 "el",         //                      # __CFBundleLanguageNamesArray (had 2 entries "Greek" for "el", "grc")
	"Greenlandic":           "kl",         //                      # __CFBundleLanguageNamesArray
	"Guarani":               "gn",         //                      # __CFBundleLanguageNamesArray
	"Gujarati":              "gu",         //                      # __CFBundleLanguageNamesArray
	"Hawaiian":              "haw",        //                      # handle new languages
	"Hebrew":                "he",         //                      # __CFBundleLanguageNamesArray
	"Hindi":                 "hi",         //                      # __CFBundleLanguageNamesArray
	"Hungarian":             "hu",         //                      # __CFBundleLanguageNamesArray
	"Icelandic":             "is",         //                      # __CFBundleLanguageNamesArray
	"Indonesian":            "id",         //                      # __CFBundleLanguageNamesArray
	"Inuktitut":             "iu",         //                      # __CFBundleLanguageNamesArray
	"Irish":                 "ga",         //                      # __CFBundleLanguageNamesArray (had 2 entries "Irish" for "ga", "ga-dots")
	"Italian":               "it",         //                      # __CFBundleLanguageNamesArray
	"Japanese":              "ja",         //                      # __CFBundleLanguageNamesArray
	"Javanese":              "jv",         //                      # __CFBundleLanguageNamesArray
	"Kalaallisut":           "kl",         //                      # handle other names
	"Kannada":               "kn",         //                      # __CFBundleLanguageNamesArray
	"Kashmiri":              "ks",         //                      # __CFBundleLanguageNamesArray
	"Kazakh":                "kk",         //                      # __CFBundleLanguageNamesArray
	"Khmer":                 "km",         //                      # __CFBundleLanguageNamesArray
	"Kinyarwanda":           "rw",         //                      # __CFBundleLanguageNamesArray
	"Kirghiz":               "ky",         //                      # __CFBundleLanguageNamesArray
	"Korean":                "ko",         //                      # __CFBundleLanguageNamesArray
	"Kurdish":               "ku",         //                      # __CFBundleLanguageNamesArray
	"Lao":                   "lo",         //                      # __CFBundleLanguageNamesArray
	"Latin":                 "la",         //                      # __CFBundleLanguageNamesArray
	"Latvian":               "lv",         //                      # __CFBundleLanguageNamesArray
	"Lithuanian":            "lt",         //                      # __CFBundleLanguageNamesArray
	"Macedonian":            "mk",         //                      # __CFBundleLanguageNamesArray
	"Malagasy":              "mg",         //                      # __CFBundleLanguageNamesArray
	"Malay":                 "ms",         // -Latn,-Arab?         # __CFBundleLanguageNamesArray (had 2 entries "Malay" for "ms-Latn", "ms-Arab")
	"Malayalam":             "ml",         //                      # __CFBundleLanguageNamesArray
	"Maltese":               "mt",         //                      # __CFBundleLanguageNamesArray
	"Manx":                  "gv",         //                      # __CFBundleLanguageNamesArray
	"Marathi":               "mr",         //                      # __CFBundleLanguageNamesArray
	"Moldavian":             "mo",         //                      # __CFBundleLanguageNamesArray
	"Mongolian":             "mn",         // -Mong,-Cyrl?         # __CFBundleLanguageNamesArray (had 2 entries "Mongolian" for "mn-Mong", "mn-Cyrl")
	"Nepali":                "ne",         //                      # __CFBundleLanguageNamesArray
	"Norwegian":             "nb",         //                      # __CFBundleLanguageNamesArray (had "Norwegian" mapping to "no")
	"Nyanja":                "ny",         //                      # __CFBundleLanguageNamesArray
	"Nynorsk":               "nn",         //                      # handle other names (no entry in __CFBundleLanguageNamesArray)
	"Oriya":                 "or",         //                      # __CFBundleLanguageNamesArray
	"Oromo":                 "om",         //                      # __CFBundleLanguageNamesArray
	"Panjabi":               "pa",         //                      # handle other names
	"Pashto":                "ps",         //                      # __CFBundleLanguageNamesArray
	"Persian":               "fa",         //                      # handle other names
	"Polish":                "pl",         //                      # __CFBundleLanguageNamesArray
	"Portuguese":            "pt",         //                      # __CFBundleLanguageNamesArray
	"Portuguese, Brazilian": "pt-BR",      //                      # handle other names
	"Punjabi":               "pa",         //                      # __CFBundleLanguageNamesArray
	"Pushto":                "ps",         //                      # handle other names
	"Quechua":               "qu",         //                      # __CFBundleLanguageNamesArray
	"Romanian":              "ro",         //                      # __CFBundleLanguageNamesArray
	"Ruanda":                "rw",         //                      # handle other names
	"Rundi":                 "rn",         //                      # __CFBundleLanguageNamesArray
	"Russian":               "ru",         //                      # __CFBundleLanguageNamesArray
	"Sami":                  "se",         //                      # __CFBundleLanguageNamesArray
	"Sanskrit":              "sa",         //                      # __CFBundleLanguageNamesArray
	"Scottish":              "gd",         //                      # __CFBundleLanguageNamesArray
	"Serbian":               "sr",         //                      # __CFBundleLanguageNamesArray
	"Simplified Chinese":    "zh-Hans",    //                      # handle other names
	"Sindhi":                "sd",         //                      # __CFBundleLanguageNamesArray
	"Sinhalese":             "si",         //                      # __CFBundleLanguageNamesArray
	"Slovak":                "sk",         //                      # __CFBundleLanguageNamesArray
	"Slovenian":             "sl",         //                      # __CFBundleLanguageNamesArray
	"Somali":                "so",         //                      # __CFBundleLanguageNamesArray
	"Spanish":               "es",         //                      # __CFBundleLanguageNamesArray
	"Sundanese":             "su",         //                      # __CFBundleLanguageNamesArray
	"Swahili":               "sw",         //                      # __CFBundleLanguageNamesArray
	"Swedish":               "sv",         //                      # __CFBundleLanguageNamesArray
	"Tagalog":               "fil",        //                      # __CFBundleLanguageNamesArray
	"Tajik":                 "tg",         //                      # handle other names
	"Tajiki":                "tg",         //                      # __CFBundleLanguageNamesArray
	"Tamil":                 "ta",         //                      # __CFBundleLanguageNamesArray
	"Tatar":                 "tt",         //                      # __CFBundleLanguageNamesArray
	"Telugu":                "te",         //                      # __CFBundleLanguageNamesArray
	"Thai":                  "th",         //                      # __CFBundleLanguageNamesArray
	"Tibetan":               "bo",         //                      # __CFBundleLanguageNamesArray
	"Tigrinya":              "ti",         //                      # __CFBundleLanguageNamesArray
	"Tongan":                "to",         //                      # __CFBundleLanguageNamesArray
	"Traditional Chinese":   "zh-Hant",    //                      # handle other names
	"Turkish":               "tr",         //                      # __CFBundleLanguageNamesArray
	"Turkmen":               "tk",         //                      # __CFBundleLanguageNamesArray
	"Uighur":                "ug",         //                      # __CFBundleLanguageNamesArray
	"Ukrainian":             "uk",         //                      # __CFBundleLanguageNamesArray
	"Urdu":                  "ur",         //                      # __CFBundleLanguageNamesArray
	"Uzbek":                 "uz",         //                      # __CFBundleLanguageNamesArray
	"Vietnamese":            "vi",         //                      # __CFBundleLanguageNamesArray
	"Welsh":                 "cy",         //                      # __CFBundleLanguageNamesArray
	"Yiddish":               "yi",         //                      # __CFBundleLanguageNamesArray
	"ar_??":                 "ar",         //                      # from old MapScriptInfoAndISOCodes
	"az.Ar":                 "az-Arab",    //                      # from old LocaleRefGetPartString
	"az.Cy":                 "az-Cyrl",    //                      # from old LocaleRefGetPartString
	"az.La":                 "az",         //                      # from old LocaleRefGetPartString
	"be_??":                 "be_BY",      //                      # from old MapScriptInfoAndISOCodes
	"bn_??":                 "bn",         //                      # from old LocaleRefGetPartString
	"bo_??":                 "bo",         //                      # from old MapScriptInfoAndISOCodes
	"br_??":                 "br",         //                      # from old MapScriptInfoAndISOCodes
	"cy_??":                 "cy",         //                      # from old MapScriptInfoAndISOCodes
	"de-96":                 "de-1996",    //                      # from old MapScriptInfoAndISOCodes                     // <1.9>
	"de_96":                 "de-1996",    //                      # from old MapScriptInfoAndISOCodes                     // <1.9>
	"de_??":                 "de-1996",    //                      # from old MapScriptInfoAndISOCodes
	"el.El-P":               "grc",        //                      # from old LocaleRefGetPartString
	"en-ascii":              "en_001",     //                      # from earlier version of tables in this file!
	"en_??":                 "en_001",     //                      # from old MapScriptInfoAndISOCodes
	"eo_??":                 "eo",         //                      # from old MapScriptInfoAndISOCodes
	"es_??":                 "es_419",     //                      # from old MapScriptInfoAndISOCodes
	"es_XL":                 "es_419",     //                      # from earlier version of tables in this file!
	"fr_??":                 "fr_001",     //                      # from old MapScriptInfoAndISOCodes
	"ga-dots":               "ga-Latg",    //                      # from earlier version of tables in this file!          // <1.8>
	"ga-dots_IE":            "ga-Latg_IE", //                      # from earlier version of tables in this file!          // <1.8>
	"ga.Lg":                 "ga-Latg",    //                      # from old LocaleRefGetPartString                       // <1.8>
	"ga.Lg_IE":              "ga-Latg_IE", //                      # from old LocaleRefGetPartString                       // <1.8>
	"gd_??":                 "gd",         //                      # from old MapScriptInfoAndISOCodes
	"gv_??":                 "gv",         //                      # from old MapScriptInfoAndISOCodes
	"jv.La":                 "jv",         //                      # logical extension                                     // <1.9>
	"jw.La":                 "jv",         //                      # from old LocaleRefGetPartString
	"kk.Cy":                 "kk",         //                      # from old LocaleRefGetPartString
	"kl.La":                 "kl",         //                      # from old LocaleRefGetPartString
	"kl.La_GL":              "kl_GL",      //                      # from old LocaleRefGetPartString                       // <1.9>
	"lp_??":                 "se",         //                      # from old MapScriptInfoAndISOCodes
	"mk_??":                 "mk_MK",      //                      # from old MapScriptInfoAndISOCodes
	"mn.Cy":                 "mn",         //                      # from old LocaleRefGetPartString
	"mn.Mn":                 "mn-Mong",    //                      # from old LocaleRefGetPartString
	"ms.Ar":                 "ms-Arab",    //                      # from old LocaleRefGetPartString
	"ms.La":                 "ms",         //                      # from old LocaleRefGetPartString
	"nl-be":                 "nl-BE",      //                      # from old LocaleRefGetPartString
	"nl-be_BE":              "nl_BE",      //                      # from old LocaleRefGetPartString
	"no-NO":                 "nb-NO",      //                      # not handled by localeStringPrefixToCanonical
	"no-NO_NO":              "nb-NO_NO",   //                      # not handled by localeStringPrefixToCanonical
	"pa_??":                 "pa",         //                      # from old LocaleRefGetPartString
	"sa.Dv":                 "sa",         //                      # from old LocaleRefGetPartString
	"sl_??":                 "sl_SI",      //                      # from old MapScriptInfoAndISOCodes
	"sr_??":                 "sr_RS",      //                      # from old MapScriptInfoAndISOCodes						// <1.18>
	"su.La":                 "su",         //                      # from old LocaleRefGetPartString
	"yi.He":                 "yi",         //                      # from old LocaleRefGetPartString
	"zh-simp":               "zh-Hans",    //                      # from earlier version of tables in this file!
	"zh-trad":               "zh-Hant",    //                      # from earlier version of tables in this file!
	"zh.Ha-S":               "zh-Hans",    //                      # from old LocaleRefGetPartString
	"zh.Ha-S_CN":            "zh_CN",      //                      # from old LocaleRefGetPartString
	"zh.Ha-T":               "zh-Hant",    //                      # from old LocaleRefGetPartString
	"zh.Ha-T_TW":            "zh_TW",      //                      # from old LocaleRefGetPartString
}
