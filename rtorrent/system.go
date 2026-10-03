package rtorrent

import "fmt"

// A SystemService is a wrapper for Client methods which operate on rTorrent as a whole rather than on one download.
type SystemService struct {
	C Client
}

// ClientVersion retrieves rTorrent's own version, like "0.9.8".
func (s *SystemService) ClientVersion() (string, error) {
	return s.C.getString("system.client_version", "")
}

// LibraryVersion retrieves the version of libtorrent that rTorrent is built against.
func (s *SystemService) LibraryVersion() (string, error) {
	return s.C.getString("system.library_version", "")
}

// DefaultDirectory retrieves the directory new downloads go into unless they're told otherwise.
func (s *SystemService) DefaultDirectory() (string, error) {
	return s.C.getString("directory.default", "")
}

// GlobalDownloadLimit retrieves the global download cap in bytes per second, where 0 means unlimited.
func (s *SystemService) GlobalDownloadLimit() (int, error) {
	return s.C.getInt("throttle.global_down.max_rate", "")
}

// GlobalUploadLimit retrieves the global upload cap in bytes per second, where 0 means unlimited.
func (s *SystemService) GlobalUploadLimit() (int, error) {
	return s.C.getInt("throttle.global_up.max_rate", "")
}

// SetGlobalDownloadLimit caps every download at bytesPerSec combined, where 0 removes the cap.
// rTorrent rounds down to whole KiB/s, so 1536 bytes/s becomes 1024 bytes/s. We reject nonzero
// values below 1024 with ErrBadData because a positive sub-KiB/s limit would become unlimited.
func (s *SystemService) SetGlobalDownloadLimit(bytesPerSec int) error {
	if bytesPerSec != 0 && bytesPerSec < 1024 {
		return fmt.Errorf("%w: global download limit must be 0 (unlimited) or at least 1024 bytes/s", ErrBadData)
	}
	_, err := s.C.Call("throttle.global_down.max_rate.set", "", bytesPerSec)
	return err
}

// SetGlobalUploadLimit caps every upload at bytesPerSec combined, where 0 removes the cap.
// rTorrent rounds down to whole KiB/s, so 1536 bytes/s becomes 1024 bytes/s. We reject nonzero
// values below 1024 with ErrBadData because a positive sub-KiB/s limit would become unlimited.
func (s *SystemService) SetGlobalUploadLimit(bytesPerSec int) error {
	if bytesPerSec != 0 && bytesPerSec < 1024 {
		return fmt.Errorf("%w: global upload limit must be 0 (unlimited) or at least 1024 bytes/s", ErrBadData)
	}
	_, err := s.C.Call("throttle.global_up.max_rate.set", "", bytesPerSec)
	return err
}
