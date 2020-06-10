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
dryrun(){
	if [ "{{ .Client }}" != "" ];then {{ .Client }} apply --dry-run=$(inputs.params.DRYRUN) -k $1; fi
}
if [ -d ${argo_path} ]
then
	printf "Apply $(basename ${argo_path}) applications\n"
	dryrun "${argo_path}/config"
	is_argocd=true
fi
printf "Apply $(basename ${cicd_path}) environment\n"
dryrun "${cicd_path}/overlays"
for dir in $(ls -d environments/*/)
do
	if ! $is_argocd
	then
		printf "Apply $(basename ${dir}) environment\n"
		dryrun "${dir}env/overlays"
	else
		if [ -d "${dir}apps" ]
		then
			for app in $(ls -d ${dir}apps/*/)
			do
				printf "Apply $(basename ${app}) application\n"
				dryrun $app
			done
		else
			printf "Apply $(basename ${dir}) environment\n"
			dryrun "${dir}env/overlays"
		fi
	fi
done`

type templateParam struct {
	Client  string
	CICDEnv string
}

// MakeScript will create a script that can dry-run/apply
// across all environments/applications
func MakeScript(client, cicdEnv string) (string, error) {
	params := templateParam{CICDEnv: cicdEnv, Client: client}
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
