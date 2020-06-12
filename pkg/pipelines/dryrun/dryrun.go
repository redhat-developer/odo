package dryrun

import (
	"bytes"
	"fmt"
	"text/template"
)

const scriptTemplate = `#!/bin/sh
is_argocd=false
argo_path="config/argocd"
cicd_path="config/{{ .CICDEnv }}"
cmd={{ .Cmd }}
overall_exit=0

execute() {
  if [[ ! -z "${cmd}" ]]; then $cmd apply --dry-run=$(inputs.params.DRYRUN) -k $1; fi
  e=$?
  if [ $e -gt $overall_exit ]; then
    overall_exit=$e
  fi
}

if [[ -d "${argo_path}" ]]; then
  printf "Apply $(basename ${argo_path}) applications\n"
  execute "${argo_path}/config"
  is_argocd=true
fi

printf "Apply $(basename ${cicd_path}) environment\n"
execute "${cicd_path}/overlays"

for dir in $(ls -d environments/*/); do
  if ! $is_argocd; then
    printf "Apply $(basename ${dir}) environment\n"
    execute "${dir}env/overlays"
  else
    if [[ -d "${dir}apps" ]]; then
      for app in $(ls -d ${dir}apps/*/); do
        printf "Apply $(basename ${app}) application\n"
        execute $app
      done
    else
      printf "Apply $(basename ${dir}) environment\n"
      execute "${dir}env/overlays"
    fi
  fi
done

exit $overall_exit
`

type templateParam struct {
	Cmd     string
	CICDEnv string
}

// MakeScript will create a script that can dry-run/apply
// across all environments/applications
func MakeScript(command, cicdEnv string) (string, error) {
	params := templateParam{CICDEnv: cicdEnv, Cmd: command}
	template, err := template.New("dryrun_script").Parse(scriptTemplate)
	if err != nil {
		return "", fmt.Errorf("unable to parse template: %v", err)
	}
	var buf bytes.Buffer
	err = template.Execute(&buf, params)
	if err != nil {
		return "", fmt.Errorf("unable to execute template: %v", err)
	}
	return buf.String(), nil
}
