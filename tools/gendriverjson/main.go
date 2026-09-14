// Command gendriverjson writes the driver.json metadata file a custom-installed driver archive
// needs at its root (see the "Install as a custom driver" section in the README), generated from
// the exact same DriverMetadata NewDenonAVRClient sets on its Integration at runtime - so the two
// can never drift apart.
//
// Run from the repository root:
//
//	go run ./tools/gendriverjson -out driver.json
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/splattner/goucrt/pkg/integration"
	denonavrclient "github.com/splattner/remotetwo-integration-denonavr/pkg/clients/denonavr"
)

func main() {
	out := flag.String("out", "driver.json", "output file path")
	flag.Parse()

	i, err := integration.NewIntegration(integration.Config{})
	if err != nil {
		fatalf("NewIntegration: %v", err)
	}

	// NewDenonAVRClient only builds its DriverMetadata and registers function pointers - it
	// doesn't start any network activity, so it's safe to call here without InitClient/Run.
	denonavrclient.NewDenonAVRClient(i)

	if i.Metadata == nil {
		fatalf("NewDenonAVRClient didn't set driver metadata")
	}

	data, err := json.MarshalIndent(i.Metadata, "", "  ")
	if err != nil {
		fatalf("marshal driver metadata: %v", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(*out, data, 0o644); err != nil {
		fatalf("write %s: %v", *out, err)
	}

	fmt.Printf("wrote %s\n", *out)
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
