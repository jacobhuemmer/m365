package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// files writes the gate's three inputs into a temp dir.
func files(t *testing.T, cover, cyclo, baseline string) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	paths := []string{filepath.Join(dir, "func.txt"), filepath.Join(dir, "cyclo.txt"), filepath.Join(dir, "baseline.txt")}
	for i, body := range []string{cover, cyclo, baseline} {
		if err := os.WriteFile(paths[i], []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return paths[0], paths[1], paths[2]
}

const (
	testCover = module + "/a.go:1:\tOK\t100.0%\n" + module + "/b.go:1:\tBig\t0.0%\n"
	testCyclo = "3 p OK a.go:1:1\n4 p Big b.go:1:1\n"
)

func TestRunPassesWithBaseline(t *testing.T) {
	cover, cyclo, base := files(t, testCover, testCyclo, "# waiver\n20.0 b.go Big\n")
	var out, errw bytes.Buffer
	code := run([]string{"-module", module, "-cover", cover, "-cyclo", cyclo, "-baseline", base}, &out, &errw)
	if code != 0 || out.String() != "crap: ok (2 functions, 1 in the baseline)\n" || errw.Len() != 0 {
		t.Fatalf("code %d stdout %q stderr %q", code, out.String(), errw.String())
	}
}

func TestRunReportsViolations(t *testing.T) {
	cover, cyclo, base := files(t, testCover, testCyclo, "5.0 a.go OK\n10.0 gone.go Old\n")
	var out, errw bytes.Buffer
	code := run([]string{"-module", module, "-cover", cover, "-cyclo", cyclo, "-baseline", base}, &out, &errw)
	want := "crap: a.go OK now scores 3.0 (baseline 5.0); remove it from " + base + "\n" +
		"crap: b.go Big scores 20.0 (> 15); split it or cover it with tests\n" +
		"crap: gone.go Old is in the baseline but no longer exists; remove it from " + base + "\n"
	if code != 1 || errw.String() != want {
		t.Fatalf("code %d stderr\n%s\nwant\n%s", code, errw.String(), want)
	}
}

func TestRunReportsRise(t *testing.T) {
	cover, cyclo, base := files(t, testCover, testCyclo, "18.0 b.go Big\n")
	var out, errw bytes.Buffer
	code := run([]string{"-module", module, "-cover", cover, "-cyclo", cyclo, "-baseline", base}, &out, &errw)
	if want := "crap: b.go Big rose to 20.0 (baseline 18.0); bring it back down\n"; code != 1 || errw.String() != want {
		t.Fatalf("code %d stderr %q want %q", code, errw.String(), want)
	}
}

func TestRunPrintOffenders(t *testing.T) {
	cover, cyclo, _ := files(t, testCover, testCyclo, "")
	var out, errw bytes.Buffer
	code := run([]string{"-module", module, "-cover", cover, "-cyclo", cyclo, "-print-offenders"}, &out, &errw)
	if code != 0 || out.String() != "20.0 b.go Big\n" {
		t.Fatalf("code %d stdout %q stderr %q", code, out.String(), errw.String())
	}
}

func TestRunInputErrorsExitTwo(t *testing.T) {
	var out, errw bytes.Buffer
	if code := run([]string{"-cover", "/nonexistent", "-cyclo", "/nonexistent"}, &out, &errw); code != 2 || errw.Len() == 0 {
		t.Fatalf("code %d stderr %q", code, errw.String())
	}
	if code := run([]string{"-unknown-flag"}, &out, &errw); code != 2 {
		t.Fatalf("bad flag: code %d", code)
	}
}
