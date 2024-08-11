package rtorrent

const (
	// downloadList is used in methods which retrieve a list of downloads.
	downloadList = "download_list"

	// downloadListMultiCall is used in methods which retrieve a list of downloads along with subsequent commands to call on each
	// See: https://rtorrent-docs.readthedocs.io/en/latest/cmd-ref.html#download-items-and-attributes for more info
	downloadListMultiCall = "d.multicall2"
)

// A DownloadService is a wrapper for Client methods which operate on downloads.
type DownloadService struct {
	c *Client
}

// All retrieves a list of all downloads from rTorrent.
func (s *DownloadService) All() ([]string, error) {
	return s.c.getStringSlice(downloadList)
}

// Started retrieves a list of started downloads from rTorrent.
func (s *DownloadService) Started() ([]string, error) {
	return s.c.getStringSlice(downloadList, "started")
}

// Stopped retrieves a list of stopped downloads from rTorrent.
func (s *DownloadService) Stopped() ([]string, error) {
	return s.c.getStringSlice(downloadList, "stopped")
}

// Complete retrieves a list of complete downloads from rTorrent.
func (s *DownloadService) Complete() ([]string, error) {
	return s.c.getStringSlice(downloadList, "complete")
}

// Incomplete retrieves a list of incomplete downloads from rTorrent.
func (s *DownloadService) Incomplete() ([]string, error) {
	return s.c.getStringSlice(downloadList, "incomplete")
}

// Hashing retrieves a list of hashing downloads from rTorrent.
func (s *DownloadService) Hashing() ([]string, error) {
	return s.c.getStringSlice(downloadList, "hashing")
}

// Seeding retrieves a list of seeding downloads from rTorrent.
func (s *DownloadService) Seeding() ([]string, error) {
	return s.c.getStringSlice(downloadList, "seeding")
}

// Leeching retrieves a list of leeching downloads from rTorrent.
func (s *DownloadService) Leeching() ([]string, error) {
	return s.c.getStringSlice(downloadList, "leeching")
}

// Active retrieves a list of active downloads from rTorrent.
func (s *DownloadService) Active() ([]string, error) {
	return s.c.getStringSlice(downloadList, "active")
}

// DownloadWithDetails retrieves a list of downloads from rTorrent along with additional details as specified by the commands slice.
func (s *DownloadService) DownloadWithDetails(commands []string) ([][]any, error) {
	newCmds := append([]string{"active"}, commands...)
	return s.c.getSliceSlice(downloadListMultiCall, newCmds...)
}

// BaseFilename retrieves the base filename shown in the rTorrent UI for a specific download, by its info-hash.
func (s *DownloadService) BaseFilename(infoHash string) (string, error) {
	return s.c.getString("d.base_filename", infoHash)
}

// DownloadRate retrieves the current download rate in bytes for a specific download, by its info-hash.
func (s *DownloadService) DownloadRate(infoHash string) (int, error) {
	return s.c.getInt("d.down.rate", infoHash)
}

// DownloadTotal retrieves the total bytes downloaded for a specific download, by its info-hash.
func (s *DownloadService) DownloadTotal(infoHash string) (int, error) {
	return s.c.getInt("d.down.total", infoHash)
}

// UploadRate retrieves the current upload rate in bytes for a specific download, by its info-hash.
func (s *DownloadService) UploadRate(infoHash string) (int, error) {
	return s.c.getInt("d.up.rate", infoHash)
}

// UploadTotal retrieves the total bytes uploaded for a specific download, by its info-hash.
func (s *DownloadService) UploadTotal(infoHash string) (int, error) {
	return s.c.getInt("d.up.total", infoHash)
}
