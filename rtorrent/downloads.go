package rtorrent

import "slices"

const (
	// downloadList is used in methods which retrieve a list of downloads.
	downloadList = "download_list"

	// downloadListMultiCall is used in methods which retrieve a list of downloads along with subsequent commands to call on each
	// See: https://rtorrent-docs.readthedocs.io/en/latest/cmd-ref.html#download-items-and-attributes for more info
	downloadListMultiCall = "d.multicall2"
)

// A DownloadService is a wrapper for Client methods which operate on downloads.
type DownloadService struct {
	C Client
}

// All retrieves a list of all downloads from rTorrent.
func (s *DownloadService) All() ([]string, error) {
	return s.C.getStringSlice(downloadList)
}

// Started retrieves a list of started downloads from rTorrent.
func (s *DownloadService) Started() ([]string, error) {
	return s.C.getStringSlice(downloadList, "started")
}

// Stopped retrieves a list of stopped downloads from rTorrent.
func (s *DownloadService) Stopped() ([]string, error) {
	return s.C.getStringSlice(downloadList, "stopped")
}

// Complete retrieves a list of complete downloads from rTorrent.
func (s *DownloadService) Complete() ([]string, error) {
	return s.C.getStringSlice(downloadList, "complete")
}

// Incomplete retrieves a list of incomplete downloads from rTorrent.
func (s *DownloadService) Incomplete() ([]string, error) {
	return s.C.getStringSlice(downloadList, "incomplete")
}

// Hashing retrieves a list of hashing downloads from rTorrent.
func (s *DownloadService) Hashing() ([]string, error) {
	return s.C.getStringSlice(downloadList, "hashing")
}

// Seeding retrieves a list of seeding downloads from rTorrent.
func (s *DownloadService) Seeding() ([]string, error) {
	return s.C.getStringSlice(downloadList, "seeding")
}

// Leeching retrieves a list of leeching downloads from rTorrent.
func (s *DownloadService) Leeching() ([]string, error) {
	return s.C.getStringSlice(downloadList, "leeching")
}

// Active retrieves a list of active downloads from rTorrent.
func (s *DownloadService) Active() ([]string, error) {
	return s.C.getStringSlice(downloadList, "active")
}

// DownloadWithDetails retrieves a list of downloads from rTorrent along with additional details as specified by the commands slice.
func (s *DownloadService) DownloadWithDetails(commands []string) ([][]any, error) {
	return s.C.getSliceSlice(downloadListMultiCall, slices.Concat([]string{"default"}, commands)...)
}

// DownloadDetails retrieves the given commands for one download in a single round trip, returning one value per
// command in order. Commands take the same d.multicall2 form DownloadWithDetails expects, so "d.name=" and
// "d.custom=label" both work, as does a bare "d.name".
func (s *DownloadService) DownloadDetails(infoHash string, commands []string) ([]any, error) {
	return s.C.multicallByHash(infoHash, commands...)
}

// Start starts a download, by its info-hash.
func (s *DownloadService) Start(infoHash string) error {
	return s.C.commandByHash("d.start", infoHash)
}

// Stop stops a download without closing its files, by its info-hash.
func (s *DownloadService) Stop(infoHash string) error {
	return s.C.commandByHash("d.stop", infoHash)
}

// Open opens a download's files, by its info-hash.
func (s *DownloadService) Open(infoHash string) error {
	return s.C.commandByHash("d.open", infoHash)
}

// Close closes a download's files, by its info-hash. rTorrent refuses to move a download's directory while it's open.
func (s *DownloadService) Close(infoHash string) error {
	return s.C.commandByHash("d.close", infoHash)
}

// Erase removes a download from rTorrent, by its info-hash. It never touches the data on disk.
func (s *DownloadService) Erase(infoHash string) error {
	return s.C.commandByHash("d.erase", infoHash)
}

// CheckHash queues a hash check of a download's data, by its info-hash. It returns before the check finishes.
func (s *DownloadService) CheckHash(infoHash string) error {
	return s.C.commandByHash("d.check_hash", infoHash)
}

// SetDirectory points a download at a new directory, by its info-hash. For a multi-file download rTorrent appends the
// download's name, so dir is the parent of the download's own folder.
func (s *DownloadService) SetDirectory(infoHash, dir string) error {
	return s.C.commandByHash("d.directory.set", infoHash, dir)
}

// SetMessage replaces a download's message, by its info-hash. An empty msg clears it.
func (s *DownloadService) SetMessage(infoHash, msg string) error {
	return s.C.commandByHash("d.message.set", infoHash, msg)
}

// SetCustom1 sets a download's custom1 field, by its info-hash, which ruTorrent and most tools use as the label.
func (s *DownloadService) SetCustom1(infoHash, value string) error {
	return s.C.commandByHash("d.custom1.set", infoHash, value)
}

// BaseFilename retrieves the base filename shown in the rTorrent UI for a specific download, by its info-hash.
func (s *DownloadService) BaseFilename(infoHash string) (string, error) {
	return s.C.getString("d.base_filename", infoHash)
}

// DownloadRate retrieves the current download rate in bytes for a specific download, by its info-hash.
func (s *DownloadService) DownloadRate(infoHash string) (int, error) {
	return s.C.getInt("d.down.rate", infoHash)
}

// DownloadTotal retrieves the total bytes downloaded for a specific download, by its info-hash.
func (s *DownloadService) DownloadTotal(infoHash string) (int, error) {
	return s.C.getInt("d.down.total", infoHash)
}

// UploadRate retrieves the current upload rate in bytes for a specific download, by its info-hash.
func (s *DownloadService) UploadRate(infoHash string) (int, error) {
	return s.C.getInt("d.up.rate", infoHash)
}

// UploadTotal retrieves the total bytes uploaded for a specific download, by its info-hash.
func (s *DownloadService) UploadTotal(infoHash string) (int, error) {
	return s.C.getInt("d.up.total", infoHash)
}
