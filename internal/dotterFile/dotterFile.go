package dotterFile

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
