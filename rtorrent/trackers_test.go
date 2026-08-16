package rtorrent

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const (
	testID  = "test_id"
	testURL = "test_url"
)

func TestTrackerIndex_String(t *testing.T) {
	t.Parallel()

	ti := &TrackerIndex{InfoHash: "12345", Index: 1}
	assert.Equal(t, "12345:1", ti.String())

	ti = &TrackerIndex{InfoHash: "12345", Index: -1}
	assert.Equal(t, "12345", ti.String())
}

func TestTrackerField_AsXMLRPCArgument(t *testing.T) {
	t.Parallel()

	tf := TrackerField("test_field")
	assert.Equal(t, "t.test_field=", tf.AsXMLRPCArgument())
}

func TestTrackerEvent_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		event    TrackerEvent
		expected string
	}{
		{EventNone, "None"},
		{EventCompleted, "Completed"},
		{EventStarted, "Started"},
		{EventStopped, "Stopped"},
		{EventScrape, "Scrape"},
		{TrackerEvent(999), unknownStr}, // Test for an unknown event
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.event.String())
		})
	}
}

func TestTrackerType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		trackerType TrackerType
		expected    string
	}{
		{TypeHTTP, "HTTP"},
		{TypeUDP, "UDP"},
		{TypeDHT, "DHT"},
		{TrackerType(999), unknownStr}, // Test for an unknown type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.trackerType.String())
		})
	}
}

func TestAllTrackerFields(t *testing.T) {
	t.Parallel()

	fields := AllTrackerFields()
	assert.Len(t, fields, len(fieldStringers))
	assert.IsIncreasing(t, fields, "AllTrackerFields should come back sorted")
	assert.Contains(t, fields, FieldURL)

	// Callers get their own copy, so mangling it must not affect what the package considers a valid field
	fields[0] = TrackerField("clobbered")
	assert.NotContains(t, AllTrackerFields(), TrackerField("clobbered"))
}

func TestTracker_CloneWithTrackerIndex(t *testing.T) {
	t.Parallel()

	ti := &TrackerIndex{InfoHash: "12345", Index: 1}
	tracker := &Tracker{ti: ti}
	newTi := &TrackerIndex{InfoHash: "67890", Index: 2}
	clonedTracker := tracker.CloneWithTrackerIndex(newTi)
	assert.Equal(t, newTi, clonedTracker.ti)
	assert.Equal(t, newTi.String(), clonedTracker.ti.String())
	assert.NotEqual(t, tracker.ti, clonedTracker.ti)
}

func TestTracker_GetFieldValueAsString(t *testing.T) {
	t.Parallel()

	tracker := &Tracker{tData: map[TrackerField]any{
		FieldID:          testID,
		FieldLatestEvent: 2,
		FieldMinInterval: "not a number",
	}}

	assert.Equal(t, testID, tracker.GetFieldValueAsString(FieldID))
	assert.Equal(t, "Started", tracker.GetFieldValueAsString(FieldLatestEvent))
	assert.Equal(t, noValueStr, tracker.GetFieldValueAsString(FieldMinInterval), "unreadable value")
	assert.Equal(t, noFieldStr, tracker.GetFieldValueAsString(TrackerField("nope")), "unknown field")
	assert.Equal(t, noValueStr, tracker.GetFieldValueAsString(FieldURL), "known field, absent from this tracker")
}

func TestTracker_StringIsDeterministic(t *testing.T) {
	t.Parallel()

	tracker := &Tracker{
		ti: NewTrackerWithIndex("12345", 0),
		tData: map[TrackerField]any{
			FieldURL:            testURL,
			FieldID:             testID,
			FieldSuccessCounter: 3,
		},
	}

	want := tracker.String()
	// Map ordering is randomized per range, so repeated calls are what shake out any dependence on it
	for range 50 {
		assert.Equal(t, want, tracker.String())
	}

	assert.True(t, strings.HasPrefix(want, "Tracker: TrackerIndex: <12345:0>, data: <"))
	// Fields render in sorted order: id, success_counter, url
	assert.Contains(t, want, "id: test_id, success_counter: 3, url: test_url")
}

func TestNewTrackerNoIndex(t *testing.T) {
	t.Parallel()

	infoHash := "12345"
	ti := NewTrackerNoIndex(infoHash)
	assert.Equal(t, infoHash, ti.InfoHash)
	assert.Equal(t, -1, ti.Index)
}

