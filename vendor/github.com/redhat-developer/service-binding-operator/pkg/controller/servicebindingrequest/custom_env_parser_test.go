package servicebindingrequest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
)

func TestCustomEnvParser(t *testing.T) {
	type wantedVar struct {
		Name       string
		Value      string
		ErrMessage string
	}

	type args struct {
		in           map[string]interface{}
		wanted       []wantedVar
		varTemplates []corev1.EnvVar
	}

	testCase := func(args args) func(t *testing.T) {
		return func(t *testing.T) {
			parser := NewCustomEnvParser(args.varTemplates, args.in)
			values, err := parser.Parse()
			require.NoError(t, err)

			for _, w := range args.wanted {
				require.Equal(t, values[w.Name], w.Value, w.ErrMessage)
			}
		}
	}

	t.Run("spec and status only", testCase(args{
		in: map[string]interface{}{
			"spec": map[string]interface{}{
				"dbName": "database-name",
			},
			"status": map[string]interface{}{
				"creds": map[string]interface{}{
					"user": "database-user",
					"pass": "database-pass",
				},
			},
		},
		wanted: []wantedVar{
			{Name: "JDBC_CONNECTION_STRING", Value: "database-name:database-user@database-pass"},
			{Name: "ANOTHER_STRING", Value: "database-name_database-user"},
		},
		varTemplates: []corev1.EnvVar{
			corev1.EnvVar{
				Name:  "JDBC_CONNECTION_STRING",
				Value: `{{ .spec.dbName }}:{{ .status.creds.user }}@{{ .status.creds.pass }}`,
			},
			corev1.EnvVar{
				Name:  "ANOTHER_STRING",
				Value: `{{ .spec.dbName }}_{{ .status.creds.user }}`,
			},
		},
	}))
}

func TestCustomEnvPath_Parse(t *testing.T) {
	type args struct {
		envVarCtx map[string]interface{}
		templates []corev1.EnvVar
		expected  map[string]interface{}
		wantErr   error
	}

	assertParse := func(args args) func(*testing.T) {
		return func(t *testing.T) {
			customEnvParser := NewCustomEnvParser(args.templates, args.envVarCtx)
			actual, err := customEnvParser.Parse()
			if args.wantErr != nil {
				require.Error(t, args.wantErr)
				require.Equal(t, args.wantErr, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, args.expected, actual)
			}
		}
	}

	envVarCtx := map[string]interface{}{
		"spec": map[string]interface{}{
			"dbName": "database-name",
		},
		"status": map[string]interface{}{
			"creds": map[string]interface{}{
				"user": "database-user",
				"pass": "database-pass",
			},
		},
	}

	t.Run("JDBC connection string template", assertParse(args{
		envVarCtx: envVarCtx,
		templates: []corev1.EnvVar{
			{
				Name:  "JDBC_CONNECTION_STRING",
				Value: `{{ .spec.dbName }}:{{ .status.creds.user }}@{{ .status.creds.pass }}`,
			},
		},
		expected: map[string]interface{}{
			"JDBC_CONNECTION_STRING": "database-name:database-user@database-pass",
		},
	}))

	t.Run("incomplete template", assertParse(args{
		envVarCtx: envVarCtx,
		templates: []corev1.EnvVar{
			{
				Name:  "INCOMPLETE_TEMPLATE",
				Value: `{{ .spec.dbName `,
			},
		},
		wantErr: errors.New("template: set:1: unclosed action"),
	}))
}

func TestCustomEnvPath_Parse_exampleCase(t *testing.T) {
	cache := map[string]interface{}{
		"status": map[string]interface{}{
			"dbConfigMap": map[string]interface{}{
				"db.user":     "database-user",
				"db.password": "database-pass",
			},
		},
	}

	envMap := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "JDBC_USERNAME",
			Value: `{{ index .status.dbConfigMap "db.user" }}`,
		},
		corev1.EnvVar{
			Name:  "JDBC_PASSWORD",
			Value: `{{ index .status.dbConfigMap "db.password" }}`,
		},
	}

	customEnvPath := NewCustomEnvParser(envMap, cache)
	values, err := customEnvPath.Parse()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	str := values["JDBC_USERNAME"]
	require.Equal(t, "database-user", str, "Connection string is not matching")
	str2 := values["JDBC_PASSWORD"]
	require.Equal(t, "database-pass", str2, "Connection string is not matching")
}

func TestCustomEnvPath_Parse_ToJson(t *testing.T) {
	cache := map[string]interface{}{
		"spec": map[string]interface{}{
			"dbName": "database-name",
		},
		"status": map[string]interface{}{
			"creds": map[string]interface{}{
				"user": "database-user",
				"pass": "database-pass",
			},
		},
	}

	envMap := []corev1.EnvVar{
		corev1.EnvVar{
			Name:  "root",
			Value: `{{ json . }}`,
		},
		corev1.EnvVar{
			Name:  "spec",
			Value: `{{ json .spec }}`,
		},
		corev1.EnvVar{
			Name:  "status",
			Value: `{{ json .status }}`,
		},
		corev1.EnvVar{
			Name:  "creds",
			Value: `{{ json .status.creds }}`,
		},
		corev1.EnvVar{
			Name:  "dbName",
			Value: `{{ json .spec.dbName }}`,
		},
		corev1.EnvVar{
			Name:  "notExist",
			Value: `{{ json .notExist }}`,
		},
	}
	customEnvPath := NewCustomEnvParser(envMap, cache)
	values, err := customEnvPath.Parse()
	if err != nil {
		t.Error(err)
		t.Fail()
	}
	str := values["root"]
	require.Equal(t, `{"spec":{"dbName":"database-name"},"status":{"creds":{"pass":"database-pass","user":"database-user"}}}`, str, "root path json string is not matching")
	str2 := values["spec"]
	require.Equal(t, `{"dbName":"database-name"}`, str2, "spec json string is not matching")
	str3 := values["status"]
	require.Equal(t, `{"creds":{"pass":"database-pass","user":"database-user"}}`, str3, "status json string is not matching")
	str4 := values["creds"]
	require.Equal(t, `{"pass":"database-pass","user":"database-user"}`, str4, "creds json string is not matching")
	str5 := values["dbName"]
	require.Equal(t, `"database-name"`, str5, "dbName json string is not matching")
	str6 := values["notExist"]
	require.Equal(t, "null", str6, "notExist json string is not matching")
}
