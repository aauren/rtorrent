package rtorrent

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// trackerListMultiCall is used in methods which retrieve a list of trackers along with subsequent commands to call on each
	// See: https://rtorrent-docs.readthedocs.io/en/latest/cmd-ref.html#download-items-and-attributes for more info
	trackerListMultiCall = "t.multicall"
)

var (
	// Custom error definitions
	ErrNilTrackerIndex   = errors.New("nil tracker index")
	ErrNoField           = errors.New("no field found")
	ErrUnknownField      = errors.New("unknown field")
	ErrBadData           = errors.New("bad data")
	ErrNoDataFromTracker = errors.New("no data from tracker")
	ErrMultipleTrackers  = errors.New("multiple trackers returned")

	// List of all possible fields that can be retrieved from a tracker
	AllTrackerFields = []TrackerField{FieldCanScrape, FieldIsUsable, FieldIsEnabled, FieldFailedCounter, FieldActivityLast,
		FieldActivityNext, FieldFailedLast, FieldFailedNext, FieldID, FieldIsBusy, FieldIsOpen, FiledIsExtraTracker, FieldLatestEvent,
		FieldMinInterval, FieldNormalInterval, FieldSuccessCounter, FieldSuccessLast, FieldSuccessNext, FieldType, FieldURL}
)

const (
	// XMLRPC Tracker Fields
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

	// Tracker Events
	EventNone      = 0
	EventCompleted = 1
	EventStarted   = 2
	EventStopped   = 3
	EventScrape    = 4

	// Tracker Types
	TypeHTTP = 1
	TypeUDP  = 2
	TypeDHT  = 3
)

// TrackerService is used to interact with the tracker information gatherer methods in rTorrent
type TrackerService struct {
	C *Client
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
		return "Unknown"
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
		return "Unknown"
	}
}

// Tracker is used to represent information about a tracker in rTorrent
type Tracker struct {
	ti    *TrackerIndex
	tData map[TrackerField]interface{}
}

func (t *Tracker) CloneWithTrackerIndex(ti *TrackerIndex) *Tracker {
	return &Tracker{ti: ti, tData: t.tData}
}

func (t *Tracker) GetFieldValueAsString(f TrackerField) string {
	var tmpBool bool
	var tmpInt int
	var tmpTime time.Time
	var tmpEvent TrackerEvent
	var tmpType TrackerType
	var str string
	var err error
	switch f {
	case FieldCanScrape:
		tmpBool, err = t.CanScrape()
		str = strconv.FormatBool(tmpBool)
	case FieldIsUsable:
		tmpBool, err = t.IsUsable()
		str = strconv.FormatBool(tmpBool)
	case FieldIsEnabled:
		tmpBool, err = t.IsEnabled()
		str = strconv.FormatBool(tmpBool)
	case FieldFailedCounter:
		tmpInt, err = t.FailedCounter()
		str = strconv.Itoa(tmpInt)
	case FieldActivityLast:
		tmpTime, err = t.ActivityLastTime()
		str = tmpTime.String()
	case FieldActivityNext:
		tmpTime, err = t.ActivityTimeNext()
		str = tmpTime.String()
	case FieldFailedLast:
		tmpTime, err = t.FailedTimeLast()
		str = tmpTime.String()
	case FieldFailedNext:
		tmpTime, err = t.FailedTimeNext()
		str = tmpTime.String()
	case FieldID:
		str, err = t.ID()
	case FieldIsBusy:
		tmpBool, err = t.IsBusy()
		str = strconv.FormatBool(tmpBool)
	case FieldIsOpen:
		tmpBool, err = t.IsOpen()
		str = strconv.FormatBool(tmpBool)
	case FiledIsExtraTracker:
		tmpBool, err = t.IsExtraTracker()
		str = strconv.FormatBool(tmpBool)
	case FieldLatestEvent:
		tmpEvent, err = t.LatestEvent()
		str = tmpEvent.String()
	case FieldMinInterval:
		tmpInt, err = t.MinInterval()
		str = strconv.Itoa(tmpInt)
	case FieldNormalInterval:
		tmpInt, err = t.NormalInterval()
		str = strconv.Itoa(tmpInt)
	case FieldSuccessCounter:
		tmpInt, err = t.SuccessCounter()
		str = strconv.Itoa(tmpInt)
	case FieldSuccessLast:
		tmpTime, err = t.SuccessTimeLast()
		str = tmpTime.String()
	case FieldSuccessNext:
		tmpTime, err = t.SuccessTimeNext()
		str = tmpTime.String()
	case FieldType:
		tmpType, err = t.Type()
		str = tmpType.String()
	case FieldURL:
		str, err = t.URL()
	default:
		return "<ne>"
	}

	if err != nil {
		return "<na>"
	}
	return str
}

