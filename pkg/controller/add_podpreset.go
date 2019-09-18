package controller

import (
	"github.com/redhat-cop/podpreset-webhook/pkg/controller/podpreset"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, podpreset.Add)
}
