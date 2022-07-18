package util

import (
	"fmt"
)

var InfoPure = Info{}

type Info struct {
	Resources []ResourceDef
}

func (i Info) String() string {
	re := ""
	for j, input := range i.Resources {
		re += input.String()
		if j != len(i.Resources)-1 {
			re += " "
		}
	}
	return re
}

func (i Info) StringWith(args []string) string {
	re := ""
	for j, resource := range i.Resources {
		re += resource.StringWith(args[j])
		if j != len(i.Resources)-1 {
			re += " "
		}
	}
	return re
}

type ResourceDef struct {
	Name string
	Arg  int
}

func (r ResourceDef) String() string {
	return fmt.Sprintf("%s(%d)", r.Name, r.Arg)
}

func (r ResourceDef) StringWith(arg string) string {
	return fmt.Sprintf("%s(%s)", r.Name, arg)
}

type Resource struct {
	Name string
	Arg  string
}

func (r Resource) String() string {
	return fmt.Sprintf("%s(%s)", r.Name, r.Arg)
}
