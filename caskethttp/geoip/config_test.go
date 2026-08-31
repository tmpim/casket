package geoip

import (
	"reflect"
	"testing"

	"github.com/tmpim/casket"
)

func TestParseConfig(t *testing.T) {
	controller := casket.NewTestController("http", `
		localhost:8080
		geoip path/to/maxmind/db
	`)
	actual, err := parseConfig(controller)
	if err != nil {
		t.Errorf("parseConfig return err: %v", err)
	}
	expected := Config{
		DatabasePath: "path/to/maxmind/db",
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("Expected %v actual %v", expected, actual)
	}
}
