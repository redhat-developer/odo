package application

type Client interface {
	List() ([]string, error)
	Exists(app string) (bool, error)
	Delete(name string) error
	GetMachineReadableFormat(appName string, projectName string) App
	GetMachineReadableFormatForList(apps []App) AppList
}
