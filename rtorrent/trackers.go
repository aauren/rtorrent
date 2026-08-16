package rtorrent

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// trackerListMultiCall is used in methods which retrieve a list of trackers along with subsequent commands to call on each
	// See: https://rtorrent-docs.readthedocs.io/en/latest/cmd-ref.html#download-items-and-attributes for more info
	trackerListMultiCall = "t.multicall"

	// unknownStr is returned by TrackerEvent.String and TrackerType.String for values with no known name
	unknownStr = "Unknown"

	// noFieldStr and noValueStr are returned by GetFieldValueAsString for unknown fields and unreadable values respectively
	noFieldStr = "<ne>"
	noValueStr = "<na>"
)

var (
	// Custom error definitions
	ErrNilTrackerIndex   = errors.New("nil tracker index")
	ErrNoField           = errors.New("no field found")
	ErrUnknownField      = errors.New("unknown field")
	ErrBadData           = errors.New("bad data")
	ErrNoDataFromTracker = errors.New("no data from tracker")
	ErrMultipleTrackers  = errors.New("multiple trackers returned")
)

// XMLRPC Tracker Fields
const (
	FieldCanScrape      = TrackerField("can_scrape")
	FieldIsUsable       = TrackerField("is_usable")
	FieldIsEnabled      = TrackerField("is_enabled")
	FieldFailedCounter  = TrackerField("failed_counter")
	FieldActivityLast   = TrackerField("activity_time_last")
	FieldActivityNext   = TrackerField("activity_time_next")
	FieldFailedLast     = TrackerField("failed_time_last")
	FieldFailedNext     = TrackerField("failed_time_next")
	FieldID             = TrackerField("id")
	FieldIsBusy         = TrackerField("is_busy")
	FieldIsOpen         = TrackerField("is_open")
	FiledIsExtraTracker = TrackerField("is_extra_tracker")
	FieldLatestEvent    = TrackerField("latest_event")
	FieldMinInterval    = TrackerField("min_interval")
	FieldNormalInterval = TrackerField("normal_interval")
	FieldSuccessCounter = TrackerField("success_counter")
	FieldSuccessLast    = TrackerField("success_time_last")
	FieldSuccessNext    = TrackerField("success_time_next")
	FieldType           = TrackerField("type")
	FieldURL            = TrackerField("url")
)

// Tracker Events
const (
	EventNone TrackerEvent = iota
	EventCompleted
	EventStarted
	EventStopped
	EventScrape
)

// Tracker Types
const (
	TypeHTTP TrackerType = iota + 1
	TypeUDP
	TypeDHT
)

// fieldStringers maps every retrievable field to a renderer. Its key set is the authoritative list of valid fields
// (see AllTrackerFields) so the two can't drift apart.
var fieldStringers = map[TrackerField]func(*Tracker) (string, error){
	FieldCanScrape:      stringerFor((*Tracker).CanScrape, strconv.FormatBool),
	FieldIsUsable:       stringerFor((*Tracker).IsUsable, strconv.FormatBool),
	FieldIsEnabled:      stringerFor((*Tracker).IsEnabled, strconv.FormatBool),
	FieldFailedCounter:  stringerFor((*Tracker).FailedCounter, strconv.Itoa),
	FieldActivityLast:   stringerFor((*Tracker).ActivityLastTime, time.Time.String),
	FieldActivityNext:   stringerFor((*Tracker).ActivityTimeNext, time.Time.String),
	FieldFailedLast:     stringerFor((*Tracker).FailedTimeLast, time.Time.String),
	FieldFailedNext:     stringerFor((*Tracker).FailedTimeNext, time.Time.String),
	FieldID:             stringerFor((*Tracker).ID, identity),
	FieldIsBusy:         stringerFor((*Tracker).IsBusy, strconv.FormatBool),
	FieldIsOpen:         stringerFor((*Tracker).IsOpen, strconv.FormatBool),
	FiledIsExtraTracker: stringerFor((*Tracker).IsExtraTracker, strconv.FormatBool),
	FieldLatestEvent:    stringerFor((*Tracker).LatestEvent, TrackerEvent.String),
	FieldMinInterval:    stringerFor((*Tracker).MinInterval, strconv.Itoa),
	FieldNormalInterval: stringerFor((*Tracker).NormalInterval, strconv.Itoa),
	FieldSuccessCounter: stringerFor((*Tracker).SuccessCounter, strconv.Itoa),
	FieldSuccessLast:    stringerFor((*Tracker).SuccessTimeLast, time.Time.String),
	FieldSuccessNext:    stringerFor((*Tracker).SuccessTimeNext, time.Time.String),
	FieldType:           stringerFor((*Tracker).Type, TrackerType.String),
	FieldURL:            stringerFor((*Tracker).URL, identity),
}

