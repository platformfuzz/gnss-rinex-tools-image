package check_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/platformfuzz/gnss-rinex-tools-image/internal/check"
)

func TestAnalyzeFixture(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "minimal_obs_3.05.rnx")
	r, err := check.AnalyzeFile(path)
	if err != nil {
		t.Fatalf("AnalyzeFile: %v", err)
	}
	if r.FailReason != "" {
		t.Fatalf("unexpected fail: %s", r.FailReason)
	}
	if !strings.HasPrefix(strings.TrimLeft(r.VersionLine, " "), "3.") {
		t.Fatalf("version_line=%q", r.VersionLine)
	}
	if r.Epochs != 3 {
		t.Fatalf("epochs=%d want 3", r.Epochs)
	}
	if r.ObsLinesBySystem["G"] < 1 {
		t.Fatalf("expected GPS obs lines, got %#v", r.ObsLinesBySystem)
	}
	if r.OK {
		t.Fatal("minimal fixture should be WEAK/CHECK (small file)")
	}
	if check.ExitCode(r, nil) != 1 {
		t.Fatalf("exit=%d want 1", check.ExitCode(r, nil))
	}

	var buf bytes.Buffer
	check.WriteReport(&buf, r)
	out := buf.String()
	for _, want := range []string{"file:", "RESULT: WEAK/CHECK", "epochs: 3"} {
		if !strings.Contains(out, want) {
			t.Fatalf("report missing %q:\n%s", want, out)
		}
	}
}

func TestAnalyzeNotRinex3(t *testing.T) {
	text := "     2.11           OBSERVATION DATA    G                   RINEX VERSION / TYPE\n" +
		"                                                            END OF HEADER\n"
	r := check.Analyze(text, int64(len(text)), "old.rnx")
	if r.FailReason != "not RINEX 3.x" {
		t.Fatalf("fail=%q", r.FailReason)
	}
	if check.ExitCode(r, nil) != 1 {
		t.Fatalf("exit=%d", check.ExitCode(r, nil))
	}
}

func TestAnalyzeMissingEndOfHeader(t *testing.T) {
	text := "     3.05           OBSERVATION DATA    M                   RINEX VERSION / TYPE\n"
	r := check.Analyze(text, int64(len(text)), "bad.rnx")
	if r.FailReason != "no END OF HEADER" {
		t.Fatalf("fail=%q", r.FailReason)
	}
}

func TestAnalyzeFileMissing(t *testing.T) {
	_, err := check.AnalyzeFile(filepath.Join(t.TempDir(), "nope.rnx"))
	if err == nil {
		t.Fatal("expected error")
	}
	if check.ExitCode(nil, err) != 2 {
		t.Fatalf("exit=%d", check.ExitCode(nil, err))
	}
}

func TestAnalyzeOKThreshold(t *testing.T) {
	var b strings.Builder
	b.WriteString("     3.05           OBSERVATION DATA    M                   RINEX VERSION / TYPE\n")
	b.WriteString("AVLN                                                        MARKER NAME\n")
	b.WriteString("                                                            END OF HEADER\n")
	for i := 0; i < 60; i++ {
		b.WriteString("> 2026 09 18 18 16 36.0000000  0  1\n")
		b.WriteString("G01  23456789.012 1 123456.789 1   -1234.567 1  45.000    \n")
	}
	// Pad to >100 KiB so size threshold passes.
	for b.Len() <= 100_000 {
		b.WriteByte("\n"[0])
	}
	text := b.String()
	r := check.Analyze(text, int64(len(text)), "big.rnx")
	if r.FailReason != "" {
		t.Fatalf("fail=%s", r.FailReason)
	}
	if !r.OK {
		t.Fatalf("expected OK epochs=%d size=%d systems=%v", r.Epochs, r.SizeBytes, r.ObsLinesBySystem)
	}
	if check.ExitCode(r, nil) != 0 {
		t.Fatalf("exit=%d", check.ExitCode(r, nil))
	}
}

func TestWriteReportFail(t *testing.T) {
	r := check.Analyze("", 0, "empty.rnx")
	var buf bytes.Buffer
	check.WriteReport(&buf, r)
	if !strings.Contains(buf.String(), "FAIL:") {
		t.Fatalf("report=%s", buf.String())
	}
}

func TestFixtureExists(t *testing.T) {
	path := filepath.Join("..", "..", "fixtures", "minimal_obs_3.05.rnx")
	if _, err := os.Stat(path); err != nil {
		t.Fatal(err)
	}
}
