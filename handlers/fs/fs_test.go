package fs_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "fs")
}

var currPath, _ = filepath.Abs(".")
var fixturesPath = filepath.Join(currPath, "..", "..", "testhelper", "fixtures")
