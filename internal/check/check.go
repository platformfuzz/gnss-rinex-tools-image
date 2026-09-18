// Package check implements a custom quick RINEX 3 OBS sanity check
// (header + epoch count + constellation systems). It is not a wrapper
// around an external RINEX processor.
package check

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Result holds a parsed RINEX 3 OBS summary.
type Result struct {
	FileName         string
	SizeBytes        int64
	VersionLine      string
	HeaderFields     map[string]string
	Epochs           int
	FirstEpoch       string
	LastEpoch        string
	ObsLinesBySystem map[string]int
	OK               bool
	FailReason       string
}

var headerKeys = []string{
	"MARKER NAME",
	"REC # / TYPE / VERS",
	"ANT # / TYPE",
	"APPROX POSITION XYZ",
	"SYS / # / OBS TYPES",
	"TIME OF FIRST OBS",
	"TIME OF LAST OBS",
}

// AnalyzeFile reads path and returns a RINEX 3 OBS summary.
func AnalyzeFile(path string) (*Result, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("missing: %s", path)
		}
		return nil, err
	}
	// Path comes from the CLI user; intentional local file read.
	data, err := os.ReadFile(path) //nolint:gosec // G304: caller-provided RINEX path
	if err != nil {
		return nil, err
	}
	r := Analyze(string(data), info.Size(), filepath.Base(path))
	return r, nil
}

// Analyze parses RINEX OBS text. sizeBytes should match the on-disk size
// (used for the OK threshold).
func Analyze(text string, sizeBytes int64, fileName string) *Result {
	r := &Result{
		FileName:         fileName,
		SizeBytes:        sizeBytes,
		HeaderFields:     make(map[string]string),
		ObsLinesBySystem: make(map[string]int),
	}

	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if len(lines) == 1 && lines[0] == "" {
		r.FailReason = "empty file"
		return r
	}

	verLine, end := findHeaderBounds(lines)
	r.VersionLine = strings.TrimSpace(verLine)
	if verLine == "" {
		r.VersionLine = "MISSING"
	}

	trimmedVer := strings.TrimLeft(verLine, " \t")
	if verLine == "" || !strings.HasPrefix(trimmedVer, "3.") {
		r.FailReason = "not RINEX 3.x"
		return r
	}
	if end < 0 {
		r.FailReason = "no END OF HEADER"
		return r
	}

	fillHeaderFields(r, lines[:end+1])
	fillBodyStats(r, lines[end+1:])
	_, hasG := r.ObsLinesBySystem["G"]
	r.OK = sizeBytes > 100_000 && r.Epochs >= 60 && hasG
	return r
}

func findHeaderBounds(lines []string) (verLine string, end int) {
	end = -1
	for i, line := range lines {
		if strings.Contains(line, "RINEX VERSION / TYPE") {
			verLine = line
		}
		if strings.Contains(line, "END OF HEADER") {
			end = i
			break
		}
	}
	return verLine, end
}

func fillHeaderFields(r *Result, hdr []string) {
	for _, key := range headerKeys {
		for _, line := range hdr {
			if !strings.Contains(line, key) {
				continue
			}
			display := line
			if len(display) > 72 {
				display = display[:72] + "..."
			}
			r.HeaderFields[key] = display
			break
		}
	}
}

func fillBodyStats(r *Result, body []string) {
	epochs := make([]string, 0, 64)
	for _, line := range body {
		if strings.HasPrefix(line, ">") {
			epochs = append(epochs, line)
			continue
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		sys := string(line[0])
		r.ObsLinesBySystem[sys]++
	}
	r.Epochs = len(epochs)
	if len(epochs) > 0 {
		r.FirstEpoch = truncate(epochs[0], 60)
		r.LastEpoch = truncate(epochs[len(epochs)-1], 60)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

// WriteReport prints a human-readable QC summary to w.
func WriteReport(w io.Writer, r *Result) error {
	if _, err := fmt.Fprintf(w, "file: %s\n", r.FileName); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "size_bytes: %d\n", r.SizeBytes); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "version_line: %s\n", r.VersionLine); err != nil {
		return err
	}
	if r.FailReason != "" {
		_, err := fmt.Fprintf(w, "FAIL: %s\n", r.FailReason)
		return err
	}
	for _, key := range headerKeys {
		val, ok := r.HeaderFields[key]
		if !ok {
			if _, err := fmt.Fprintf(w, "hdr[%s]: None\n", key); err != nil {
				return err
			}
			continue
		}
		if _, err := fmt.Fprintf(w, "hdr[%s]: %s\n", key, val); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "epochs: %d\n", r.Epochs); err != nil {
		return err
	}
	if r.Epochs > 0 {
		if _, err := fmt.Fprintf(w, "first_epoch: %s\n", r.FirstEpoch); err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "last_epoch:  %s\n", r.LastEpoch); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "obs_lines_by_system: %s\n", formatSystems(r.ObsLinesBySystem)); err != nil {
		return err
	}
	if r.OK {
		_, err := fmt.Fprintln(w, "RESULT: OK")
		return err
	}
	_, err := fmt.Fprintln(w, "RESULT: WEAK/CHECK")
	return err
}

func formatSystems(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%q: %d", k, m[k]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

// ExitCode maps a result to a process exit code (0 OK, 1 weak/fail, 2 usage/missing).
func ExitCode(r *Result, analyzeErr error) int {
	if analyzeErr != nil {
		return 2
	}
	if r.FailReason != "" {
		return 1
	}
	if r.OK {
		return 0
	}
	return 1
}
