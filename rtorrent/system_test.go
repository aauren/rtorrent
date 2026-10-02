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

			mockClient := NewMockClient(gomock.NewController(t))
			mockClient.EXPECT().Call(tt.method, "", testBytes).Return(int64(0), nil)

			require.NoError(t, tt.call(&SystemService{C: mockClient}, testBytes))
		})
	}
}
