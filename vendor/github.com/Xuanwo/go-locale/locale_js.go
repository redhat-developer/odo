package locale

var detectors = []detector{
	detectViaEnvLanguage,
	detectViaEnvLc,
}
