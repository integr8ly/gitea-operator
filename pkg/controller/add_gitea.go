package controller

import (
	"github.com/integr8ly/gitea-operator/pkg/controller/gitea"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, gitea.Add)
}
