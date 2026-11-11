// Command cespec validates, lints, formats, and renders consulting
// engagement agreements, and exports them to DOCX or PDF through Pandoc.
package main

import (
	"os"

	"github.com/ProductBuildersHQ/consulting-engagement-spec/internal/cli"
)

func main() {
	if err := cli.NewRoot().Execute(); err != nil {
		os.Exit(1)
	}
}
