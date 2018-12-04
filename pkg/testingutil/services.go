package testingutil

import (
	"encoding/json"
	"github.com/pkg/errors"
)

// M is an alias for map[string]interface{}
type M map[string]interface{}

// FakePlanExternalMetaDataRaw creates fake plan metadata for testing purposes
func FakePlanExternalMetaDataRaw() ([][]byte, error) {
	planExternalMetaData1 := make(map[string]string)
	planExternalMetaData1["displayName"] = "plan-name-1"

	planExternalMetaData2 := make(map[string]string)
	planExternalMetaData2["displayName"] = "plan-name-2"

	planExternalMetaDataRaw1, err := json.Marshal(planExternalMetaData1)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	planExternalMetaDataRaw2, err := json.Marshal(planExternalMetaData2)
	if err != nil {
		return nil, errors.Wrap(err, "")
	}

	var data [][]byte
	data = append(data, planExternalMetaDataRaw1)
	data = append(data, planExternalMetaDataRaw2)

	return data, nil
}

// FakePlanServiceInstanceCreateParameterSchemasRaw creates 2 create parameter schemas for testing purposes
func FakePlanServiceInstanceCreateParameterSchemasRaw() ([][]byte, error) {
	planServiceInstanceCreateParameterSchema1 := make(M)
	planServiceInstanceCreateParameterSchema1["required"] = []string{"PLAN_DATABASE_URI", "PLAN_DATABASE_USERNAME", "PLAN_DATABASE_PASSWORD"}
	planServiceInstanceCreateParameterSchema1["properties"] = map[string]M{
		"PLAN_DATABASE_URI": {
			"default": "someuri",
			"type":    "string",
		},
		"PLAN_DATABASE_USERNAME": {
			"default": "name",
			"type":    "string",
		},
		"PLAN_DATABASE_PASSWORD": {
			"type": "string",
		},
		"SOME_OTHER": {
			"default": "other",
			"type":    "string",
		},
	}

	planServiceInstanceCreateParameterSchema2 := make(M)
	planServiceInstanceCreateParameterSchema2["required"] = []string{"PLAN_DATABASE_USERNAME_2", "PLAN_DATABASE_PASSWORD"}
	planServiceInstanceCreateParameterSchema2["properties"] = map[string]M{
		"PLAN_DATABASE_USERNAME_2": {
			"default": "user2",
			"type":    "string",
		},
		"PLAN_DATABASE_PASSWORD": {
			"type": "string",
		},
	}

	planServiceInstanceCreateParameterSchemaRaw1, err := json.Marshal(planServiceInstanceCreateParameterSchema1)
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	planServiceInstanceCreateParameterSchemaRaw2, err := json.Marshal(planServiceInstanceCreateParameterSchema2)
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "")
		}
	}

	var data [][]byte
	data = append(data, planServiceInstanceCreateParameterSchemaRaw1)
	data = append(data, planServiceInstanceCreateParameterSchemaRaw2)

	return data, nil
}
