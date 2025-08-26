package data

import (
	"bytes"
	"math"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func TestPrintOffersResults_LogsAllOffers(t *testing.T) {
	// Capture and isolate logging.
	var buf bytes.Buffer
	prevLogger := log.Logger
	prevLevel := zerolog.GlobalLevel()
	defer func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	}()

	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ch := make(chan *Offer)

	go func() {
		ch <- &Offer{Hotelid: 111111}
		ch <- &Offer{Hotelid: 222222}
		ch <- &Offer{Hotelid: 333333}
		close(ch)
	}()

	PrintOffersResults(ch)

	out := buf.String()
	for _, id := range []int{111111, 222222, 333333} {
		if !strings.Contains(out, strconv.Itoa(id)) {
			t.Fatalf("log output missing hotel id %q.\nOutput:\n%s", id, out)
		}
	}

	if got := strings.Count(out, "Offer:"); got != 3 {
		t.Fatalf("expected 3 offer log lines, got %d.\nOutput:\n%s", got, out)
	}
}

func TestPrintOffersResults_EmptyChannel_NoLogs(t *testing.T) {
	var buf bytes.Buffer
	prevLogger := log.Logger
	prevLevel := zerolog.GlobalLevel()
	defer func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	}()

	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	ch := make(chan *Offer)
	close(ch)

	PrintOffersResults(ch)

	time.Sleep(5 * time.Millisecond)

	if s := buf.String(); s != "" {
		t.Fatalf("expected no logs for empty channel, got:\n%s", s)
	}
}

func TestCalculateDurationInDays(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected int
	}{
		{
			name:     "Same moment",
			start:    time.Date(2025, 8, 25, 12, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 8, 25, 12, 0, 0, 0, time.UTC),
			expected: 0, // ceil(0) = 0
		},
		{
			name:     "Less than a day (round up to 1)",
			start:    time.Date(2025, 8, 25, 12, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 8, 25, 18, 0, 0, 0, time.UTC), // 6h
			expected: 1,
		},
		{
			name:     "Exactly one day",
			start:    time.Date(2025, 8, 25, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 8, 26, 0, 0, 0, 0, time.UTC),
			expected: 1,
		},
		{
			name:     "Fractional day (1.2 days, should ceil to 2)",
			start:    time.Date(2025, 8, 25, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 8, 26, 5, 0, 0, 0, time.UTC), // 29h
			expected: 2,
		},
		{
			name:     "Two full days",
			start:    time.Date(2025, 8, 25, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 8, 27, 0, 0, 0, 0, time.UTC),
			expected: 2,
		},
		{
			name:     "End before start (negative duration)",
			start:    time.Date(2025, 8, 26, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2025, 8, 25, 0, 0, 0, 0, time.UTC),
			expected: int(math.Ceil(-24.0 / 24.0)), // = -1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateDurationInDays(tt.start, tt.end)
			if got != tt.expected {
				t.Errorf("calculateDurationInDays(%s, %s) = %d; want %d",
					tt.start, tt.end, got, tt.expected)
			}
		})
	}
}
func TestOfferBatchWorker_LogsLifecycleAndExitsOnClose(t *testing.T) {
	// capture logs
	var buf bytes.Buffer
	prevLogger, prevLevel := log.Logger, zerolog.GlobalLevel()
	defer func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	}()
	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	workerChan := make(chan []dataRecord)
	results := make(chan *Offer, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go offerBatchWorker(workerChan, results, &wg)

	// send two empty batches just to exercise the range loop; processing can be no-op
	workerChan <- []dataRecord{}
	workerChan <- []dataRecord{}
	close(workerChan)

	// worker must call wg.Done()
	wg.Wait()

	out := buf.String()
	if !strings.Contains(out, "Started Offer Batch Worker goroutine") {
		t.Fatalf("missing start log, got:\n%s", out)
	}
	if !strings.Contains(out, "Stopped Offer Batch Worker goroutine") {
		t.Fatalf("missing stop log, got:\n%s", out)
	}
}

