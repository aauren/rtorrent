package rtorrent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	testInfoHash  = strings.Repeat("A", 40)
	testDownloads = []string{strings.Repeat("A", 40), strings.Repeat("B", 40), strings.Repeat("C", 40)}
)

func TestDownloadServiceLists(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		filter string
		call   func(*DownloadService) ([]string, error)
	}{
		{"all", "", (*DownloadService).All},
		{"started", "started", (*DownloadService).Started},
		{"stopped", "stopped", (*DownloadService).Stopped},
		{"complete", "complete", (*DownloadService).Complete},
		{"incomplete", "incomplete", (*DownloadService).Incomplete},
		{"hashing", "hashing", (*DownloadService).Hashing},
		{"seeding", "seeding", (*DownloadService).Seeding},
		{"leeching", "leeching", (*DownloadService).Leeching},
		{"active", "active", (*DownloadService).Active},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Every download_list call leads with an empty string, with the filter appended only when there is one
			wantParams := []string{""}
			if tt.filter != "" {
				wantParams = append(wantParams, tt.filter)
			}

			ds := &DownloadService{C: testClient(t, downloadList, wantParams, testDownloads)}

			got, err := tt.call(ds)
			require.NoError(t, err)
			assert.Equal(t, testDownloads, got)
		})
	}
}

func TestDownloadServiceCountersByInfoHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		call   func(*DownloadService, string) (int, error)
	}{
		{"download rate", "d.down.rate", (*DownloadService).DownloadRate},
		{"download total", "d.down.total", (*DownloadService).DownloadTotal},
		{"upload rate", "d.up.rate", (*DownloadService).UploadRate},
		{"upload total", "d.up.total", (*DownloadService).UploadTotal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ds := &DownloadService{C: testClient(t, tt.method, []string{testInfoHash}, testBytes)}

			got, err := tt.call(ds, testInfoHash)
			require.NoError(t, err)
			assert.Equal(t, testBytes, got)
		})
	}
}

func TestDownloadServiceBaseFilename(t *testing.T) {
	t.Parallel()

	const wantName = "foobar"

	ds := &DownloadService{C: testClient(t, "d.base_filename", []string{testInfoHash}, wantName)}

	got, err := ds.BaseFilename(testInfoHash)
	require.NoError(t, err)
	assert.Equal(t, wantName, got)
}

func TestDownloadServiceWithDetails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockClient := NewMockClient(ctrl)
	ds := &DownloadService{C: mockClient}

	// DownloadWithDetails always prepends "default" to the caller's commands
	mockClient.EXPECT().getSliceSlice(downloadListMultiCall, "default", "d.name=").
		Return([][]any{{"a name"}}, nil)

	got, err := ds.DownloadWithDetails([]string{"d.name="})
	require.NoError(t, err)
	assert.Equal(t, [][]any{{"a name"}}, got)
}
