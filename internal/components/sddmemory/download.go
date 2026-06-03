package sddmemory

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/KevG1t/SpecAI/internal/system"
)

const (
	sddMemoryOwner = "KevG1t"
	sddMemoryRepo  = "sdd-memory"
	sddMemoryName  = "sdd-memory"
)

// Package-level vars for testability.
var (
	sddMemoryHTTPClient    = &http.Client{Timeout: 5 * time.Minute}
	sddMemoryGitHubBaseURL = "https://github.com"
	sddMemoryInstallDirFn  = sddMemoryInstallDir
	sddMemoryChecksumURLFn = sddMemoryChecksumURL
)

// DownloadLatestBinary fetches the latest sdd-memory release from GitHub and
// installs it to the appropriate directory for the given platform.
// It returns the full path to the installed binary.
//
// Checksum verification is mandatory: the install fails if checksums.txt is
// unavailable, if the archive is not listed, or if the digest does not match.
//
// This is the non-brew installation method for Linux and Windows.
// On macOS, brew handles sdd-memory transitively and this should not be called.
func DownloadLatestBinary(profile system.PlatformProfile) (string, error) {
	ctx := context.Background()

	// 1. Fetch the latest version tag from GitHub API.
	version, err := fetchLatestSDDMemoryVersion()
	if err != nil {
		return "", fmt.Errorf("fetch latest sdd-memory version: %w", err)
	}

	// 2. Determine binary name and archive URL.
	goos := profile.OS
	goarch := normalizeArch(runtime.GOARCH)
	assetURL := sddMemoryAssetURL(sddMemoryGitHubBaseURL, version, goos, goarch)
	archiveName := sddMemoryArchiveName(version, goos, goarch)
	checksumURL := sddMemoryChecksumURLFn(sddMemoryGitHubBaseURL, version)

	// 3. Determine install directory.
	installDir := sddMemoryInstallDirFn(goos)
	if err := os.MkdirAll(installDir, 0o755); err != nil {
		return "", fmt.Errorf("create sdd-memory install dir %q: %w", installDir, err)
	}

	// 4. Download archive to a temp dir so we can verify before extracting.
	binaryName := sddMemoryName
	if goos == "windows" {
		binaryName = sddMemoryName + ".exe"
	}
	outPath := filepath.Join(installDir, binaryName)

	tmpDir, err := os.MkdirTemp("", "specai-sdd-memory-*")
	if err != nil {
		return "", fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archiveName)
	actualDigest, err := sddMemoryDownloadToFile(ctx, assetURL, archivePath)
	if err != nil {
		return "", fmt.Errorf("download sdd-memory archive: %w", err)
	}

	// 5. Verify checksum — fail closed if checksums.txt is unavailable or mismatched.
	checksumsContent, err := sddMemoryFetchChecksums(ctx, checksumURL)
	if err != nil {
		return "", fmt.Errorf("checksum verification failed: checksums.txt unavailable: %w", err)
	}
	expectedDigest, err := sddMemoryExpectedChecksumFor(checksumsContent, archiveName)
	if err != nil {
		return "", fmt.Errorf("checksum verification failed: %w", err)
	}
	if actualDigest != expectedDigest {
		return "", fmt.Errorf("checksum mismatch for %s:\n  expected: %s\n  got:      %s",
			archiveName, expectedDigest, actualDigest)
	}

	// 6. Extract the verified binary.
	f, err := os.Open(archivePath)
	if err != nil {
		return "", fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	if strings.HasSuffix(assetURL, ".zip") {
		data, err := io.ReadAll(f)
		if err != nil {
			return "", fmt.Errorf("read zip archive: %w", err)
		}
		if err := sddMemoryExtractZipBinary(data, binaryName, outPath); err != nil {
			return "", fmt.Errorf("extract sdd-memory zip: %w", err)
		}
	} else {
		if err := sddMemoryExtractBinaryFromTarGz(f, sddMemoryName, outPath); err != nil {
			return "", fmt.Errorf("extract sdd-memory tar.gz: %w", err)
		}
	}

	return outPath, nil
}

// fetchLatestSDDMemoryVersion queries the GitHub Releases API for the latest sdd-memory
// release and returns the version string (without leading "v").
func fetchLatestSDDMemoryVersion() (string, error) {
	token := sddMemoryGitHubToken()
	version, status, err := fetchLatestSDDMemoryVersionRequest(token)
	if err == nil {
		return version, nil
	}

	if token != "" && (status == http.StatusUnauthorized || status == http.StatusForbidden) {
		version, _, retryErr := fetchLatestSDDMemoryVersionRequest("")
		if retryErr == nil {
			return version, nil
		}
	}

	return "", err
}

func fetchLatestSDDMemoryVersionRequest(token string) (string, int, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest",
		sddMemoryOwner, sddMemoryRepo)

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return "", 0, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := sddMemoryHTTPClient.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("call GitHub API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", resp.StatusCode, fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", resp.StatusCode, fmt.Errorf("decode release JSON: %w", err)
	}

	version := strings.TrimPrefix(release.TagName, "v")
	if version == "" {
		return "", resp.StatusCode, fmt.Errorf("empty tag_name in GitHub release response")
	}

	return version, resp.StatusCode, nil
}

// sddMemoryGitHubToken returns a GitHub API token from the environment, if available.
func sddMemoryGitHubToken() string {
	if t := os.Getenv("GITHUB_TOKEN"); t != "" {
		return t
	}
	return os.Getenv("GH_TOKEN")
}

