package init

import (
	"fmt"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/redhat-developer/odo/pkg/catalog"
)

// InteractiveBuilder is a backend that will ask init parameters interactively
type InteractiveBuilder struct{}

func (o *InteractiveBuilder) IsAdequate(flags map[string]string) bool {
	return len(flags) == 0
}

func (o *InteractiveBuilder) ParamsBuild() (initParams, error) {
	devfileEntries, _ := catalog.ListDevfileComponents("")
	langs := devfileEntries.GetLanguages()
	lang, err := askLanguage(langs)
	if err != nil {
		return initParams{}, err
	}
	types := devfileEntries.GetProjectTypes(lang)
	typ, details, err := askType(types)
	if err != nil {
		return initParams{}, err
	}
	fmt.Printf("typ: %s, details: %+v\n", typ, details)
	return initParams{}, nil
}

func askLanguage(langs []string) (string, error) {
	sort.Strings(langs)
	question := &survey.Select{
		Message: "Select language:",
		Options: langs,
	}
	var answer string
	err := survey.AskOne(question, &answer)
	if err != nil {
		return "", err
	}
	return answer, nil
}

func askType(types catalog.TypesWithDetails) (string, catalog.TypeDetails, error) {
	stringTypes := types.GetOrderedLabels()
	question := &survey.Select{
		Message: "Select project type:",
		Options: stringTypes,
	}
	var answerPos int
	err := survey.AskOne(question, &answerPos)
	if err != nil {
		return "", catalog.TypeDetails{}, err
	}
	return types.GetAtOrderedPosition(answerPos)
}