// AllTrackerFields returns every retrievable tracker field, sorted. We hand back a fresh slice each call so callers
// can't mutate the package's own view of what a valid field is.
func AllTrackerFields() []TrackerField {
	return slices.Sorted(maps.Keys(fieldStringers))
}

// identity satisfies the format argument of stringerFor for string-valued fields, since Go has no builtin
func identity(s string) string { return s }

// stringerFor adapts a Tracker getter into the signature fieldStringers wants, deferring to format for the rendering
func stringerFor[T any](get func(*Tracker) (T, error), format func(T) string) func(*Tracker) (string, error) {
	return func(t *Tracker) (string, error) {
		v, err := get(t)
		if err != nil {
			return "", err
		}
		return format(v), nil
	}
}

// trackerField looks f up in the tracker's data and converts it with conv, the shape every getter below follows
func trackerField[T any](t *Tracker, f TrackerField, conv func(any) (T, error)) (T, error) {
	data, ok := t.tData[f]
	if !ok {
		var zero T
		return zero, ErrNoField
	}
	return conv(data)
}

// enumFromAny converts a raw value into any int-backed tracker enum, which covers both TrackerEvent and TrackerType
func enumFromAny[T ~int](data any) (T, error) {
	v, err := intFromAny(data)
	if err != nil {
		return 0, err
	}
	return T(v), nil
}

// TrackerService is used to interact with the tracker information gatherer methods in rTorrent
type TrackerService struct {
	C Client
}

// TrackerIndex is used to specify which tracker to retrieve information about
type TrackerIndex struct {
	InfoHash string
	Index    int
}

// String returns the string representation of the TrackerIndex. If index is -1, it will only return the infoHash, otherwise it will
// return the infoHash and index joined by a colon
func (ti *TrackerIndex) String() string {
	if ti.Index == -1 {
		return ti.InfoHash
	}
	return ti.InfoHash + ":" + strconv.Itoa(ti.Index)
}

// TrackerField is used to specify tracker related fields that can be retrieved from rTorrent
type TrackerField string

func (tf TrackerField) AsXMLRPCArgument() string {
	return "t." + string(tf) + "="
}

func (tf TrackerField) String() string {
	return string(tf)
}

// TrackerEvent is used to specify the type of event that occurred with a tracker
type TrackerEvent int

// String returns the string representation of the TrackerEvent
func (te TrackerEvent) String() string {
	switch te {
	case EventNone:
		return "None"
	case EventCompleted:
		return "Completed"
	case EventStarted:
		return "Started"
	case EventStopped:
		return "Stopped"
	case EventScrape:
		return "Scrape"
	default:
		return unknownStr
	}
}

// TrackerType is used to specify the type of tracker
type TrackerType int

// String returns the string representation of the TrackerType
func (tt TrackerType) String() string {
	switch tt {
	case TypeHTTP:
		return "HTTP"
	case TypeUDP:
		return "UDP"
	case TypeDHT:
		return "DHT"
	default:
		return unknownStr
	}
}

// Tracker is used to represent information about a tracker in rTorrent
type Tracker struct {
	ti    *TrackerIndex
	tData map[TrackerField]any
}

func (t *Tracker) CloneWithTrackerIndex(ti *TrackerIndex) *Tracker {
	return &Tracker{ti: ti, tData: t.tData}
}

// GetFieldValueAsString renders the value of f as a string, returning "<ne>" if the field isn't one we know about and
// "<na>" if it is known but couldn't be read off this particular tracker
func (t *Tracker) GetFieldValueAsString(f TrackerField) string {
	stringer, ok := fieldStringers[f]
	if !ok {
		return noFieldStr
	}
	str, err := stringer(t)
	if err != nil {
		return noValueStr
	}
	return str
}

