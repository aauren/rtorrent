package rtorrent

// fileListMultiCall retrieves a download's files along with subsequent commands to call on each
const fileListMultiCall = "f.multicall"

// A FileService is a wrapper for Client methods which operate on the files within a download.
type FileService struct {
	C Client
}

// FilesWithDetails retrieves one row per file in a download, holding the result of each command in order. Commands
// take the d.multicall2 form, e.g. "f.path=".
func (s *FileService) FilesWithDetails(infoHash string, commands []string) ([][]any, error) {
	return s.C.getSliceSliceByHash(fileListMultiCall, append([]string{infoHash}, commands...)...)
}