func TestTrackerService_TrackerWithDetails(t *testing.T) {
	t.Parallel()

	t.Run("test a single tracker return", func(t *testing.T) {
		t.Parallel()

		mockClient := NewMockClient(gomock.NewController(t))
		ts := &TrackerService{C: mockClient}

		ti := &TrackerIndex{InfoHash: "12345", Index: 1}
		fields := []TrackerField{FieldID, FieldURL}
		xmlRPCFields := []string{ti.String(), FieldID.AsXMLRPCArgument(), FieldURL.AsXMLRPCArgument()}

		mockClient.EXPECT().getSliceSliceByHash("t.multicall", xmlRPCFields).Return([][]any{{testID, testURL}}, nil)

		tracker, err := ts.TrackerWithDetails(t.Context(), ti, fields)
		require.NoError(t, err)
		assert.Equal(t, testID, tracker[0].tData[FieldID])
		assert.Equal(t, testURL, tracker[0].tData[FieldURL])
	})

	t.Run("test multiple trackers return", func(t *testing.T) {
		t.Parallel()

		mockClient := NewMockClient(gomock.NewController(t))
		ts := &TrackerService{C: mockClient}

		ti := &TrackerIndex{InfoHash: "12345", Index: -1}
		fields := []TrackerField{FieldID, FieldURL}
		xmlRPCFields := []string{ti.InfoHash, FieldID.AsXMLRPCArgument(), FieldURL.AsXMLRPCArgument()}

		mockClient.EXPECT().getSliceSliceByHash("t.multicall", xmlRPCFields).
			Return([][]any{{testID, testURL}, {"test_id2", "test_url2"}}, nil)

		tracker, err := ts.TrackerWithDetails(t.Context(), ti, fields)
		require.NoError(t, err)
		assert.Equal(t, testID, tracker[0].tData[FieldID])
		assert.Equal(t, testURL, tracker[0].tData[FieldURL])
		assert.Equal(t, 0, tracker[0].ti.Index)
		assert.Equal(t, "test_id2", tracker[1].tData[FieldID])
		assert.Equal(t, "test_url2", tracker[1].tData[FieldURL])
		assert.Equal(t, 1, tracker[1].ti.Index)
	})

	t.Run("nil tracker index is rejected without a request", func(t *testing.T) {
		t.Parallel()

		ts := &TrackerService{C: NewMockClient(gomock.NewController(t))}

		tracker, err := ts.TrackerWithDetails(t.Context(), nil, nil)
		require.ErrorIs(t, err, ErrNilTrackerIndex)
		assert.Nil(t, tracker)
	})

	t.Run("unknown field is rejected without a request", func(t *testing.T) {
		t.Parallel()

		ts := &TrackerService{C: NewMockClient(gomock.NewController(t))}

		ti := NewTrackerNoIndex("12345")
		tracker, err := ts.TrackerWithDetails(t.Context(), ti, []TrackerField{TrackerField("bogus")})
		require.ErrorIs(t, err, ErrUnknownField)
		require.Len(t, tracker, 1)
		assert.Equal(t, ti, tracker[0].TrackerIndex())
	})
}

func TestTrackerService_contextWrapGetSliceSliceByHash(t *testing.T) {
	t.Parallel()

	t.Run("ensure that the general flow of logic is correct for contextWrapGetSliceSliceByHash", func(t *testing.T) {
		t.Parallel()

		mockClient := NewMockClient(gomock.NewController(t))
		ts := &TrackerService{C: mockClient}

		method := "test_method"
		args := []string{"arg1", "arg2"}

		mockClient.EXPECT().getSliceSliceByHash(method, args).Return([][]any{{"result"}}, nil)

		result, err := ts.contextWrapGetSliceSliceByHash(t.Context(), method, args...)
		require.NoError(t, err)
		assert.Equal(t, [][]any{{"result"}}, result)
	})

	t.Run("ensure that the contextWrapGetSliceSliceByHash returns immediately with an error if the context is cancelled", func(t *testing.T) {
		t.Parallel()

		mockClient := NewMockClient(gomock.NewController(t))
		ts := &TrackerService{C: mockClient}

		// Cancelling frees the caller but not the in-flight request, so we expect the call and wait for it rather
		// than letting it land on the mock after the controller has finished
		called := make(chan struct{})
		mockClient.EXPECT().getSliceSliceByHash(gomock.Any(), gomock.Any(), gomock.Any()).
			DoAndReturn(func(string, ...string) ([][]any, error) {
				close(called)
				return nil, nil
			})

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		result, err := ts.contextWrapGetSliceSliceByHash(ctx, "test_method", "arg1", "arg2")
		assert.Nil(t, result)
		require.ErrorIs(t, err, context.Canceled)

		<-called
	})
}

func TestTrackerDataFromSlice(t *testing.T) {
	t.Parallel()

	fields := []TrackerField{FieldID, FieldURL}

	t.Run("pairs fields with values", func(t *testing.T) {
		t.Parallel()

		result, err := TrackerDataFromSlice(fields, []any{testID, testURL})
		require.NoError(t, err)
		assert.Equal(t, testID, result[FieldID])
		assert.Equal(t, testURL, result[FieldURL])
	})

	t.Run("empty data is reported as no data", func(t *testing.T) {
		t.Parallel()

		result, err := TrackerDataFromSlice(fields, nil)
		require.ErrorIs(t, err, ErrNoDataFromTracker)
		assert.Nil(t, result)
	})

	t.Run("short data is reported rather than panicking", func(t *testing.T) {
		t.Parallel()

		result, err := TrackerDataFromSlice(fields, []any{testID})
		require.ErrorIs(t, err, ErrBadData)
		assert.Nil(t, result)
	})
}
