package parser

import (
	"testing"

	v100 "github.com/cli-playground/devfile-parser/pkg/devfile/parser/data/1.0.0"
)

const (
	validJson100 = `{"apiVersion":"1.0.0","metadata":{"name":"java-web-spring"},"projects":[{"name":"java-web-spring","source":{"type":"git","location":"https://github.com/spring-projects/spring-petclinic.git"}}],"components":[{"type":"chePlugin","id":"redhat/java/latest","memoryLimit":"1512Mi"},{"alias":"tools","type":"dockerimage","image":"quay.io/eclipse/che-java8-maven:nightly","memoryLimit":"768Mi"}],"commands":[{"actions":[{"command":"mvn clean install","component":"tools","type":"build","workdir":"${CHE_PROJECTS_ROOT}/java-web-spring"}],"name":"maven build"},{"actions":[{"command":"java -jar -Xdebug -Xrunjdwp:transport=dt_socket,server=y,suspend=n,address=5005 \\\ntarget/*.jar\n","component":"tools","type":"run","workdir":"${CHE_PROJECTS_ROOT}/java-web-spring"}],"name":"run webapp"}]}`
)

func TestValidateDevfileSchema(t *testing.T) {

	t.Run("valid 1.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v100.JsonSchema100,
				rawContent: validJsonRawContent100(),
			}
		)

		err := d.ValidateDevfileSchema()
		if err != nil {
			t.Errorf("unexpected error: '%v'", err)
		}
	})

	t.Run("invalid 1.0.0 json schema", func(t *testing.T) {

		var (
			d = DevfileCtx{
				jsonSchema: v100.JsonSchema100,
				rawContent: []byte("{}"),
			}
		)

		err := d.ValidateDevfileSchema()
		if err == nil {
			t.Errorf("expected error, didn't get one")
		}
	})
}

func validJsonRawContent100() []byte {
	return []byte(validJson100)
}