// normalizeArch maps Go's runtime.GOARCH to the architecture names used in
// sdd-memory release assets.
func normalizeArch(goarch string) string {
	switch goarch {
	case "386":
		return "amd64"
	case "arm":
		return "arm64"
	default:
		return goarch
	}
}

// sddMemoryArchiveName returns the GoReleaser archive filename for the given
// version/os/arch combination.
func sddMemoryArchiveName(version, goos, goarch string) string {
	ext := ".tar.gz"
	if goos == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("%s_%s_%s_%s%s", sddMemoryRepo, version, goos, goarch, ext)
}

// sddMemoryAssetURL constructs the download URL for the sdd-memory release asset.
func sddMemoryAssetURL(baseURL, version, goos, goarch string) string {
	filename := sddMemoryArchiveName(version, goos, goarch)
	return fmt.Sprintf("%s/%s/%s/releases/download/v%s/%s",
		baseURL, sddMemoryOwner, sddMemoryRepo, version, filename)
}

// sddMemoryChecksumURL constructs the GitHub Releases URL for checksums.txt.
func sddMemoryChecksumURL(baseURL, version string) string {
	return fmt.Sprintf("%s/%s/%s/releases/download/v%s/checksums.txt",
		baseURL, sddMemoryOwner, sddMemoryRepo, version)
}

// sddMemoryDownloadToFile downloads the resource at url to outPath and returns
// the SHA256 hex digest of the downloaded content.
func sddMemoryDownloadToFile(ctx context.Context, url string, outPath string) (hexDigest string, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	resp, err := sddMemoryHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", fmt.Errorf("create dir: %w", err)
	}
	f, err := os.Create(outPath)
	if err != nil {
		return "", fmt.Errorf("create %s: %w", outPath, err)
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(io.MultiWriter(f, h), resp.Body); err != nil {
		return "", fmt.Errorf("write %s: %w", outPath, err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// sddMemoryFetchChecksums downloads checksums.txt from url and returns its content.
func sddMemoryFetchChecksums(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("build request: %w", err)
	}
	resp, err := sddMemoryHTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch checksums.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksums.txt: HTTP %d", resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read checksums.txt: %w", err)
	}
	return string(data), nil
}

// sddMemoryExpectedChecksumFor parses checksums.txt content and returns the SHA256
// hex digest for filename.
func sddMemoryExpectedChecksumFor(content, filename string) (string, error) {
	for _, line := range strings.Split(content, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 2 && fields[1] == filename {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("%q not listed in checksums.txt", filename)
}

// sddMemoryExtractZipBinary extracts the binary named binaryName from the zip data
// and writes it to outPath.
func sddMemoryExtractZipBinary(data []byte, binaryName, outPath string) error {
	zr, err := zip.NewReader(&byteReaderAt{data: data}, int64(len(data)))
	if err != nil {
		return fmt.Errorf("open zip: %w", err)
	}

	for _, f := range zr.File {
		if filepath.Base(f.Name) == binaryName && !f.FileInfo().IsDir() {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("open zip entry %q: %w", f.Name, err)
			}
			defer rc.Close()
			return sddMemoryWriteExecutable(rc, outPath)
		}
	}

	return fmt.Errorf("binary %q not found in zip archive", binaryName)
}

// sddMemoryInstallDir returns the directory where the sdd-memory binary should be installed.
func sddMemoryInstallDir(goos string) string {
	if goos == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, "sdd-memory", "bin")
	}

	// Linux/macOS: try /usr/local/bin first.
	candidate := "/usr/local/bin"
	if isWritableDir(candidate) {
		return candidate
	}

	// Fallback to ~/.local/bin.
	home, err := os.UserHomeDir()
	if err != nil {
		return "/usr/local/bin"
	}
	return filepath.Join(home, ".local", "bin")
}

// isWritableDir reports whether the directory exists and the process can write to it.
func isWritableDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}
	tmp, err := os.CreateTemp(dir, ".sdd-memory-write-test-*")
	if err != nil {
		return false
	}
	tmp.Close()
	os.Remove(tmp.Name())
	return true
}

// sddMemoryExtractBinaryFromTarGz reads a .tar.gz stream and extracts the first file
// whose base name matches binaryName, writing it to outPath.
func sddMemoryExtractBinaryFromTarGz(r io.Reader, binaryName, outPath string) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gr.Close()

	tr := tar.NewReader(gr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}

		if filepath.Base(hdr.Name) == binaryName &&
			(hdr.Typeflag == tar.TypeReg || hdr.Typeflag == tar.TypeRegA) {
			return sddMemoryWriteExecutable(tr, outPath)
		}
	}

	return fmt.Errorf("binary %q not found in archive", binaryName)
}

// sddMemoryWriteExecutable writes the content from r to outPath with executable permissions.
func sddMemoryWriteExecutable(r io.Reader, outPath string) error {
	dir := filepath.Dir(outPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create parent dir: %w", err)
	}

	tmp, err := os.CreateTemp(dir, ".sdd-memory-upgrade-*.tmp")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	tmpPath := tmp.Name()

	defer func() {
		if tmpPath != "" {
			os.Remove(tmpPath)
		}
	}()

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, 0o755); err != nil {
		return fmt.Errorf("chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, outPath); err != nil {
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, outPath, err)
	}

	tmpPath = ""
	return nil
}

// byteReaderAt implements io.ReaderAt over a byte slice.
type byteReaderAt struct {
	data []byte
}

func (b *byteReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 || int(off) >= len(b.data) {
		return 0, io.EOF
	}
	n := copy(p, b.data[off:])
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}
