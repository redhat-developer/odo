package generator

import (
	"github.com/devfile/api/v2/pkg/apis/workspaces/v1alpha2"
	"github.com/devfile/api/v2/pkg/attributes"
)

type CompositeCommandParams struct {
	Id         string
	Attributes *attributes.Attributes

	Commands []string
	Parallel *bool

	Label     *string
	Kind      v1alpha2.CommandGroupKind
	IsDefault *bool
}

func GetCompositeCommand(params CompositeCommandParams) v1alpha2.Command {
	cmd := v1alpha2.Command{
		Id: params.Id,
		CommandUnion: v1alpha2.CommandUnion{
			Composite: &v1alpha2.CompositeCommand{
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind:      params.Kind,
							IsDefault: params.IsDefault,
						},
					},
				},
				Commands: params.Commands,
				Parallel: params.Parallel,
			},
		},
	}
	if params.Attributes != nil {
		cmd.Attributes = *params.Attributes
	}
	if params.Label != nil {
		cmd.Composite.Label = *params.Label
	}
	return cmd
}

type ExecCommandParams struct {
	Id         string
	Attributes *attributes.Attributes

	CommandLine      string
	Component        string
	WorkingDir       string
	Env              []v1alpha2.EnvVar
	HotReloadCapable *bool

	Label     *string
	Kind      v1alpha2.CommandGroupKind
	IsDefault *bool
}

func GetExecCommand(params ExecCommandParams) v1alpha2.Command {
	cmd := v1alpha2.Command{
		Id: params.Id,
		CommandUnion: v1alpha2.CommandUnion{
			Exec: &v1alpha2.ExecCommand{
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind:      params.Kind,
							IsDefault: params.IsDefault,
						},
					},
				},
				CommandLine:      params.CommandLine,
				Component:        params.Component,
				WorkingDir:       params.WorkingDir,
				Env:              params.Env,
				HotReloadCapable: params.HotReloadCapable,
			},
		},
	}
	if params.Attributes != nil {
		cmd.Attributes = *params.Attributes
	}
	if params.Label != nil {
		cmd.Composite.Label = *params.Label
	}
	return cmd
}

type ApplyCommandParams struct {
	Id         string
	Attributes *attributes.Attributes

	Component string

	Label     *string
	Kind      v1alpha2.CommandGroupKind
	IsDefault *bool
}

func GetApplyCommand(params ApplyCommandParams) v1alpha2.Command {
	cmd := v1alpha2.Command{
		Id: params.Id,
		CommandUnion: v1alpha2.CommandUnion{
			Apply: &v1alpha2.ApplyCommand{
				LabeledCommand: v1alpha2.LabeledCommand{
					BaseCommand: v1alpha2.BaseCommand{
						Group: &v1alpha2.CommandGroup{
							Kind:      params.Kind,
							IsDefault: params.IsDefault,
						},
					},
				},
				Component: params.Component,
			},
		},
	}
	if params.Attributes != nil {
		cmd.Attributes = *params.Attributes
	}
	if params.Label != nil {
		cmd.Composite.Label = *params.Label
	}
	return cmd
}
