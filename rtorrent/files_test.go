package rtorrent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestFileServiceWithDetails(t *testing.T) {
	t.Parallel()

	mockClient := NewMockClient(gomock.NewController(t))
	fs := &FileService{C: mockClient}

	mockClient.EXPECT().getSliceSliceByHash(fileListMultiCall, testInfoHash, "f.path=", "f.size_bytes=").
		Return([][]any{{"a.mkv", int64(10)}, {"b.nfo", int64(1)}}, nil)

	got, err := fs.FilesWithDetails(testInfoHash, []string{"f.path=", "f.size_bytes="})
	require.NoError(t, err)
	assert.Equal(t, [][]any{{"a.mkv", int64(10)}, {"b.nfo", int64(1)}}, got)
}
