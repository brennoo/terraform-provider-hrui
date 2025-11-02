package provider

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/recorder"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// findProjectRoot walks up the directory tree to find the project root
// (where go.mod is located).
func findProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root directory without finding go.mod
			return wd, nil // Fallback to current directory
		}
		dir = parent
	}
}

// TestAccProtoV6ProviderFactories is the primary helper.
// It returns a map of provider factories configured to use a VCR cassette.
func TestAccProtoV6ProviderFactories(t *testing.T, cassetteName string) map[string]func() (tfprotov6.ProviderServer, error) {
	// Get the VCR-backed HTTP client
	vcrClient := newVCRClient(t, cassetteName)

	return map[string]func() (tfprotov6.ProviderServer, error){
		"hrui": func() (tfprotov6.ProviderServer, error) {
			// Create a provider instance configured for testing,
			// injecting the VCR client.
			// This relies on the `NewForTest` constructor we added.
			p := NewForTest("test", vcrClient)
			return providerserver.NewProtocol6WithError(p)()
		},
	}
}

// newVCRClient creates an *http.Client configured to use a VCR cassette.
// The cassette path is dynamically determined by the HRUI_FW_VERSION env var.
func newVCRClient(t *testing.T, cassetteName string) *http.Client {
	// 1. Find project root directory (where go.mod is located)
	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	// 2. Determine firmware version from env var, default to "v1.9"
	fwVersion := os.Getenv("HRUI_FW_VERSION")
	if fwVersion == "" {
		fwVersion = "v1.9" // Default to our v1.9 cassettes
	}

	// 3. Construct the full cassette path relative to project root
	// Example: <projectRoot>/internal/testdata/cassettes/v1.9/eee_resource_test
	cassetteDir := filepath.Join(projectRoot, "internal", "testdata", "cassettes", fwVersion)
	cassettePath := filepath.Join(cassetteDir, cassetteName)

	// 4. Determine VCR mode (replay, record, or passthrough)
	// Default to "replay" to fail if cassettes are missing.
	// Set VCR_MODE=record to record new cassettes.
	vcrModeEnv := os.Getenv("VCR_MODE")
	var r *recorder.Recorder

	switch vcrModeEnv {
	case "record":
		// Ensure the directory exists before recording
		if err := os.MkdirAll(cassetteDir, 0o750); err != nil {
			t.Fatalf("Failed to create cassette directory %s: %v", cassetteDir, err)
		}
		// Recording mode: record new interactions
		r, err = recorder.NewAsMode(cassettePath, recorder.ModeRecording, nil)
		if err != nil {
			t.Fatalf("Failed to create VCR recorder at %s: %v", cassettePath, err)
		}
		fmt.Printf("VCR: RECORDING to %s\n", cassettePath)
	case "passthrough":
		// Passthrough mode: make live requests without recording/replaying
		fmt.Println("VCR: PASSTHROUGH (live requests)")
		return &http.Client{
			Timeout: 30 * time.Second,
		}
	default:
		// Replay mode: use existing cassette, fail if missing
		r, err = recorder.NewAsMode(cassettePath, recorder.ModeReplaying, nil)
		if err != nil {
			t.Fatalf("Failed to create VCR recorder at %s: %v", cassettePath, err)
		}
		fmt.Printf("VCR: REPLAYING from %s\n", cassettePath)
	}

	// 5. Store reference to recorder for cleanup (only if recorder was created)
	recorderInstance := r

	// 6. Automatically stop the recorder when the test finishes
	t.Cleanup(func() {
		if recorderInstance != nil {
			if err := recorderInstance.Stop(); err != nil {
				t.Logf("Failed to stop VCR recorder: %v", err)
			}
		}
	})

	// 7. Return an HTTP client that uses the recorder as its transport
	return &http.Client{
		Transport: r,
		Timeout:   30 * time.Second,
	}
}
