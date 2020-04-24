package config

import (
	"fmt"
	"path/filepath"

	"github.com/mkmik/multierror"
	"k8s.io/apimachinery/pkg/api/validation"
)

type validateVisitor struct {
	errs []error
}

func (m *Manifest) Validate() error {
	vv := &validateVisitor{errs: []error{}}
	m.Walk(vv)
	return multierror.Join(vv.errs)
}

func (vv *validateVisitor) Environment(env *Environment) error {
	if err := validateName(env.Name, PathForEnvironment(env)); err != nil {
		vv.errs = append(vv.errs, err)
	}
	if err := validatePipelines(env.Pipelines, PathForEnvironment(env)); err != nil {
		vv.errs = append(vv.errs, err...)
	}
	return nil
}

func (vv *validateVisitor) Application(env *Environment, app *Application) error {
	if err := validateName(app.Name, PathForApplication(env, app)); err != nil {
		vv.errs = append(vv.errs, err)
	}
	return nil
}

func (vv *validateVisitor) Service(env *Environment, app *Application, svc *Service) error {
	if err := validateName(svc.Name, PathForService(env, svc)); err != nil {
		vv.errs = append(vv.errs, err)
	}
	if err := validateWebhook(svc.Webhook, PathForService(env, svc)); err != nil {
		vv.errs = append(vv.errs, err...)
	}
	if err := validatePipelines(svc.Pipelines, PathForService(env, svc)); err != nil {
		vv.errs = append(vv.errs, err...)
	}

	return nil
}

func validateWebhook(hook *Webhook, path string) []error {
	errs := []error{}
	if hook == nil {
		return nil
	}
	if hook.Secret == nil {
		return list(notFoundError("secret", path))
	}
	if err := validateName(hook.Secret.Name, filepath.Join(path, "secret")); err != nil {
		errs = append(errs, err)
	}
	if err := validateName(hook.Secret.Namespace, filepath.Join(path, "secret")); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func validatePipelines(pipelines *Pipelines, path string) []error {
	errs := []error{}
	if pipelines == nil {
		return nil
	}
	if pipelines.Integration == nil {
		return list(notFoundError("pipelines", path))
	}
	if err := validateName(pipelines.Integration.Template, filepath.Join(path, "pipelines")); err != nil {
		errs = append(errs, err)
	}
	if err := validateName(pipelines.Integration.Binding, filepath.Join(path, "pipelines")); err != nil {
		errs = append(errs, err)
	}
	return errs
}

func validateName(name, path string) error {
	err := validation.NameIsDNS1035Label(name, true)
	if len(err) > 0 {
		return fmt.Errorf("%q is not a valid name at %v: \n%v", name, path, err)
	}
	return nil
}

func notFoundError(field string, at string) error {
	return fmt.Errorf("%v not found at %v", field, at)
}

func list(errs ...error) []error {
	return errs
}