func (t *Tracker) String() string {
	var sb strings.Builder
	for k := range t.tData {
		sb.WriteString(k.String())
		sb.WriteString(": ")
		sb.WriteString(t.GetFieldValueAsString(k))
		sb.WriteString(", ")
	}
	s := sb.String()
	if len(s) > 2 {
		s = s[:len(s)-2]
	}

	return fmt.Sprintf("Tracker: TrackerIndex: <%s>, data: <%s>", t.ti, s)
}

// TrackerIndex returns the TrackerIndex for the tracker
func (t *Tracker) TrackerIndex() *TrackerIndex {
	return t.ti
}

// CanScrape Checks if the announce URL is scrapeable. rTorrent considers a HTTP tracker scrapeable if the announce URL contains the string
// /announce somewhere after the rightmost / (inclusively).
func (t *Tracker) CanScrape() (bool, error) {
	data, ok := t.tData[FieldCanScrape]
	if !ok {
		return false, ErrNoField
	}
	return boolFromAny(data)
}

// IsUsable Checks if the tracker is usable. A tracker is considered usable if it is enabled and not marked as failed.
func (t *Tracker) IsUsable() (bool, error) {
	data, ok := t.tData[FieldIsUsable]
	if !ok {
		return false, ErrNoField
	}
	return boolFromAny(data)
}

// IsEnabled Checks if the tracker is enabled. A tracker is considered enabled if it is not marked as disabled.
func (t *Tracker) IsEnabled() (bool, error) {
	data, ok := t.tData[FieldIsEnabled]
	if !ok {
		return false, ErrNoField
	}
	return boolFromAny(data)
}

// FailedCounter Returns the number of failed requests to the tracker. Note that this value resets to 0 if a request succeeds.
func (t *Tracker) FailedCounter() (int, error) {
	data, ok := t.tData[FieldFailedCounter]
	if !ok {
		return 0, ErrNoField
	}
	return intFromAny(data)
}

// ActivityTimeLast Returns the last time there was an attempt to announce to this tracker, regardless of whether or not the announce
// succeeded.
func (t *Tracker) ActivityLastTime() (time.Time, error) {
	data, ok := t.tData[FieldActivityLast]
	if !ok {
		return time.Time{}, ErrNoField
	}
	return timeFromAny(data)
}

// ActivityTimeNext Returns when rtorrent will attempt to announce to the tracker next. In most cases, t.activity_time_next -
// t.activity_time_last will equal t.normal_interval.
func (t *Tracker) ActivityTimeNext() (time.Time, error) {
	data, ok := t.tData[FieldActivityNext]
	if !ok {
		return time.Time{}, ErrNoField
	}
	return timeFromAny(data)
}

// FailedTimeLast Returns the last time there was a failed attempt to announce to this tracker.
func (t *Tracker) FailedTimeLast() (time.Time, error) {
	data, ok := t.tData[FieldFailedLast]
	if !ok {
		return time.Time{}, ErrNoField
	}
	return timeFromAny(data)
}

// FailedTimeNext Returns the time at when the next request is planned to happen after a failed request. rTorrent backs off failed requests
// exponentially, i.e. each time a request fails, it doubles the interval until it tries again.
func (t *Tracker) FailedTimeNext() (time.Time, error) {
	data, ok := t.tData[FieldFailedNext]
	if !ok {
		return time.Time{}, ErrNoField
	}
	return timeFromAny(data)
}

// ID If a previous HTTP tracker response contains the tracker id key, t.id will contain that value, and it will be added as a parameter to
// any subsequent requests to that same tracker.
func (t *Tracker) ID() (string, error) {
	data, ok := t.tData[FieldID]
	if !ok {
		return "", ErrNoField
	}
	return stringFromAny(data)
}

// IsBusy Returns true if the request is in the middle of processing, and false otherwise (this is identical to IsOpen())
func (t *Tracker) IsBusy() (bool, error) {
	data, ok := t.tData[FieldIsBusy]
	if !ok {
		return false, ErrNoField
	}
	return boolFromAny(data)
}

// IsOpen Returns true if the request is in the middle of processing, and false otherwise (this is identical to IsBusy())
func (t *Tracker) IsOpen() (bool, error) {
	data, ok := t.tData[FieldIsOpen]
	if !ok {
		return false, ErrNoField
	}
	return boolFromAny(data)
}

// IsExtraTracker Returns true if the tracker was added via d.tracker.insert, rather than existing in the original metafile.
func (t *Tracker) IsExtraTracker() (bool, error) {
	data, ok := t.tData[FiledIsExtraTracker]
	if !ok {
		return false, ErrNoField
	}
	return boolFromAny(data)
}

// LatestEvent Returns the latest event that occurred with the tracker. The possible values are:
// 0: None
// 1: Completed
// 2: Started
// 3: Stopped
// 4: Scrape (this isn't an actual event key the BitTorrent spec defines, instead this indicates that the tracker is currently processing
//
//	a scrape request)
func (t *Tracker) LatestEvent() (TrackerEvent, error) {
	data, ok := t.tData[FieldLatestEvent]
	if !ok {
		return 0, ErrNoField
	}
	eventInt, err := intFromAny(data)
	if err != nil {
		return 0, err
	}
	return TrackerEvent(eventInt), nil
}

