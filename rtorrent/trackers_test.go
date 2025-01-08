package rtorrent

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	gomock "go.uber.org/mock/gomock"
)

func TestTrackerIndex_String(t *testing.T) {
	ti := &TrackerIndex{InfoHash: "12345", Index: 1}
	assert.Equal(t, "12345:1", ti.String())

	ti = &TrackerIndex{InfoHash: "12345", Index: -1}
	assert.Equal(t, "12345", ti.String())
}

func TestTrackerField_AsXMLRPCArgument(t *testing.T) {
	tf := TrackerField("test_field")
	assert.Equal(t, "t.test_field=", tf.AsXMLRPCArgument())
}

func TestTrackerEvent_String(t *testing.T) {
	tests := []struct {
		event    TrackerEvent
		expected string
	}{
		{EventNone, "None"},
		{EventCompleted, "Completed"},
		{EventStarted, "Started"},
		{EventStopped, "Stopped"},
		{EventScrape, "Scrape"},
		{TrackerEvent(999), "Unknown"}, // Test for an unknown event
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.event.String())
		})
	}
}

func TestTrackerType_String(t *testing.T) {
	tests := []struct {
		trackerType TrackerType
		expected    string
	}{
		{TypeHTTP, "HTTP"},
		{TypeUDP, "UDP"},
		{TypeDHT, "DHT"},
		{TrackerType(999), "Unknown"}, // Test for an unknown type
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.trackerType.String())
		})
	}
}

func TestTracker_CloneWithTrackerIndex(t *testing.T) {
	ti := &TrackerIndex{InfoHash: "12345", Index: 1}
	tracker := &Tracker{ti: ti}
	newTi := &TrackerIndex{InfoHash: "67890", Index: 2}
	clonedTracker := tracker.CloneWithTrackerIndex(newTi)
	assert.Equal(t, newTi, clonedTracker.ti)
	assert.Equal(t, newTi.String(), clonedTracker.ti.String())
	assert.NotEqual(t, tracker.ti, clonedTracker.ti)
}

func TestTracker_GetFieldValueAsString(t *testing.T) {
	tracker := &Tracker{tData: map[TrackerField]interface{}{
		FieldID: "test_id",
	}}
	assert.Equal(t, "test_id", tracker.GetFieldValueAsString(FieldID))
}

func TestNewTrackerNoIndex(t *testing.T) {
	infoHash := "12345"
	ti := NewTrackerNoIndex(infoHash)
	assert.Equal(t, infoHash, ti.InfoHash)
	assert.Equal(t, -1, ti.Index)
}

func TestTrackerService_TrackerWithDetails(t *testing.T) {
	t.Run("test a single tracker return", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockClient(ctrl)
		ts := &TrackerService{C: mockClient}

		ti := &TrackerIndex{InfoHash: "12345", Index: 1}
		fields := []TrackerField{FieldID, FieldURL}
		xmlRPCFields := []string{ti.String(), FieldID.AsXMLRPCArgument(), FieldURL.AsXMLRPCArgument()}

		mockClient.EXPECT().getSliceSliceByHash("t.multicall", xmlRPCFields).Return([][]interface{}{{"test_id", "test_url"}}, nil)

		tracker, err := ts.TrackerWithDetails(context.Background(), ti, fields)
		assert.NoError(t, err)
		assert.Equal(t, "test_id", tracker[0].tData[FieldID])
		assert.Equal(t, "test_url", tracker[0].tData[FieldURL])
	})

	t.Run("test multiple trackers return", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockClient(ctrl)
		ts := &TrackerService{C: mockClient}

		ti := &TrackerIndex{InfoHash: "12345", Index: -1}
		fields := []TrackerField{FieldID, FieldURL}
		xmlRPCFields := []string{ti.InfoHash, FieldID.AsXMLRPCArgument(), FieldURL.AsXMLRPCArgument()}

		mockClient.EXPECT().getSliceSliceByHash("t.multicall", xmlRPCFields).Return([][]interface{}{{"test_id", "test_url"}, {"test_id2", "test_url2"}}, nil)

		tracker, err := ts.TrackerWithDetails(context.Background(), ti, fields)
		assert.NoError(t, err)
		assert.Equal(t, "test_id", tracker[0].tData[FieldID])
		assert.Equal(t, "test_url", tracker[0].tData[FieldURL])
		assert.Equal(t, 0, tracker[0].ti.Index)
		assert.Equal(t, "test_id2", tracker[1].tData[FieldID])
		assert.Equal(t, "test_url2", tracker[1].tData[FieldURL])
		assert.Equal(t, 1, tracker[1].ti.Index)
	})
}

func TestTrackerService_contextWrapGetSliceSliceByHash(t *testing.T) {
	t.Run("ensure that the general flow of logic is correct for contextWrapGetSliceSliceByHash", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockClient(ctrl)
		ts := &TrackerService{C: mockClient}

		method := "test_method"
		args := []string{"arg1", "arg2"}

		mockClient.EXPECT().getSliceSliceByHash(method, args).Return([][]interface{}{{"result"}}, nil)

		result, err := ts.contextWrapGetSliceSliceByHash(context.Background(), method, args...)
		assert.NoError(t, err)
		assert.Equal(t, [][]interface{}{{"result"}}, result)
	})

	t.Run("ensure that the contextWrapGetSliceSliceByHash returns immediately with an error if the context is cancelled", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockClient := NewMockClient(ctrl)
		ts := &TrackerService{C: mockClient}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := ts.contextWrapGetSliceSliceByHash(ctx, "test_method", "arg1", "arg2")
		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestTrackerDataFromSlice(t *testing.T) {
	fields := []TrackerField{FieldID, FieldURL}
	data := []interface{}{"test_id", "test_url"}

	result, err := TrackerDataFromSlice(fields, data)
	assert.NoError(t, err)
	assert.Equal(t, "test_id", result[FieldID])
	assert.Equal(t, "test_url", result[FieldURL])
}
