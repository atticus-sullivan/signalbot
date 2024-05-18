package dotterFile

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
	"io"
	"os/exec"
	"signalbot_go/internal/attachment"
	"strconv"
)

type Dotter interface {
	WriteDot(io.Writer) error
}

// CreateFigure function creates a PNG file from a Dotter graph by running the
// "dot" command with "-Tpng" option and writes the output to the specified
// file path. The Dotter interface is used to provide the input to the "dot"
// command. In case of any errors during the process, the function returns an
// error and closes the file, otherwise it returns the still-open file.
func CreateFigure(d Dotter, dpi float64) (ofile attachments.File, err error) {
	// create output file
	ofile, err = attachments.NewFileImpl("png")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil && ofile != nil {
			ofile.Close()
			ofile = nil
		}
	}()

	// Ausführen des dot-Befehls mit der Pipe als Eingabe
	cmd := exec.Command("dot", "-Tpng", "-o", ofile.Path(), "-Gdpi"+strconv.FormatFloat(dpi, 'f', 2, 64))

	var p io.WriteCloser
	p, err = cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	if err = cmd.Start(); err != nil {
		return
	}

	if err = d.WriteDot(p); err != nil {
		return
	}
	if err = p.Close(); err != nil {
		return
	}

	if err = cmd.Wait(); err != nil {
		return
	}
	return
}
