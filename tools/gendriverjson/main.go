// Command gendriverjson writes the driver.json metadata file a custom-installed driver archive
// needs at its root (see the "Install as a custom driver" section in the README), using
// integration.GenerateDriverJSON - see that function's doc comment for how and why.
//
// Run from the repository root:
//
//	go run ./tools/gendriverjson -out driver.json
package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/splattner/goucrt/pkg/integration"
	denonavrclient "github.com/splattner/remotetwo-integration-denonavr/pkg/clients/denonavr"
)

func main() {
	out := flag.String("out", "driver.json", "output file path")
	flag.Parse()

	data, err := integration.GenerateDriverJSON(func(i *integration.Integration) {
		denonavrclient.NewDenonAVRClient(i)
	})
	if err != nil {
		fatalf("%v", err)
	}

	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fatalf("write %s: %v", *out, err)
	}

	fmt.Printf("wrote %s\n", *out)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
