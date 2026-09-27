package rtorrent

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestPeerServiceWithDetails(t *testing.T) {
	t.Parallel()

	mockClient := NewMockClient(gomock.NewController(t))
	ps := &PeerService{C: mockClient}

	mockClient.EXPECT().getSliceSliceByHash(peerListMultiCall, testInfoHash, "p.address=").
		Return([][]any{{"192.0.2.1"}}, nil)

	got, err := ps.PeersWithDetails(testInfoHash, []string{"p.address="})
	require.NoError(t, err)
	assert.Equal(t, [][]any{{"192.0.2.1"}}, got)
}
