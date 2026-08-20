// Command zpecs starts the zpecs CLI binary.
package main

import (
	"os"

	"github.com/zon/specs/internal/cli"
)

func main() {
	os.Exit(cli.Main(os.Args[1:]))
}
