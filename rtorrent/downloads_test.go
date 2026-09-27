package rtorrent

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

const testName = "a name"

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

func TestDownloadServiceCommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		wantParams []string
		call       func(*DownloadService) error
	}{
		{"start", "d.start", []string{testInfoHash}, func(s *DownloadService) error { return s.Start(testInfoHash) }},
		{"stop", "d.stop", []string{testInfoHash}, func(s *DownloadService) error { return s.Stop(testInfoHash) }},
		{"open", "d.open", []string{testInfoHash}, func(s *DownloadService) error { return s.Open(testInfoHash) }},
		{"close", "d.close", []string{testInfoHash}, func(s *DownloadService) error { return s.Close(testInfoHash) }},
		{"erase", "d.erase", []string{testInfoHash}, func(s *DownloadService) error { return s.Erase(testInfoHash) }},
		{"check hash", "d.check_hash", []string{testInfoHash}, func(s *DownloadService) error { return s.CheckHash(testInfoHash) }},
		{
			"set directory", "d.directory.set", []string{testInfoHash, "/data/label"},
			func(s *DownloadService) error { return s.SetDirectory(testInfoHash, "/data/label") },
		},
		{
			"set message", "d.message.set", []string{testInfoHash, "hi"},
			func(s *DownloadService) error { return s.SetMessage(testInfoHash, "hi") },
		},
		{
			"set custom1", "d.custom1.set", []string{testInfoHash, "tv"},
			func(s *DownloadService) error { return s.SetCustom1(testInfoHash, "tv") },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ds := &DownloadService{C: testClient(t, tt.method, tt.wantParams, 0)}
			require.NoError(t, tt.call(ds))
		})
	}
}

func TestDownloadServiceCommandsRequireInfoHash(t *testing.T) {
	t.Parallel()

	ds := &DownloadService{C: &XMLRPCClient{}}
	require.ErrorIs(t, ds.Erase(""), ErrBadData)
	require.ErrorIs(t, ds.SetDirectory("", "/tmp"), ErrBadData)
}

func TestDownloadServiceDetails(t *testing.T) {
	t.Parallel()

	mockClient := NewMockClient(gomock.NewController(t))
	ds := &DownloadService{C: mockClient}

	mockClient.EXPECT().multicallByHash(testInfoHash, "d.name", "d.complete=").Return([]any{testName, int64(1)}, nil)

	got, err := ds.DownloadDetails(testInfoHash, []string{"d.name", "d.complete="})
	require.NoError(t, err)
	assert.Equal(t, []any{testName, int64(1)}, got)
}

func TestDownloadServiceWithDetails(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	mockClient := NewMockClient(ctrl)
	ds := &DownloadService{C: mockClient}

	// DownloadWithDetails always prepends "default" to the caller's commands
	mockClient.EXPECT().getSliceSlice(downloadListMultiCall, "default", "d.name=").
		Return([][]any{{testName}}, nil)

	got, err := ds.DownloadWithDetails([]string{"d.name="})
	require.NoError(t, err)
	assert.Equal(t, [][]any{{testName}}, got)
}
