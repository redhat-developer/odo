// +build !integration_test

package locale

var detectors = []detector{
	detectViaEnvLanguage,
	detectViaEnvLc,
}
