package rtorrent

import (
	"errors"
	"strconv"
)

const (
	// trackerListMultiCall is used in methods which retrieve a list of trackers along with subsequent commands to call on each
	// See: https://rtorrent-docs.readthedocs.io/en/latest/cmd-ref.html#download-items-and-attributes for more info
	trackerListMultiCall = "t.muticall"
)

var (
	ErrNilTrackerIndex = errors.New("nil tracker index")
)

type TrackerService struct {
	c *Client
}

type TrackerIndex struct {
	infoHash string
	index    int
}

// NewTrackerNoIndex creates a new trackerIndex with no index specification, meaning that all trackers for the given infoHash will be
// executed upon
func NewTrackerNoIndex(infoHash string) *TrackerIndex {
	return &TrackerIndex{infoHash: infoHash, index: -1}
}

// NewTrackerWithIndex creates a new trackerIndex with an index specification, meaning that only the tracker at the given index for the
// given infoHash will be executed upon
func NewTrackerWithIndex(infoHash string, index int) *TrackerIndex {
	return &TrackerIndex{infoHash: infoHash, index: index}
}

func (ti *TrackerIndex) String() string {
	if ti.index == -1 {
		return ti.infoHash
	}
	return ti.infoHash + ":" + strconv.Itoa(ti.index)
}

// TrackerWithDetails retrieves a list of active downloads from rTorrent.
func (ts *TrackerService) TrackerWithDetails(ti *TrackerIndex, commands []string) ([][]any, error) {
	if ti == nil {
		return nil, ErrNilTrackerIndex
	}
	newCmds := append([]string{ti.String()}, commands...)
	return ts.c.getSliceSlice(trackerListMultiCall, newCmds...)
}
