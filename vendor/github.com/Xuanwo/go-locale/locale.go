package locale

import (
	"errors"

	"golang.org/x/text/language"
)

// Detect will detect current env's language.
func Detect() (tag language.Tag, err error) {
	tags, err := DetectAll()
	if err != nil {
		return language.Und, err
	}
	return tags[0], nil
}

// DetectAll will detect current env's all available language.
func DetectAll() (tags []language.Tag, err error) {
	lang, err := detect()
	if err != nil {
		return
	}

	tags = make([]language.Tag, 0, len(lang))
	for _, v := range lang {
		tags = append(tags, language.Make(v))
	}
	return
}

type detector func() ([]string, error)

func detect() (lang []string, err error) {
	for _, fn := range detectors {
		lang, err = fn()
		if err != nil && errors.Is(err, ErrNotDetected) {
			continue
		}
		if err != nil {
			return
		}
		return
	}
	return nil, &Error{"detect", ErrNotDetected}
}
