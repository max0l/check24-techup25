package data

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

//
// --- processHotelBatch ---
//

func TestProcessHotelBatch_EmitsHotels_AndRecoversOnPanic(t *testing.T) {
	// Capture error logs for the panic recovery branch.
	var buf bytes.Buffer
	prevLogger, prevLevel := log.Logger, zerolog.GlobalLevel()
	t.Cleanup(func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	})
	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	// Good record
	recOK := dataRecord{
		id: 1111,
		data: []string{
			"101",         // Hotelid (int)
			"Blue Resort", // Hotelname
			"4.5",         // Hotelstars
		},
	}
	// Bad record: too short -> index out of range inside goroutine -> recovered
	recBad := dataRecord{id: 1112, data: []string{"BAD"}}

	ch := make(chan *Hotel, 2)

	processHotelBatch([]dataRecord{recBad, recOK}, ch)

	// We should still get the good hotel
	select {
	case h := <-ch:
		if h.Hotelid != 101 {
			t.Fatalf("Hotelid = %d; want 101", h.Hotelid)
		}
		if h.Hotelname != "Blue Resort" {
			t.Fatalf("Hotelname = %q; want %q", h.Hotelname, "Blue Resort")
		}
		if h.Hotelstars != 4.5 {
			t.Fatalf("Hotelstars = %v; want 4.5", h.Hotelstars)
		}
	default:
		t.Fatalf("expected one hotel emitted, got none")
	}

	// And we should see a recovery log for the bad record
	logs := buf.String()
	if !strings.Contains(logs, "Recovered from panic") || !strings.Contains(logs, "BAD") {
		t.Fatalf("expected recovery log mentioning record BAD; got logs:\n%s", logs)
	}
}

//
// --- hotelBatchWorker ---
//

func TestHotelBatchWorker_LogsLifecycle_AndProcessesBatches(t *testing.T) {
	// Capture info logs for lifecycle messages.
	var buf bytes.Buffer
	prevLogger, prevLevel := log.Logger, zerolog.GlobalLevel()
	t.Cleanup(func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	})
	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	workerChan := make(chan []dataRecord)
	results := make(chan *Hotel, 10)

	// Make a helper to craft valid records quickly.
	makeRec := func(id int, name string, stars string) dataRecord {
		return dataRecord{
			id: int(time.Now().UnixMilli()), // not used by processor here
			data: []string{
				// id is an int in your domain; we pass it as string for parseInt
				// (that's what processHotelBatch expects).
				strconvI(id),
				name,
				stars,
			},
		}
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go hotelBatchWorker(workerChan, results, &wg)

	// Two batches: total 3 hotels
	workerChan <- []dataRecord{
		makeRec(1, "A", "3"),
		makeRec(2, "B", "4"),
	}
	workerChan <- []dataRecord{
		makeRec(3, "C", "5"),
	}
	close(workerChan)
	wg.Wait() // worker must exit and log "Stopped ..."

	// Drain results
	count := 0
	for {
		select {
		case <-results:
			count++
		default:
			goto DONE
		}
	}
DONE:
	if count != 3 {
		t.Fatalf("expected 3 hotels emitted, got %d", count)
	}

	out := buf.String()
	if !strings.Contains(out, "Started Hotel Batch Worker goroutine") {
		t.Fatalf("missing start log; got:\n%s", out)
	}
	if !strings.Contains(out, "Stopped Hotel Batch Worker goroutine") {
		t.Fatalf("missing stop log; got:\n%s", out)
	}
}

// tiny helper because strconv.Atoi requires error handling and we know inputs
func strconvI(i int) string { return strconvIbuf(i) }
func strconvIbuf(i int) string {
	// minimal, allocation-free conversion for small ints
	return fmt.Sprintf("%d", i)
}

//
// --- PrintHotelsResults ---
//

func TestPrintHotelsResults_LogsIDs(t *testing.T) {
	var buf bytes.Buffer
	prevLogger, prevLevel := log.Logger, zerolog.GlobalLevel()
	t.Cleanup(func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	})
	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ch := make(chan *Hotel)

	go func() {
		ch <- &Hotel{Hotelid: 10}
		ch <- &Hotel{Hotelid: 20}
		ch <- &Hotel{Hotelid: 30}
		close(ch)
	}()

	PrintHotelsResults(ch)

	out := buf.String()
	for _, id := range []string{"10", "20", "30"} {
		if !strings.Contains(out, id) {
			t.Fatalf("expected log to contain id %s; got:\n%s", id, out)
		}
	}
}

func TestPrintHotelsResults_EmptyChannel_NoLogs(t *testing.T) {
	var buf bytes.Buffer
	prevLogger, prevLevel := log.Logger, zerolog.GlobalLevel()
	t.Cleanup(func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	})
	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ch := make(chan *Hotel)
	close(ch)

	PrintHotelsResults(ch)

	if s := buf.String(); s != "" {
		t.Fatalf("expected no logs for empty channel, got:\n%s", s)
	}
}
