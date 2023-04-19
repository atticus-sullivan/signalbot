package attachments

import (
	"errors"
	"os"
	"regexp"
)

// represents a temporary file which is being removed when getting closed.
type File interface {
	// returns the Path of the temporary File
	Path() string
	// returns the temporary File
	File() *os.File
	// closes and removes the temporary File
	Close() error
}

// implements the file interface
type FileImpl struct {
	path string
	file *os.File
}

// errors
var (
	ErrInvalidExt error = errors.New("Invalid extension")
)

var extRe *regexp.Regexp = regexp.MustCompile(`[a-zA-Z0-9]`)

// create a new temporary namend file with the given extension. Currently only
// alphanumeric extensions are allowed
func NewFileImpl(ext string) (*FileImpl, error) {
	if !extRe.Match([]byte(ext)) {
		return nil, ErrInvalidExt
	}
	tmpfile, err := os.CreateTemp("", "signalbot_go-*."+ext)
	if err != nil {
		return nil, err
	}
	return &FileImpl{
		path: tmpfile.Name(),
		file: tmpfile,
	}, nil
}

// returns the Path of the temporary File
func (fi *FileImpl) Path() string {
	return fi.path
}

// returns the temporary File
func (fi *FileImpl) File() *os.File {
	return fi.file
}

// closes and removes the temporary File
func (fi *FileImpl) Close() error {
	errA := fi.file.Close()
	errB := os.Remove(fi.path)
	if errA != nil {
		return errA
	}
	return errB
}