// MinInterval Returns the values for the minimum announce intervals as returned from the tracker request.
func (t *Tracker) MinInterval() (int, error) {
	data, ok := t.tData[FieldMinInterval]
	if !ok {
		return 0, ErrNoField
	}
	return intFromAny(data)
}

// NormalInterval Returns the values for the normal announce intervals as returned from the tracker request.
func (t *Tracker) NormalInterval() (int, error) {
	data, ok := t.tData[FieldNormalInterval]
	if !ok {
		return 0, ErrNoField
	}
	return intFromAny(data)
}

// SuccessCounter Returns the number of successful requests to the tracker.
func (t *Tracker) SuccessCounter() (int, error) {
	data, ok := t.tData[FieldSuccessCounter]
	if !ok {
		return 0, ErrNoField
	}
	return intFromAny(data)
}

// SuccessTimeLast Returns the last time there was a successful attempt to announce to this tracker.
func (t *Tracker) SuccessTimeLast() (time.Time, error) {
	data, ok := t.tData[FieldSuccessLast]
	if !ok {
		return time.Time{}, ErrNoField
	}
	return timeFromAny(data)
}

// SuccessTimeNext Returns the time at when the next request is planned to happen after a successful request.
func (t *Tracker) SuccessTimeNext() (time.Time, error) {
	data, ok := t.tData[FieldSuccessNext]
	if !ok {
		return time.Time{}, ErrNoField
	}
	return timeFromAny(data)
}

// Type Returns the type of the tracker. The possible values are:
// 1: HTTP
// 2: UDP
// 3: DHT
func (t *Tracker) Type() (TrackerType, error) {
	data, ok := t.tData[FieldType]
	if !ok {
		return 0, ErrNoField
	}
	typeInt, err := intFromAny(data)
	if err != nil {
		return 0, err
	}
	return TrackerType(typeInt), nil
}

func (t *Tracker) URL() (string, error) {
	data, ok := t.tData[FieldURL]
	if !ok {
		return "", ErrNoField
	}
	return stringFromAny(data)
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

// TrackerWithDetails retrieves a list of active downloads from rTorrent.
//
// If the TrackerIndexed passed is nil, then ErrNilTrackerIndex will be returned. This error is special and will result in a nil Tracker
//
// All other errors will are the result of a request to rtorrent and will Tracker will be returned with any data fields that were able to
// be collected along with an intact version of the TrackerIndex set inside the Tracker.
func (ts *TrackerService) TrackerWithDetails(ctx context.Context, ti *TrackerIndex, fields []TrackerField) ([]*Tracker, error) {
	if ti == nil {
		return nil, ErrNilTrackerIndex
	}
	t := Tracker{ti: ti, tData: make(map[TrackerField]interface{})}
	tSlice := []*Tracker{&t}
	newCmds := []string{ti.String()}
	for _, field := range fields {
		if !slices.Contains(AllTrackerFields, field) {
			return tSlice, ErrUnknownField
		}
		newCmds = append(newCmds, field.AsXMLRPCArgument())
	}
	sliceOfSlices, err := ts.contextWrapGetSliceSliceByHash(ctx, trackerListMultiCall, newCmds...)
	if err != nil {
		return tSlice, err
	}

	tSlice = make([]*Tracker, len(sliceOfSlices))
	for i, slice := range sliceOfSlices {
		if len(sliceOfSlices) > 1 {
			tSlice[i] = &Tracker{ti: NewTrackerWithIndex(ti.InfoHash, i), tData: make(map[TrackerField]interface{})}
		} else {
			tSlice[i] = &Tracker{ti: ti, tData: make(map[TrackerField]interface{})}
		}
		tSlice[i].tData, err = TrackerDataFromSlice(fields, slice)
		if err != nil {
			return tSlice, err
		}
	}

	return tSlice, nil
}

func (ts *TrackerService) contextWrapGetSliceSliceByHash(ctx context.Context, method string, args ...string) ([][]any, error) {
	// Create a channel to receive the result
	resultChan := make(chan struct {
		sliceOfSlices [][]any
		err           error
	}, 1)

	// Run the request in a separate goroutine
	go func() {
		sliceOfSlices, err := ts.C.getSliceSliceByHash(method, args...)
		resultChan <- struct {
			sliceOfSlices [][]any
			err           error
		}{sliceOfSlices, err}
	}()

	select {
	case <-ctx.Done():
		// Context was cancelled or timed out
		return nil, ctx.Err()
	case result := <-resultChan:
		// Request completed
		return result.sliceOfSlices, result.err
	}
}

// TrackerFromSlice creates a new Tracker from a slice of data
func TrackerDataFromSlice(fields []TrackerField, data []any) (map[TrackerField]interface{}, error) {
	if len(data) == 0 {
		return nil, ErrNoDataFromTracker
	}
	tData := make(map[TrackerField]interface{})
	for i, field := range fields {
		tData[field] = data[i]
	}
	return tData, nil
}
