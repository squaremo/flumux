package main

import (
	"reflect"
	"testing"
)

var testCases = map[string][]string{
	"nginx":                                 []string{"index.docker.io", "library/nginx", ""},
	"nginx:v1.0":                            []string{"index.docker.io", "library/nginx", "v1.0"},
	"squaremo/flumux":                       []string{"index.docker.io", "squaremo/flumux", ""},
	"squaremo/flumux:master-abc123":         []string{"index.docker.io", "squaremo/flumux", "master-abc123"},
	"quay.io/squaremo/flumux":               []string{"quay.io", "squaremo/flumux", ""},
	"quay.io/squaremo/flumux:master-abc123": []string{"quay.io", "squaremo/flumux", "master-abc123"},
}

func TestImageParts(t *testing.T) {
	for repo, parts := range testCases {
		host, image, tag, err := imageParts(repo)
		if err != nil {
			t.Error(err)
			continue
		}
		if !reflect.DeepEqual(parts, []string{host, image, tag}) {
			t.Errorf("Expected %+v, got [%s, %s, %s]", parts, host, image, tag)
		}
	}
}
