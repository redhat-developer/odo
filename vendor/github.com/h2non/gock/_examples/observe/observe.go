package main

import (
	"bytes"
	"gopkg.in/h2non/gock.v1"
	"net/http"
)

func main() {
	defer gock.Off()
	gock.Observe(gock.DumpRequest)

	gock.New("http://foo.com").
		Post("/bar").
		MatchType("json").
		JSON(map[string]string{"foo": "bar"}).
		Reply(200)

	body := bytes.NewBuffer([]byte(`{"foo":"bar"}`))
	http.Post("http://foo.com/bar", "application/json", body)
}