func TestOfferBatchWorker_ProcessesBatchesEndToEnd(t *testing.T) {
	// This uses the real processOfferBatch pipeline to ensure records → offers.
	workerChan := make(chan []dataRecord)
	results := make(chan *Offer, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go offerBatchWorker(workerChan, results, &wg)

	// Build two valid batches (mirrors your parseOffer expectations).
	makeRec := func(id int, out, in time.Time) dataRecord {
		return dataRecord{
			id: id,
			data: []string{
				"101",                    // Hotelid
				out.Format(time.RFC3339), // Outbounddeparturedatetime
				in.Format(time.RFC3339),  // Inbounddeparturedatetime
				"2",                      // Countadults
				"0",                      // Countchildren
				"999.99",                 // Price
				"AAA",                    // Inbounddepartureairport
				"BBB",                    // Inboundarrivalairport
				in.Add(2 * time.Hour).Format(time.RFC3339), // Inboundarrivaldatetime
				"CCC", // Outbounddepartureairport
				"DDD", // Outboundarrivalairport
				out.Add(3 * time.Hour).Format(time.RFC3339), // Outboundarrivaldatetime
				"RO",    // Mealtype
				"false", // Oceanview
				"Std",   // Roomtype
			},
		}
	}

	out1 := time.Date(2025, 8, 25, 8, 0, 0, 0, time.UTC)
	in1 := out1.Add(36 * time.Hour) // 1.5 days → ceil = 2
	out2 := time.Date(2025, 9, 1, 12, 0, 0, 0, time.UTC)
	in2 := out2.Add(48 * time.Hour) // 2 days → 2
	out3 := time.Date(2025, 9, 10, 7, 0, 0, 0, time.UTC)
	in3 := out3.Add(6 * time.Hour) // <1 day → ceil = 1

	batch1 := []dataRecord{
		makeRec(111111, out1, in1),
		makeRec(222222, out2, in2),
	}
	batch2 := []dataRecord{
		makeRec(333333, out3, in3),
	}

	workerChan <- batch1
	workerChan <- batch2
	close(workerChan)

	// wait for the worker to finish draining
	wg.Wait()

	// Collect all offers emitted
	var got []*Offer
	for {
		select {
		case o := <-results:
			got = append(got, o)
		default:
			goto DONE
		}
	}
DONE:
	if len(got) != 3 {
		t.Fatalf("expected 3 offers emitted, got %d", len(got))
	}

	// Map by id to assert durations
	m := map[int]*Offer{}
	for _, o := range got {
		m[o.Offerid] = o
	}
	if m[111111] == nil || m[222222] == nil || m[333333] == nil {
		t.Fatalf("missing one or more offers: %+v", got)
	}
	if m[111111].DurationInDays != 2 {
		t.Errorf("O1 DurationInDays = %d; want 2", m[111111].DurationInDays)
	}
	if m[222222].DurationInDays != 2 {
		t.Errorf("O2 DurationInDays = %d; want 2", m[222222].DurationInDays)
	}
	if m[333333].DurationInDays != 1 {
		t.Errorf("O3 DurationInDays = %d; want 1", m[333333].DurationInDays)
	}
}

func TestParseOffer_ValidRecord(t *testing.T) {
	// Prepare RFC3339 timestamps that match typical parseTime expectations.
	outDep := time.Date(2025, 8, 25, 8, 0, 0, 0, time.UTC)
	inDep := time.Date(2025, 9, 2, 14, 30, 0, 0, time.UTC)
	inArr := time.Date(2025, 9, 2, 17, 30, 0, 0, time.UTC)
	outArr := time.Date(2025, 8, 25, 12, 0, 0, 0, time.UTC)

	rec := dataRecord{
		id: 123,
		data: []string{
			"101",                       // Hotelid
			outDep.Format(time.RFC3339), // Outbounddeparturedatetime
			inDep.Format(time.RFC3339),  // Inbounddeparturedatetime
			"2",                         // Countadults
			"1",                         // Countchildren
			"1234.56",                   // Price
			"JFK",                       // Inbounddepartureairport
			"MUC",                       // Inboundarrivalairport
			inArr.Format(time.RFC3339),  // Inboundarrivaldatetime
			"MUC",                       // Outbounddepartureairport
			"JFK",                       // Outboundarrivalairport
			outArr.Format(time.RFC3339), // Outboundarrivaldatetime
			"HB",                        // Mealtype
			"true",                      // Oceanview
			"Suite",                     // Roomtype
		},
	}

	got := parseOffer(rec)

	if got.Offerid != 123 {
		t.Fatalf("Offerid = %q; want %q", got.Offerid, "O123")
	}
	if got.Hotelid != 101 {
		t.Errorf("Hotelid = %d; want %d", got.Hotelid, 101)
	}
	if !got.Outbounddeparturedatetime.Equal(outDep) {
		t.Errorf("Outbounddeparturedatetime = %v; want %v", got.Outbounddeparturedatetime, outDep)
	}
	if !got.Inbounddeparturedatetime.Equal(inDep) {
		t.Errorf("Inbounddeparturedatetime = %v; want %v", got.Inbounddeparturedatetime, inDep)
	}
	if got.Countadults != 2 {
		t.Errorf("Countadults = %d; want %d", got.Countadults, 2)
	}
	if got.Countchildren != 1 {
		t.Errorf("Countchildren = %d; want %d", got.Countchildren, 1)
	}
	if got.Price != 1234.56 {
		t.Errorf("Price = %v; want %v", got.Price, 1234.56)
	}
	if got.Inbounddepartureairport != "JFK" {
		t.Errorf("Inbounddepartureairport = %q; want %q", got.Inbounddepartureairport, "JFK")
	}
	if got.Inboundarrivalairport != "MUC" {
		t.Errorf("Inboundarrivalairport = %q; want %q", got.Inboundarrivalairport, "MUC")
	}
	if !got.Inboundarrivaldatetime.Equal(inArr) {
		t.Errorf("Inboundarrivaldatetime = %v; want %v", got.Inboundarrivaldatetime, inArr)
	}
	if got.Outbounddepartureairport != "MUC" {
		t.Errorf("Outbounddepartureairport = %q; want %q", got.Outbounddepartureairport, "MUC")
	}
	if got.Outboundarrivalairport != "JFK" {
		t.Errorf("Outboundarrivalairport = %q; want %q", got.Outboundarrivalairport, "JFK")
	}
	if !got.Outboundarrivaldatetime.Equal(outArr) {
		t.Errorf("Outboundarrivaldatetime = %v; want %v", got.Outboundarrivaldatetime, outArr)
	}
	if got.Mealtype != "HB" {
		t.Errorf("Mealtype = %q; want %q", got.Mealtype, "HB")
	}
	if got.Oceanview != true {
		t.Errorf("Oceanview = %v; want %v", got.Oceanview, true)
	}
	if got.Roomtype != "Suite" {
		t.Errorf("Roomtype = %q; want %q", got.Roomtype, "Suite")
	}
}

func TestParseOffer_OceanviewBool(t *testing.T) {
	cases := []struct {
		name     string
		flag     string
		expected bool
	}{
		{"true string", "true", true},
		{"false string", "false", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := dataRecord{
				id: 0,
				data: []string{
					"1",
					time.Now().UTC().Format(time.RFC3339),
					time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339),
					"2",
					"0",
					"10.0",
					"AAA",
					"BBB",
					time.Now().UTC().Add(72 * time.Hour).Format(time.RFC3339),
					"CCC",
					"DDD",
					time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
					"RO",
					tc.flag, // Oceanview
					"Std",
				},
			}
			got := parseOffer(rec)
			if got.Oceanview != tc.expected {
				t.Fatalf("Oceanview = %v; want %v (flag %q)", got.Oceanview, tc.expected, tc.flag)
			}
		})
	}
}

