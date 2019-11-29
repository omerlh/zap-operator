package controller

import (
	"github.com/omerlh/zap-operator/pkg/controller/zaproxy"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, zaproxy.Add)
}
