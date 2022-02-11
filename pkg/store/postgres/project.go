package postgres

import (
	"strings"
)

// Project define project where transaction occurred
// Warning! Don't forget update enum in database when you add/modify any of that
type Project string

const (
	Undefined Project = "undefined"
	Sport     Project = "sport"
	Casino    Project = "casino"
)

func NewProject(val string) Project {
	switch v := Project(strings.ToLower(val)); v {
	case Sport, Casino:
		return v
	default:
		return Undefined
	}
}
