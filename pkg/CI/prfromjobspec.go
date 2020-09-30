package CI

import (
	"encoding/json"
	"fmt"
)

type pull struct {
	Number int `json:"number,omitempty"`
}

type ref struct {
	Pulls []pull `json:"pulls,omitempty"`
}

type jobspec struct {
	Refs ref `json:"refs,omitempty"`
}

func PRFromJobSpec(job_spec string) (string, error) {
	var pr string
	js := &jobspec{}
	err := json.Unmarshal([]byte(job_spec), js)
	if err != nil {
		return pr, err
	}
	pr = fmt.Sprintf("%d", js.Refs.Pulls[0].Number)
	return pr, nil
}
