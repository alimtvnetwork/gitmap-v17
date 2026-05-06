package cmd

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/alimtvnetwork/gitmap-v17/gitmap/constants"
)

// maxDocsSiteSize is the maximum total extraction size for docs-site.zip (100 MB).
const maxDocsSiteSize = 100 * 1024 * 1024

// runHelpDashboard serves the docs site locally.
func runHelpDashboard(args []string) {
	checkHelp("help-dashboard", args)

	port := parseHelpDashboardFlags(args)
	binaryDir := resolveBinaryDir()
	docsDir := filepath.Join(binaryDir, constants.HDDocsDir)

	// Auto-extract docs-site.zip if docs-site/ directory doesn't exist
	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		zipPath := filepath.Join(binaryDir, constants.DocsSiteArchive)
		if _, zipErr := os.Stat(zipPath); zipErr == nil {
			fmt.Printf("  Extracting %s...\n", constants.DocsSiteArchive)
			if extractErr := extractDocsSiteZip(zipPath, binaryDir); extractErr != nil {
				fmt.Fprintf(os.Stderr, "  ✗ Failed to extract docs-site.zip: %v\n", extractErr)
				os.Exit(1)
			}
			fmt.Printf("  ✓ Docs site extracted to %s\n", docsDir)
		}
	}

	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, constants.ErrHDNoDocsDir, docsDir)
		os.Exit(1)
	}

	distDir := filepath.Join(docsDir, constants.HDDistDir)

	if info, err := os.Stat(distDir); err == nil && info.IsDir() {
		serveStatic(distDir, port)
	} else {
		fmt.Print(constants.MsgHDNoDistFallback)
		serveDev(docsDir, port)
	}
}

// extractDocsSiteZip extracts docs-site.zip into the target directory.
// Validates paths to prevent traversal (G305) and limits total size (G110).
func extractDocsSiteZip(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}
	defer r.Close()

	absTarget, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolve target dir: %w", err)
	}

	var totalSize int64

	for _, f := range r.File {
		destPath := filepath.Join(absTarget, f.Name) // #nosec G305 — validated below
		absDestPath, absErr := filepath.Abs(destPath)
		if absErr != nil || !strings.HasPrefix(absDestPath, absTarget+string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path in zip: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if mkErr := os.MkdirAll(absDestPath, constants.DirPermission); mkErr != nil {
				return fmt.Errorf("create dir %s: %w", absDestPath, mkErr)
			}
			continue
		}

		if mkErr := os.MkdirAll(filepath.Dir(absDestPath), constants.DirPermission); mkErr != nil {
			return fmt.Errorf("create parent dir: %w", mkErr)
		}

		rc, openErr := f.Open()
		if openErr != nil {
			return fmt.Errorf("open entry %s: %w", f.Name, openErr)
		}

		outFile, createErr := os.Create(absDestPath)
		if createErr != nil {
			rc.Close()
			return fmt.Errorf("create file %s: %w", absDestPath, createErr)
		}

		written, copyErr := io.CopyN(outFile, rc, maxDocsSiteSize-totalSize) // #nosec G110 — size-limited
		outFile.Close()
		rc.Close()

		if copyErr != nil && !errors.Is(copyErr, io.EOF) {
			return fmt.Errorf("write file %s: %w", absDestPath, copyErr)
		}

		totalSize += written
		if totalSize >= maxDocsSiteSize {
			return fmt.Errorf("archive exceeds maximum extraction size (%d bytes)", maxDocsSiteSize)
		}
	}

	return nil
}

// parseHelpDashboardFlags parses the --port flag.
func parseHelpDashboardFlags(args []string) int {
	fs := flag.NewFlagSet(constants.CmdHelpDashboard, flag.ExitOnError)
	port := fs.Int("port", constants.HDDefaultPort, constants.FlagDescHDPort)
	fs.Parse(args)

	return *port
}

// resolveBinaryDir returns the directory containing the gitmap binary.
func resolveBinaryDir() string {
	exe, err := os.Executable()
	if err != nil {
		return "."
	}

	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		return filepath.Dir(exe)
	}

	return filepath.Dir(resolved)
}

// serveStatic serves pre-built dist/ files over HTTP.
func serveStatic(distDir string, port int) {
	fmt.Printf(constants.MsgHDServingStatic, distDir, port)
	openBrowser(port)

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           http.FileServer(http.Dir(distDir)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go handleShutdown(server)

	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, constants.ErrHDServe, err)
		os.Exit(1)
	}

	fmt.Print(constants.MsgHDStopped)
}

// serveDev runs npm install + npm run dev as a fallback.
func serveDev(docsDir string, port int) {
	npmPath, err := exec.LookPath("npm")
	if err != nil {
		fmt.Fprint(os.Stderr, constants.ErrHDNPMNotFound)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgHDRunningNPM)

	install := exec.Command(npmPath, "install")
	install.Dir = docsDir
	install.Stdout = os.Stdout
	install.Stderr = os.Stderr

	if err := install.Run(); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrHDNPMInstall, err)
		os.Exit(1)
	}

	fmt.Printf(constants.MsgHDStartingDev, docsDir)

	dev := exec.Command(npmPath, "run", "dev", "--", "--port", fmt.Sprintf("%d", port))
	dev.Dir = docsDir
	dev.Stdout = os.Stdout
	dev.Stderr = os.Stderr

	if err := dev.Start(); err != nil {
		fmt.Fprintf(os.Stderr, constants.ErrHDDevServer, err)
		os.Exit(1)
	}

	openBrowser(port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	_ = dev.Process.Kill()
	fmt.Print(constants.MsgHDStopped)
}

// openBrowser opens the URL in the default browser.
func openBrowser(port int) {
	url := fmt.Sprintf("http://localhost:%d", port)
	fmt.Printf(constants.MsgHDOpening, port)

	var cmd *exec.Cmd

	switch runtime.GOOS {
	case constants.OSWindows:
		cmd = exec.Command(constants.CmdWindowsShell, constants.CmdArgSlashC, constants.CmdArgStart, url)
	case constants.OSDarwin:
		cmd = exec.Command(constants.CmdOpen, url)
	default:
		cmd = exec.Command(constants.CmdXdgOpen, url)
	}

	_ = cmd.Start()
}

// handleShutdown gracefully stops the static server on Ctrl+C.
func handleShutdown(server *http.Server) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	server.Close()
}