func TestParseOffer_PanicsOnShortRecord(t *testing.T) {
	rec := dataRecord{
		id:   -1,
		data: []string{"too", "short"},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic for short record, got none")
		}
	}()

	_ = parseOffer(rec)
}

func TestProcessOfferBatch_SendsOffersAndComputesDurations(t *testing.T) {
	// Arrange: timestamps
	out1 := time.Date(2025, 8, 25, 8, 0, 0, 0, time.UTC)
	in1 := out1.Add(36 * time.Hour) // 1.5 days => ceil = 2
	out2 := time.Date(2025, 9, 1, 12, 0, 0, 0, time.UTC)
	in2 := out2.Add(48 * time.Hour) // 2 days => 2

	recs := []dataRecord{
		{
			id: 1,
			data: []string{
				"101",
				out1.Format(time.RFC3339),
				in1.Format(time.RFC3339),
				"2",
				"0",
				"999.0",
				"AAA",
				"BBB",
				in1.Add(3 * time.Hour).Format(time.RFC3339),
				"CCC",
				"DDD",
				out1.Add(4 * time.Hour).Format(time.RFC3339),
				"RO",
				"false",
				"Std",
			},
		},
		{
			id: 2,
			data: []string{
				"202",
				out2.Format(time.RFC3339),
				in2.Format(time.RFC3339),
				"1",
				"1",
				"1234.56",
				"EEE",
				"FFF",
				in2.Add(2 * time.Hour).Format(time.RFC3339),
				"GGG",
				"HHH",
				out2.Add(5 * time.Hour).Format(time.RFC3339),
				"HB",
				"true",
				"Suite",
			},
		},
	}

	// Buffered channel to avoid blocking.
	ch := make(chan *Offer, len(recs))

	// Act
	processOfferBatch(recs, ch)

	// Assert: drain results
	got := []*Offer{<-ch, <-ch}

	// Map by Offerid for stable assertions.
	m := map[int]*Offer{}
	for _, o := range got {
		m[o.Offerid] = o
	}

	if m[1] == nil || m[2] == nil {
		t.Fatalf("expected offers for O1 and O2, got: %#v", got)
	}
	if m[1].DurationInDays != 2 {
		t.Errorf("O1 DurationInDays = %d; want 2", m[1].DurationInDays)
	}
	if m[2].DurationInDays != 2 {
		t.Errorf("O2 DurationInDays = %d; want 2", m[2].DurationInDays)
	}
}

