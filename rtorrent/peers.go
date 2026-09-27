package rtorrent

// peerListMultiCall retrieves a download's connected peers along with subsequent commands to call on each
const peerListMultiCall = "p.multicall"

// A PeerService is a wrapper for Client methods which operate on the peers connected to a download.
type PeerService struct {
	C Client
}

// PeersWithDetails retrieves one row per connected peer of a download, holding the result of each command in order.
// Commands take the d.multicall2 form, e.g. "p.address=".
func (s *PeerService) PeersWithDetails(infoHash string, commands []string) ([][]any, error) {
	return s.C.getSliceSliceByHash(peerListMultiCall, append([]string{infoHash}, commands...)...)
}
