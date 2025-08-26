package data

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// testBatchFunc turns each dataRecord into a pointer to its id and sends it out.
func testBatchFunc(intake <-chan []dataRecord, out chan<- *int, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range intake {
		for _, rec := range batch {
			v := rec.id
			out <- &v
		}
	}
}

func TestReadCSVFile_ProcessesAllRecordsAndCloses(t *testing.T) {
	t.Parallel()

	// Build a small CSV (header + N rows). Keep N << BatchSize so we exercise the final-flush path.
	const N = 25
	var sb strings.Builder
	sb.WriteString("col1,col2\n") // header
	for i := 0; i < N; i++ {
		sb.WriteString(fmt.Sprintf("%d,val-%d\n", i, i))
	}

	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "data.csv")
	if err := os.WriteFile(p, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("write tmp csv: %v", err)
	}

	f, err := os.Open(p)
	if err != nil {
		t.Fatalf("open tmp csv: %v", err)
	}
	defer f.Close()

	results := make(chan *int, 64)

	// Invoke the generic reader with T = int
	readCSVFile[int](f, testBatchFunc, ',', results)

	// Collect all results until the channel is closed by readCSVFile.
	var got []int
	for ptr := range results {
		got = append(got, *ptr)
	}

	if len(got) != N {
		t.Fatalf("expected %d results, got %d", N, len(got))
	}
	// Since we sent a single batch (< BatchSize), order should be the same as input (0..N-1).
	for i := 0; i < N; i++ {
		if got[i] != i {
			t.Fatalf("result[%d] = %d; want %d", i, got[i], i)
		}
	}
}

func TestReadCSVFile_EmptyAfterHeader(t *testing.T) {
	t.Parallel()

	csvContent := "h1,h2\n" // header only
	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "empty.csv")
	if err := os.WriteFile(p, []byte(csvContent), 0o644); err != nil {
		t.Fatalf("write tmp csv: %v", err)
	}

	f, err := os.Open(p)
	if err != nil {
		t.Fatalf("open tmp csv: %v", err)
	}
	defer f.Close()

	results := make(chan *int, 8)
	readCSVFile[int](f, testBatchFunc, ',', results)

	// Channel should close with no payloads.
	select {
	case v, ok := <-results:
		if ok {
			t.Fatalf("expected no results, got %v", *v)
		}
	default:
		// Drain via range to be certain it's closed and empty.
		count := 0
		for range results {
			count++
		}
		if count != 0 {
			t.Fatalf("expected zero results, got %d", count)
		}
	}
}

func TestOpenFile_Success(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	p := filepath.Join(tmpDir, "test.csv")
	if err := os.WriteFile(p, []byte("h1,h2\n1,2\n"), 0o644); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	f, err := openFile(p)
	if err != nil {
		t.Fatalf("openFile returned error: %v", err)
	}
	// Ensure the file is actually readable
	b := make([]byte, 2)
	n, rerr := f.Read(b)
	if rerr != nil && rerr != io.EOF {
		t.Fatalf("read after open failed: %v", rerr)
	}
	if n == 0 {
		t.Fatalf("expected to read some bytes from opened file")
	}
	closeFile(f)
}

func TestOpenFile_FatalOnMissingPath(t *testing.T) {
	t.Parallel()

	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "openfile-missing")
	cmd.Env = append(os.Environ(), "WANT_HELPER_PROCESS=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-nil error (process exit) when openFile logs fatal")
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		if exitErr.Success() {
			t.Fatalf("expected non-zero exit status")
		}
	}
}

func TestCloseFile_ClosesHandle(t *testing.T) {
	t.Parallel()

	f, err := os.CreateTemp(t.TempDir(), "close-test-*.txt")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	// Make sure it's usable before close.
	if _, err := f.Write([]byte("x")); err != nil {
		t.Fatalf("initial write failed: %v", err)
	}

	closeFile(f) // should close immediately (via defer inside function)

	// Any operation after close should error.
	if _, err := f.Write([]byte("y")); err == nil {
		t.Fatalf("expected write after close to fail, got nil error")
	}
}

func TestReadHeaders_Success(t *testing.T) {
	t.Parallel()

	content := "a,b,c\n1,2,3\n"
	r := csv.NewReader(strings.NewReader(content))

	// Should not panic or os.Exit; just reads one line and logs.
	readHeaders(r)

	// Verify the reader advanced to the next record (optional sanity check)
	rec, err := r.Read()
	if err != nil {
		t.Fatalf("expected to read the first data row after headers, got error: %v", err)
	}
	if want, got := []string{"1", "2", "3"}, rec; strings.Join(want, ",") != strings.Join(got, ",") {
		t.Fatalf("unexpected next record; want %v got %v", want, got)
	}
}