func TestProcessOfferBatch_RecoversFromPanicAndLogs(t *testing.T) {
	// Capture logs.
	var buf bytes.Buffer
	prevLogger := log.Logger
	prevLevel := zerolog.GlobalLevel()
	defer func() {
		log.Logger = prevLogger
		zerolog.SetGlobalLevel(prevLevel)
	}()
	log.Logger = log.Output(&buf)
	zerolog.SetGlobalLevel(zerolog.ErrorLevel)

	validOut := time.Date(2025, 8, 25, 8, 0, 0, 0, time.UTC)
	validIn := validOut.Add(24 * time.Hour)

	recs := []dataRecord{
		{
			//Bad
			id:   66666,
			data: []string{"only1"},
		},
		{
			//Good
			id: 12314,
			data: []string{
				"303",
				validOut.Format(time.RFC3339),
				validIn.Format(time.RFC3339),
				"2",
				"0",
				"10.0",
				"AAA", "BBB",
				validIn.Add(2 * time.Hour).Format(time.RFC3339),
				"CCC", "DDD",
				validOut.Add(3 * time.Hour).Format(time.RFC3339),
				"RO", "false", "Std",
			},
		},
	}

	ch := make(chan *Offer, len(recs))

	// Act
	processOfferBatch(recs, ch)

	select {
	case o := <-ch:
		if o.Offerid != 12314 {
			t.Fatalf("expected GOOD offer, got %+v", o)
		}
	default:
		t.Fatalf("expected one offer emitted after recovery, got none")
	}

	logs := buf.String()
	if !strings.Contains(logs, "Recovered from panic") || !strings.Contains(logs, strconv.Itoa(66666)) {
		t.Fatalf("expected recovery log mentioning record BAD; got logs:\n%s", logs)
	}
}