func (t *Tracker) String() string {
	var sb strings.Builder
	// We sort the keys because ranging a map directly would shuffle the field order between calls
	for i, k := range slices.Sorted(maps.Keys(t.tData)) {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(k.String())
		sb.WriteString(": ")
		sb.WriteString(t.GetFieldValueAsString(k))
	}

	return fmt.Sprintf("Tracker: TrackerIndex: <%s>, data: <%s>", t.ti, sb.String())
}

// TrackerIndex returns the TrackerIndex for the tracker
func (t *Tracker) TrackerIndex() *TrackerIndex {
	return t.ti
}

// CanScrape Checks if the announce URL is scrapeable. rTorrent considers a HTTP tracker scrapeable if the announce URL contains the string
// /announce somewhere after the rightmost / (inclusively).
func (t *Tracker) CanScrape() (bool, error) {
	return trackerField(t, FieldCanScrape, boolFromAny)
}

// IsUsable Checks if the tracker is usable. A tracker is considered usable if it is enabled and not marked as failed.
func (t *Tracker) IsUsable() (bool, error) {
	return trackerField(t, FieldIsUsable, boolFromAny)
}

// IsEnabled Checks if the tracker is enabled. A tracker is considered enabled if it is not marked as disabled.
func (t *Tracker) IsEnabled() (bool, error) {
	return trackerField(t, FieldIsEnabled, boolFromAny)
}

// FailedCounter Returns the number of failed requests to the tracker. Note that this value resets to 0 if a request succeeds.
func (t *Tracker) FailedCounter() (int, error) {
	return trackerField(t, FieldFailedCounter, intFromAny)
}

// ActivityLastTime Returns the last time there was an attempt to announce to this tracker, regardless of whether or not the announce
// succeeded.
func (t *Tracker) ActivityLastTime() (time.Time, error) {
	return trackerField(t, FieldActivityLast, timeFromAny)
}

// ActivityTimeNext Returns when rtorrent will attempt to announce to the tracker next. In most cases, t.activity_time_next -
// t.activity_time_last will equal t.normal_interval.
func (t *Tracker) ActivityTimeNext() (time.Time, error) {
	return trackerField(t, FieldActivityNext, timeFromAny)
}

// FailedTimeLast Returns the last time there was a failed attempt to announce to this tracker.
func (t *Tracker) FailedTimeLast() (time.Time, error) {
	return trackerField(t, FieldFailedLast, timeFromAny)
}

// FailedTimeNext Returns the time at when the next request is planned to happen after a failed request. rTorrent backs off failed requests
// exponentially, i.e. each time a request fails, it doubles the interval until it tries again.
func (t *Tracker) FailedTimeNext() (time.Time, error) {
	return trackerField(t, FieldFailedNext, timeFromAny)
}

// ID If a previous HTTP tracker response contains the tracker id key, t.id will contain that value, and it will be added as a parameter to
// any subsequent requests to that same tracker.
func (t *Tracker) ID() (string, error) {
	return trackerField(t, FieldID, stringFromAny)
}

// IsBusy Returns true if the request is in the middle of processing, and false otherwise (this is identical to IsOpen())
func (t *Tracker) IsBusy() (bool, error) {
	return trackerField(t, FieldIsBusy, boolFromAny)
}

// IsOpen Returns true if the request is in the middle of processing, and false otherwise (this is identical to IsBusy())
func (t *Tracker) IsOpen() (bool, error) {
	return trackerField(t, FieldIsOpen, boolFromAny)
}

// IsExtraTracker Returns true if the tracker was added via d.tracker.insert, rather than existing in the original metafile.
func (t *Tracker) IsExtraTracker() (bool, error) {
	return trackerField(t, FiledIsExtraTracker, boolFromAny)
}

// LatestEvent Returns the latest event that occurred with the tracker, one of the Event constants. EventScrape is not
// an event key the BitTorrent spec defines, it means the tracker is currently processing a scrape request.
func (t *Tracker) LatestEvent() (TrackerEvent, error) {
	return trackerField(t, FieldLatestEvent, enumFromAny[TrackerEvent])
}

// MinInterval Returns the values for the minimum announce intervals as returned from the tracker request.
func (t *Tracker) MinInterval() (int, error) {
	return trackerField(t, FieldMinInterval, intFromAny)
}

// NormalInterval Returns the values for the normal announce intervals as returned from the tracker request.
func (t *Tracker) NormalInterval() (int, error) {
	return trackerField(t, FieldNormalInterval, intFromAny)
}

