package project

type Client interface {
	SetCurrent(projectName string) error
	Create(projectName string, wait bool) error
	Delete(projectName string, wait bool) error
	List() (ProjectList, error)
	Exists(projectName string) (bool, error)
}
