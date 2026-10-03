package rtorrent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSystemServiceStrings(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		call   func(*SystemService) (string, error)
	}{
		{"client version", "system.client_version", (*SystemService).ClientVersion},
		{"library version", "system.library_version", (*SystemService).LibraryVersion},
		{"default directory", "directory.default", (*SystemService).DefaultDirectory},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss := &SystemService{C: testClient(t, tt.method, nil, testName)}

			got, err := tt.call(ss)
			require.NoError(t, err)
			assert.Equal(t, testName, got)
		})
	}
}

func TestSystemServiceLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		call   func(*SystemService) (int, error)
	}{
		{"download limit", "throttle.global_down.max_rate", (*SystemService).GlobalDownloadLimit},
		{"upload limit", "throttle.global_up.max_rate", (*SystemService).GlobalUploadLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ss := &SystemService{C: testClient(t, tt.method, nil, testBytes)}

			got, err := tt.call(ss)
			require.NoError(t, err)
			assert.Equal(t, testBytes, got)
		})
	}
}

func TestSystemServiceSetLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		call   func(*SystemService, int) error
	}{
		{"download limit", "throttle.global_down.max_rate.set", (*SystemService).SetGlobalDownloadLimit},
		{"upload limit", "throttle.global_up.max_rate.set", (*SystemService).SetGlobalUploadLimit},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limits := []struct {
				name        string
				bytesPerSec int
				wantErr     bool
			}{
				{name: "negative", bytesPerSec: -1, wantErr: true},
				{name: "unlimited", bytesPerSec: 0},
				{name: "one byte", bytesPerSec: 1, wantErr: true},
				{name: "half KiB", bytesPerSec: 512, wantErr: true},
				{name: "below minimum", bytesPerSec: 1023, wantErr: true},
				{name: "minimum", bytesPerSec: 1024},
				{name: "fractional KiB", bytesPerSec: 1536},
				{name: "multiple KiB", bytesPerSec: 2048},
			}

			for _, limit := range limits {
				t.Run(limit.name, func(t *testing.T) {
					t.Parallel()

					mockClient := NewMockClient(gomock.NewController(t))
					if !limit.wantErr {
						mockClient.EXPECT().Call(tt.method, "", limit.bytesPerSec).Return(int64(0), nil)
					}

					err := tt.call(&SystemService{C: mockClient}, limit.bytesPerSec)
					if limit.wantErr {
						require.ErrorIs(t, err, ErrBadData)
						return
					}
					require.NoError(t, err)
				})
			}
		})
	}
}
