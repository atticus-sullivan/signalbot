package attachments

// signalbot
// Copyright (C) 2024  Lukas Heindl
// 
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// 
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// 
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
