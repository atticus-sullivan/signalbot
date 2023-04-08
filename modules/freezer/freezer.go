package freezer

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	// "os"
	cmdsplit "signalbot_go/internal/cmdSplit"
	"signalbot_go/internal/dotterFile"
	"signalbot_go/internal/signalsender"
	"signalbot_go/signaldbus"

	"github.com/alexflint/go-arg"
	freezerDB_db "github.com/atticus-sullivan/freezerDB/db"
	freezerDB_models "github.com/atticus-sullivan/freezerDB/db/models"
	"golang.org/x/exp/slog"
)

type Freezer struct {
	log       *slog.Logger     `yaml:"-"`
	ConfigDir string           `yaml:"-"`
	db        *freezerDB_db.DB `yaml:"-"`
}

func NewFreezer(log *slog.Logger, cfgDir string) (*Freezer, error) {
	r := Freezer{
		log:       log,
		ConfigDir: cfgDir,
	}

	db, err := freezerDB_db.NewDB(filepath.Join(cfgDir, "freezer.yaml"))
	if err != nil {
		return nil, err
	}
	r.db = db

	// f, err := os.Open(filepath.Join(r.ConfigDir, "freezer.yaml"))
	// if err != nil {
	// 	return nil, err
	// }
	// defer f.Close()
	//
	// d := yaml.NewDecoder(f)
	// d.KnownFields(true)
	// err = d.Decode(&r)
	// if err != nil {
	// 	return nil, err
	// }

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (f *Freezer) Validate() error {
	return nil
}

func (f *Freezer) sendError(m *signaldbus.Message, signal signalsender.SignalSender, reply string) {
	if _, err := signal.Respond(reply, nil, m); err != nil {
		f.log.Error(fmt.Sprintf("Error responding to %v", m))
	}
}

type Args struct {
	ReportName *struct{} `arg:"subcommand:reportName"`
}

func (f *Freezer) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)

	if err != nil {
		f.log.Error(fmt.Sprintf("freezer module: newParser -> %v", err))
		return
	}

	vargs, err := cmdsplit.Split(m.Message)
	if err != nil {
		errMsg := fmt.Sprintf("freezer module: Error on parsing message: %v", err)
		f.log.Error(errMsg)
		f.sendError(m, signal, errMsg)
		return
	}

	err = parser.Parse(vargs)

	if err != nil {
		switch err {
		case arg.ErrVersion:
			// not implemented
			errMsg := fmt.Sprintf("freezer module: Error: %v", "Version is not implemented")
			f.log.Error(errMsg)
			f.sendError(m, signal, errMsg)
			return
		case arg.ErrHelp:
			buf := new(bytes.Buffer)
			parser.WriteHelp(buf)

			if b, err := io.ReadAll(buf); err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				f.log.Error(errMsg)
				f.sendError(m, signal, errMsg)
				return
			} else {
				errMsg := string(b)
				f.log.Info(fmt.Sprintf("freezer module: Error: %v", err))
				f.sendError(m, signal, errMsg)
				return
			}
		default:
			errMsg := fmt.Sprintf("Error: %v", err)
			f.log.Error(errMsg)
			f.sendError(m, signal, errMsg)
			return
		}
	} else {
		var att []string
		switch {
		case args.ReportName != nil:
			var items freezerDB_models.FreezerItemList
			if err := f.db.DB.Select(&items, "SELECT * FROM freezer_items ORDER BY item_name"); err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				f.log.Error(errMsg)
				f.sendError(m, signal, errMsg)
				return
			}
			ofile,err := dotterFile.CreateFigure(items, 600)
			if err != nil {
				errMsg := fmt.Sprintf("Error: %v", err)
				f.log.Error(errMsg)
				f.sendError(m, signal, errMsg)
				return
			}
			att = []string{ofile.Path()}
			defer ofile.Close()
		default:
			errMsg := fmt.Sprintf("Error: %v", "unknown/no subcommand")
			f.log.Error(errMsg)
			f.sendError(m, signal, errMsg)
			return
		}
		_, err = signal.Respond("", att, m)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			f.log.Error(errMsg)
			f.sendError(m, signal, errMsg)
		}
	}
}

func (f *Freezer) Start(virtRcv func(*signaldbus.Message)) error {
	return nil
}

func (f *Freezer) Close(virtRcv func(*signaldbus.Message)) {
	f.db.Close()
}
