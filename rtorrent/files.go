package rtorrent

import "fmt"

// fileListMultiCall retrieves a download's files along with subsequent commands to call on each
const fileListMultiCall = "f.multicall"

// FilePriority is a file's f.priority
type FilePriority int

const (
	FilePriorityOff FilePriority = iota
	FilePriorityNormal
	FilePriorityHigh
)

// A FileService is a wrapper for Client methods which operate on the files within a download.
type FileService struct {
	C Client
}

// FilesWithDetails retrieves one row per file in a download, holding the result of each command in order. Commands
// take the d.multicall2 form, e.g. "f.path=".
func (s *FileService) FilesWithDetails(infoHash string, commands []string) ([][]any, error) {
	return s.C.getSliceSliceByHash(fileListMultiCall, append([]string{infoHash}, commands...)...)
}

// SetPriority sets the priority of the file at index, in f.multicall order, within a download. Nothing changes until
// DownloadService.UpdatePriorities is called, so several files can be set and applied at once.
func (s *FileService) SetPriority(infoHash string, index int, priority FilePriority) error {
	if infoHash == "" || index < 0 {
		return fmt.Errorf("%w: f.priority.set requires an info-hash and a file index", ErrBadData)
	}
	_, err := s.C.Call("f.priority.set", fmt.Sprintf("%s:f%d", infoHash, index), int(priority))
	return err
}
