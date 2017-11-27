package controller

import (
	"testing"

	"github.com/vikramsk/deepcloud/pkg/test"
)

func TestProviderInitializationFailure(t *testing.T) {
	cp, err := InitControllerServiceProvider("")
	test.Assert(t, err != nil, "expected error to not be nil")
	test.Assert(t, cp == nil, "expected provider to be nil")

}

func TestProviderInitializationSuccess(t *testing.T) {
	cp, err := InitControllerServiceProvider("$HOME/.kube/config")
	test.Assert(t, err != nil, "expected error to not be nil")
	test.Assert(t, cp == nil, "expected provider to be nil")

}
