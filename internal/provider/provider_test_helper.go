package provider

import (
	"bufio"
	"cmp"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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
			return wd, nil
		}
		dir = parent
	}
}

// TestAccProtoV6ProviderFactories returns provider factories backed by the rtl51 firmware emulator
// (or a live HRUI_URL endpoint when explicitly provided in the environment).
func TestAccProtoV6ProviderFactories(t *testing.T, testName string) map[string]func() (tfprotov6.ProviderServer, error) {
	t.Helper()

	if os.Getenv("TF_ACC") != "" && os.Getenv("TF_ACC_TERRAFORM_PATH") == "" {
		if _, err := exec.LookPath("terraform"); err != nil {
			if _, statErr := os.Stat("/tmp/terraform"); statErr == nil {
				t.Setenv("TF_ACC_TERRAFORM_PATH", "/tmp/terraform")
			}
		}
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	targetURL := os.Getenv("HRUI_URL")
	if targetURL == "" || targetURL == "http://localhost" {
		t.Setenv("HRUI_USERNAME", "admin")
		t.Setenv("HRUI_PASSWORD", "admin")
		targetURL = startEmulatorForTest(t, testName)
	}

	return map[string]func() (tfprotov6.ProviderServer, error){
		"hrui": func() (tfprotov6.ProviderServer, error) {
			p := NewForTest("test", client, targetURL)
			return providerserver.NewProtocol6WithError(p)()
		},
	}
}

func resolveFirmwareFilename(testName string) string {
	fwEnv := strings.TrimSpace(cmp.Or(os.Getenv("HRUI_FW_IMAGE"), os.Getenv("HRUI_FW_VERSION")))
	if strings.HasPrefix(testName, "poe_") {
		// PoE endpoints (/pse_system.cgi, /pse_port.cgi) only exist on PoE switch firmware.
		return "S600WP-4GT-2SX-SEV1.9.bin"
	}
	requiresDHCP := strings.HasPrefix(testName, "dhcp_snooping")
	requiresV1 := strings.HasPrefix(testName, "eee_") || strings.HasPrefix(testName, "mac_limit_")

	switch fwEnv {
	case "", "v1.9", "v1.9.1", "v2.0", "SWTG118AS-v2.0.bin":
		if requiresDHCP {
			return "S600WP-4GT-2SX-SEV1.9.bin"
		}
		return "SWTG118AS-v2.0.bin"
	case "SWTG115AS-v2.0.bin", "v1.3":
		if requiresDHCP {
			return "S600WP-4GT-2SX-SEV1.9.bin"
		}
		return "SWTG115AS-v2.0.bin"
	case "v2.0.1", "SWTG018AS-A-v2.0.1.bin":
		if requiresV1 {
			return "SWTG118AS-v2.0.bin"
		}
		return "SWTG018AS-A-v2.0.1.bin"
	case "v200.2.5", "V200.2.5", "HRRF_V200.2.5_ZX310S-4T2XH_0422.bin":
		if requiresV1 {
			return "SWTG118AS-v2.0.bin"
		}
		return "HRRF_V200.2.5_ZX310S-4T2XH_0422.bin"
	case "poe", "hasivo", "S600WP-4GT-2SX-SEV1.9.bin":
		return "S600WP-4GT-2SX-SEV1.9.bin"
	default:
		return fwEnv
	}
}

func findEmulatorBinary(projectRoot string) (string, error) {
	if envBin := os.Getenv("HRUI_EMULATOR_BIN"); envBin != "" {
		return envBin, nil
	}
	if pathBin, err := exec.LookPath("rtl51"); err == nil {
		return pathBin, nil
	}
	candidates := []string{
		"/tmp/rtl51",
		filepath.Join(projectRoot, "rtl51"),
		filepath.Join(filepath.Dir(projectRoot), "rtl51", "rtl51"),
		// Fall back to the pre-rename checkout directory name (hruirun), since the repo
		// was renamed to rtl51 but a local clone's directory may not have been.
		filepath.Join(filepath.Dir(projectRoot), "hruirun", "rtl51"),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && !info.IsDir() {
			return c, nil
		}
	}
	return "", os.ErrNotExist
}

func startEmulatorForTest(t *testing.T, testName string) string {
	t.Helper()

	projectRoot, err := findProjectRoot()
	if err != nil {
		t.Fatalf("Failed to find project root: %v", err)
	}

	binPath, err := findEmulatorBinary(projectRoot)
	if err != nil {
		t.Fatalf("rtl51 emulator binary not found (set HRUI_EMULATOR_BIN or place rtl51 in PATH): %v", err)
	}

	fwFile := resolveFirmwareFilename(testName)
	fwPath := fwFile
	if !filepath.IsAbs(fwPath) {
		fwPath = filepath.Join(projectRoot, "internal", "testdata", "firmwares", fwFile)
	}
	if _, err := os.Stat(fwPath); err != nil {
		t.Fatalf("Firmware image not found at %s: %v", fwPath, err)
	}

	//#nosec G204 -- binPath and fwPath are controlled test harness inputs
	// nosemgrep: go.lang.security.audit.dangerous-exec-command.dangerous-exec-command, go_subproc_rule-subproc
	cmd := exec.CommandContext(t.Context(), binPath, "boot", "-http", "127.0.0.1:0", fwPath)
	pr, pw := io.Pipe()
	cmd.Stdout = pw
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		_ = pw.Close()
		_ = pr.Close()
		t.Fatalf("Failed to start rtl51 (%s): %v", binPath, err)
	}

	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		_ = pw.Close()
		_ = pr.Close()
		_ = cmd.Wait()
	})

	type bootResult struct {
		url string
		err error
	}
	readyCh := make(chan bootResult, 1)

	go func() {
		scanner := bufio.NewScanner(pr)
		var proxyURL string
		var booted bool
		for scanner.Scan() {
			line := scanner.Text()
			if !booted {
				if _, after, ok := strings.Cut(line, "[rtl51] Forwarding "); ok {
					if addr, _, ok := strings.Cut(after, " ->"); ok {
						proxyURL = strings.TrimSpace(addr)
					}
				}
				if strings.Contains(line, "Please input uboot password") || strings.Contains(line, "calibrate Pass!") {
					booted = true
					readyCh <- bootResult{url: proxyURL}
				}
			}
		}
		if !booted {
			readyCh <- bootResult{err: scanner.Err()}
		}
	}()

	select {
	case res := <-readyCh:
		if res.err != nil || res.url == "" {
			t.Fatalf("rtl51 exited before completing boot (url=%q, err=%v)", res.url, res.err)
		}
		return res.url
	case <-time.After(15 * time.Second):
		t.Fatalf("Timed out waiting for rtl51 to boot firmware %s", fwFile)
		return ""
	}
}
