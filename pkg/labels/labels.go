package labels

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	dfutil "github.com/devfile/library/pkg/util"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/klog"

	"github.com/redhat-developer/odo/pkg/version"
)

var _replacementMap = map[string]string{
	".": "dot",
	"#": "sharp",
}

var (
	_regexpInvalidCharacters        = regexp.MustCompile(`[^a-zA-Z0-9._-]`)
	_regexpStartingWithAlphanumeric = regexp.MustCompile(`^[a-z0-9A-Z]`)
	_regexpEndingWithAlphanumeric   = regexp.MustCompile(`[a-z0-9A-Z]$`)
)

// GetLabels return labels that should be applied to every object for given component in active application
// if you need labels to filter component then use GetSelector instead
// Note: isPartOfComponent denotes if the label is required for a core resource(deployment, svc, pvc, pv) of a given component deployed with `odo dev`;
// it is the only thing that sets it apart from the resources created via other ways (`odo deploy`, deploying resource with apply command during `odo dev`)
func GetLabels(componentName string, applicationName string, runtime string, mode string, isPartOfComponent bool) map[string]string {
	labels := getLabels(componentName, applicationName, mode, true, isPartOfComponent)
	if runtime != "" {
		// 'app.openshift.io/runtime' label added by OpenShift console is always lowercase
		labels[openshiftRunTimeLabel] = strings.ToLower(sanitizeLabelValue(openshiftRunTimeLabel, runtime))
	}
	return labels
}

// sanitizeLabelValue makes sure that the value specified is a valid value for a Kubernetes label, which means:
// i) must be 63 characters or fewer (can be empty), ii) unless empty, must begin and end with an alphanumeric character ([a-z0-9A-Z]),
// iii) could contain dashes (-), underscores (_), dots (.), and alphanumerics between.
//
// As such, sanitizeLabelValue might perform the following operations (taking care of repeating the process if the result is not a valid label value):
//
// - replace leading or trailing characters (. with "dot" or "DOT", and # with "sharp" or "SHARP", depending on the value case)
//
// - replace all characters that are not dashes, underscores, dots or alphanumerics between with a dash (-)
//
// - truncate the overall result so that it is less than 63 characters.
func sanitizeLabelValue(key, value string) string {
	errs := validation.IsValidLabelValue(value)
	if len(errs) == 0 {
		return value
	}

	klog.V(4).Infof("invalid value for label %q: %q => sanitizing it: %v", key, value, strings.Join(errs, "; "))

	// Return the corresponding value if is replaceable immediately
	if v, ok := _replacementMap[value]; ok {
		return v
	}

	// Replacements if it starts or ends with a non-alphanumeric character
	value = replaceAllLeadingOrTrailingInvalidValues(value)

	// Now replace any characters that are not dashes, dots, underscores or alphanumerics between
	value = _regexpInvalidCharacters.ReplaceAllString(value, "-")

	// Truncate if length > 63
	if len(value) > validation.LabelValueMaxLength {
		value = dfutil.TruncateString(value, validation.LabelValueMaxLength)
	}

	if errs = validation.IsValidLabelValue(value); len(errs) == 0 {
		return value
	}
	return sanitizeLabelValue(key, value)
}

func replaceAllLeadingOrTrailingInvalidValues(value string) string {
	if value == "" {
		return ""
	}

	isAllCaseMatchingPredicate := func(p func(rune) bool, s string) bool {
		for _, r := range s {
			if !p(r) && unicode.IsLetter(r) {
				return false
			}
		}
		return true
	}
	getLabelValueReplacement := func(v, replacement string) string {
		if isAllCaseMatchingPredicate(unicode.IsLower, v) {
			return strings.ToLower(replacement)
		}
		if isAllCaseMatchingPredicate(unicode.IsUpper, v) {
			return strings.ToUpper(replacement)
		}
		return replacement
	}

	if !_regexpStartingWithAlphanumeric.MatchString(value) {
		vAfterFirstChar := value[1:]
		var isPrefixReplaced bool
		for k, val := range _replacementMap {
			if strings.HasPrefix(value, k) {
				value = getLabelValueReplacement(vAfterFirstChar, val) + vAfterFirstChar
				isPrefixReplaced = true
				break
			}
		}
		if !isPrefixReplaced {
			value = vAfterFirstChar
		}
		if value == "" {
			return value
		}
	}
	if !_regexpEndingWithAlphanumeric.MatchString(value) {
		vBeforeLastChar := value[:len(value)-1]
		var isSuffixReplaced bool
		for k, val := range _replacementMap {
			if strings.HasSuffix(value, k) {
				value = vBeforeLastChar + getLabelValueReplacement(vBeforeLastChar, val)
				isSuffixReplaced = true
				break
			}
		}
		if !isSuffixReplaced {
			value = vBeforeLastChar
		}
	}
	return value
}