func TestReadHeaders_FatalOnError(t *testing.T) {
	t.Parallel()

	// Empty reader -> csv.Reader.Read() returns EOF, which triggers log.Fatal in readHeaders.
	cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "readheaders-eof")
	cmd.Env = append(os.Environ(), "WANT_HELPER_PROCESS=1")
	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected non-nil error (process exit) when readHeaders logs fatal")
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		if exitErr.Success() {
			t.Fatalf("expected non-zero exit status")
		}
	} else {
		t.Fatalf("expected ExitError, got %T: %v", err, err)
	}
}

func TestParseInt(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want int
	}{
		{"0", 0},
		{"42", 42},
		{"-17", -17},
		{" 7 ", 0},    // invalid (whitespace) -> 0 with error log
		{"foo", 0},    // invalid -> 0
		{"", 0},       // invalid -> 0
		{"+100", 100}, // valid
	}
	for _, tc := range tests {
		got := parseInt(tc.in)
		if got != tc.want {
			t.Errorf("parseInt(%q) = %d; want %d", tc.in, got, tc.want)
		}
	}
}
func TestParseFloat(t *testing.T) {
	t.Parallel()

	type kind int
	const (
		finite kind = iota
		nan
		pinf
		ninf
		errVal
	)

	tests := []struct {
		in   string
		want float32
		k    kind
	}{
		{"0", 0, finite},
		{"3.14", 3.14, finite},
		{"-2.5", -2.5, finite},
		{"1e2", 100, finite},
		{"NaN", 0, nan},    // valid parse -> NaN
		{"+Inf", 0, pinf},  // valid parse -> +Inf
		{"-Inf", 0, ninf},  // valid parse -> -Inf
		{"foo", 0, errVal}, // invalid -> function returns 0
		{"", 0, errVal},    // invalid -> function returns 0
	}

	eq := func(a, b float32) bool {
		d := float64(a) - float64(b)
		if d < 0 {
			d = -d
		}
		return d < 1e-6
	}

	for _, tt := range tests {
		got := parseFloat(tt.in)
		switch tt.k {
		case finite:
			if !eq(got, tt.want) {
				t.Errorf("parseFloat(%q) = %v; want %v", tt.in, got, tt.want)
			}
		case nan:
			if !math.IsNaN(float64(got)) {
				t.Errorf("parseFloat(%q) = %v; want NaN", tt.in, got)
			}
		case pinf:
			if !math.IsInf(float64(got), +1) {
				t.Errorf("parseFloat(%q) = %v; want +Inf", tt.in, got)
			}
		case ninf:
			if !math.IsInf(float64(got), -1) {
				t.Errorf("parseFloat(%q) = %v; want -Inf", tt.in, got)
			}
		case errVal:
			if got != 0 {
				t.Errorf("parseFloat(%q) = %v; want 0 on error", tt.in, got)
			}
		}
	}
}

func TestParseTime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in       string
		wantZero bool
		wantUTC  time.Time
	}{
		{"2021-03-14T15:09:26Z", false, time.Date(2021, 3, 14, 15, 9, 26, 0, time.UTC)},
		{"2021-12-31T23:59:59Z", false, time.Date(2021, 12, 31, 23, 59, 59, 0, time.UTC)},
		{"2021/03/14", true, time.Time{}},                                                   // invalid -> zero time
		{"not-a-date", true, time.Time{}},                                                   // invalid -> zero time
		{"", true, time.Time{}},                                                             // invalid -> zero time
		{"2021-03-14T01:02:03+02:00", false, time.Date(2021, 3, 13, 23, 2, 3, 0, time.UTC)}, // normalized to UTC
	}
	for _, tc := range tests {
		got := parseTime(tc.in)
		if tc.wantZero {
			if !got.IsZero() {
				t.Errorf("parseTime(%q) expected zero time, got %v", tc.in, got)
			}
			continue
		}
		if got.IsZero() {
			t.Errorf("parseTime(%q) unexpected zero time", tc.in)
			continue
		}
		if !got.Equal(tc.wantUTC) {
			t.Errorf("parseTime(%q) = %v (%s); want %v (%s)", tc.in, got, got.Location(), tc.wantUTC, tc.wantUTC.Location())
		}
	}
}

// ---- Helper subprocess to test fatal paths without killing the main test process ----

func TestHelperProcess(t *testing.T) {
	// This function is executed in a separate process when WANT_HELPER_PROCESS=1 is set.
	if os.Getenv("WANT_HELPER_PROCESS") != "1" {
		return
	}
	args := os.Args
	sep := 0
	for i, a := range args {
		if a == "--" {
			sep = i
			break
		}
	}
	if sep == 0 || sep+1 >= len(args) {
		os.Exit(2)
	}

	switch args[sep+1] {
	case "openfile-missing":
		// Call openFile with a missing path to trigger log.Fatal -> os.Exit(1)
		_, _ = openFile(filepath.Join(t.TempDir(), "definitely-does-not-exist.csv"))
	case "readheaders-eof":
		// Empty reader causes csv.Reader.Read() to return EOF -> log.Fatal
		r := csv.NewReader(bytes.NewReader(nil))
		readHeaders(r)
	default:
		os.Exit(3)
	}
}