// SuccessCounter Returns the number of successful requests to the tracker.
func (t *Tracker) SuccessCounter() (int, error) {
	return trackerField(t, FieldSuccessCounter, intFromAny)
}

// SuccessTimeLast Returns the last time there was a successful attempt to announce to this tracker.
func (t *Tracker) SuccessTimeLast() (time.Time, error) {
	return trackerField(t, FieldSuccessLast, timeFromAny)
}

// SuccessTimeNext Returns the time at when the next request is planned to happen after a successful request.
func (t *Tracker) SuccessTimeNext() (time.Time, error) {
	return trackerField(t, FieldSuccessNext, timeFromAny)
}

// Type Returns the type of the tracker, one of the Type constants
func (t *Tracker) Type() (TrackerType, error) {
	return trackerField(t, FieldType, enumFromAny[TrackerType])
}

func (t *Tracker) URL() (string, error) {
	return trackerField(t, FieldURL, stringFromAny)
}

// NewTrackerNoIndex creates a new trackerIndex with no index specification, meaning that all trackers for the given infoHash will be
// executed upon
func NewTrackerNoIndex(infoHash string) *TrackerIndex {
	return &TrackerIndex{InfoHash: infoHash, Index: -1}
}

// NewTrackerWithIndex creates a new trackerIndex with an index specification, meaning that only the tracker at the given index for the
// given infoHash will be executed upon
func NewTrackerWithIndex(infoHash string, index int) *TrackerIndex {
	return &TrackerIndex{InfoHash: infoHash, Index: index}
}

// TrackerWithDetails retrieves a download's trackers along with the requested detail fields. A nil ti gives back
// ErrNilTrackerIndex and no trackers, while any other error still returns the trackers populated as far as they got.
func (ts *TrackerService) TrackerWithDetails(ctx context.Context, ti *TrackerIndex, fields []TrackerField) ([]*Tracker, error) {
	if ti == nil {
		return nil, ErrNilTrackerIndex
	}
	t := Tracker{ti: ti, tData: make(map[TrackerField]any)}
	tSlice := []*Tracker{&t}
	newCmds := []string{ti.String()}
	for _, field := range fields {
		if _, ok := fieldStringers[field]; !ok {
			return tSlice, fmt.Errorf("%w: %s", ErrUnknownField, field)
		}
		newCmds = append(newCmds, field.AsXMLRPCArgument())
	}
	sliceOfSlices, err := ts.contextWrapGetSliceSliceByHash(ctx, trackerListMultiCall, newCmds...)
	if err != nil {
		return tSlice, err
	}

	tSlice = make([]*Tracker, len(sliceOfSlices))
	for i, slice := range sliceOfSlices {
		// A multicall on a specific index returns one tracker, so we synthesize indexes only for the whole list
		idx := ti
		if len(sliceOfSlices) > 1 {
			idx = NewTrackerWithIndex(ti.InfoHash, i)
		}
		tSlice[i] = &Tracker{ti: idx}

		tData, err := TrackerDataFromSlice(fields, slice)
		if err != nil {
			return tSlice, err
		}
		tSlice[i].tData = tData
	}

	return tSlice, nil
}

func (ts *TrackerService) contextWrapGetSliceSliceByHash(ctx context.Context, method string, args ...string) ([][]any, error) {
	type result struct {
		sliceOfSlices [][]any
		err           error
	}

	// Buffered so the goroutine can always hand off its result and exit, even once we've bailed out on ctx.Done()
	resultChan := make(chan result, 1)

	go func() {
		sliceOfSlices, err := ts.C.getSliceSliceByHash(method, args...)
		resultChan <- result{sliceOfSlices, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case r := <-resultChan:
		return r.sliceOfSlices, r.err
	}
}

// TrackerDataFromSlice builds a tracker's data map by pairing the requested fields with the values rTorrent returned
func TrackerDataFromSlice(fields []TrackerField, data []any) (map[TrackerField]any, error) {
	if len(data) == 0 {
		return nil, ErrNoDataFromTracker
	}
	if len(data) < len(fields) {
		return nil, fmt.Errorf("%w: got %d values for %d requested fields", ErrBadData, len(data), len(fields))
	}
	tData := make(map[TrackerField]any, len(fields))
	// We range the values rather than the fields so that the index is provably in bounds for both slices
	for i, v := range data[:len(fields)] {
		tData[fields[i]] = v
	}
	return tData, nil
}
