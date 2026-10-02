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

func TestFileServiceSetPriority(t *testing.T) {
	t.Parallel()

	mockClient := NewMockClient(gomock.NewController(t))
	fs := &FileService{C: mockClient}

	// A file is addressed by its download's hash and its f.multicall index
	mockClient.EXPECT().Call("f.priority.set", testInfoHash+":f3", int(FilePriorityHigh)).Return(int64(0), nil)

	require.NoError(t, fs.SetPriority(testInfoHash, 3, FilePriorityHigh))
	require.ErrorIs(t, fs.SetPriority("", 3, FilePriorityHigh), ErrBadData)
	require.ErrorIs(t, fs.SetPriority(testInfoHash, -1, FilePriorityHigh), ErrBadData)
}
