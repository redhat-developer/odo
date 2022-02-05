package backend

import (
	"path/filepath"
	"runtime"
	"strings"

	"github.com/redhat-developer/alizer/go/pkg/apis/language"
)

// Below functions are from:
// https://github.com/redhat-developer/alizer/blob/main/go/test/apis/language_recognizer_test.go
func GetTestProjectPath(folder string) string {
	_, b, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(b)
	return filepath.Join(basepath, "..", "..", "..", "tests/examples/source/", folder)
}

func hasWantedLanguage(languages []language.Language, wantedLang string, wantedTools []string, wantedFrameworks []string) bool {
	for _, lang := range languages {
		if strings.ToLower(lang.Name) == wantedLang {
			return hasWantedTools(lang, wantedTools) && hasWantedFrameworks(lang, wantedFrameworks)
		}
	}
	return false
}

func hasWantedFrameworks(language language.Language, wantedFrameworks []string) bool {
	for _, wantedFramework := range wantedFrameworks {
		if !hasWantedFramework(language, wantedFramework) {
			return false
		}
	}
	return true
}

func hasWantedFramework(language language.Language, wantedFramework string) bool {
	for _, framework := range language.Frameworks {
		if strings.ToLower(framework) == wantedFramework {
			return true
		}
	}
	return false
}

func hasWantedTools(language language.Language, wantedTools []string) bool {
	for _, wantedTool := range wantedTools {
		if !hasWantedTool(language, wantedTool) {
			return false
		}
	}
	return true
}

func hasWantedTool(language language.Language, wantedTool string) bool {
	for _, tool := range language.Tools {
		if strings.ToLower(tool) == wantedTool {
			return true
		}
	}
	return false
}
