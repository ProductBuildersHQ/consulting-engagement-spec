// Package testfixture gives tests the path to the example agreement.
package testfixture

import (
	"path/filepath"
	"runtime"
)

// ExamplePath returns the absolute path of a file under examples/.
func ExamplePath(elem ...string) string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("testfixture: no caller information")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "examples")
	return filepath.Join(append([]string{root}, elem...)...)
}

// Nonprofit is the path of the fictional nonprofit example agreement.
func Nonprofit() string {
	return ExamplePath("nonprofit-impact-measurement", "agreement.yaml")
}