// AddStorageInfo adds labels for storage resources
func AddStorageInfo(labels map[string]string, storageName string, isSourceVolume bool) {
	labels[kubernetesStorageNameLabel] = storageName
	labels[componentLabel] = labels[kubernetesInstanceLabel]
	labels[devfileStorageLabel] = storageName
	if isSourceVolume {
		labels[sourcePVCLabel] = storageName
	}
}

func GetStorageName(labels map[string]string) string {
	return labels[kubernetesStorageNameLabel]
}

func GetDevfileStorageName(labels map[string]string) string {
	return labels[devfileStorageLabel]
}

func GetComponentName(labels map[string]string) string {
	return labels[kubernetesInstanceLabel]
}

func GetAppName(labels map[string]string) string {
	return labels[kubernetesPartOfLabel]
}

func GetManagedBy(labels map[string]string) string {
	return labels[kubernetesManagedByLabel]
}

func GetManagedByVersion(labels map[string]string) string {
	return labels[kubernetesManagedByVersionLabel]
}

func IsManagedByOdo(labels map[string]string) bool {
	return labels[kubernetesManagedByLabel] == odoManager
}

func GetMode(labels map[string]string) string {
	return labels[odoModeLabel]
}

func GetProjectType(labels map[string]string, annotations map[string]string) (string, error) {
	// For backwards compatibility with previously deployed components that could be non-odo, check the annotation first
	// then check to see if there is a label with the project type
	if typ, ok := annotations[odoProjectTypeAnnotation]; ok {
		return typ, nil
	}
	if typ, ok := labels[odoProjectTypeAnnotation]; ok {
		return typ, nil
	}
	return "", errors.New("component type not found in labels or annotations")
}

func SetProjectType(annotations map[string]string, value string) {
	annotations[odoProjectTypeAnnotation] = value
}

// GetSelector returns a selector string used for selection of resources which are part of the given component in given mode
// Note: isPartOfComponent denotes if the selector is required for a core resource(deployment, svc, pvc, pv) of a given component deployed with `odo dev`
// it is the only thing that sets it apart from the resources created via other ways (`odo deploy`, deploying resource with apply command during `odo dev`)
func GetSelector(componentName string, applicationName string, mode string, isPartOfComponent bool) string {
	labels := getLabels(componentName, applicationName, mode, false, isPartOfComponent)
	return labels.String()
}

// getLabels return labels that should be applied to every object for given component in active application
// additional labels are used only for creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
// isPartOfComponent denotes if the label is required for a core resource(deployment, svc, pvc, pv) of a given component deployed with `odo dev`
// it is the only thing that sets it apart from the resources created via other ways (`odo deploy`, deploying resource with apply command during `odo dev`)
func getLabels(componentName string, applicationName string, mode string, additional bool, isPartOfComponent bool) k8slabels.Set {
	labels := getApplicationLabels(applicationName, additional)
	labels[kubernetesInstanceLabel] = componentName
	if mode != ComponentAnyMode {
		labels[odoModeLabel] = mode
	}
	if isPartOfComponent {
		labels[componentLabel] = componentName
	}
	return labels
}

// getApplicationLabels return labels that identifies given application
// additional labels are used only when creating object
// if you are creating something use additional=true
// if you need labels to filter component then use additional=false
func getApplicationLabels(application string, additional bool) k8slabels.Set {
	labels := k8slabels.Set{
		kubernetesPartOfLabel:    application,
		kubernetesManagedByLabel: odoManager,
	}
	if additional {
		labels[appLabel] = application
		labels[kubernetesManagedByVersionLabel] = version.VERSION
	}
	return labels
}
