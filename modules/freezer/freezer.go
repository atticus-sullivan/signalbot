package freezer

import (
	"fmt"
	"path/filepath"

	"signalbot_go/internal/dotterFile"
	"signalbot_go/internal/signalsender"
	"signalbot_go/modules"
	"signalbot_go/signaldbus"

	"github.com/alexflint/go-arg"
	freezerDB_db "github.com/atticus-sullivan/freezerDB/db"
	freezerDB_models "github.com/atticus-sullivan/freezerDB/db/models"
	"golang.org/x/exp/slog"
)

type Freezer struct {
	modules.Module
	db *freezerDB_db.DB `yaml:"-"`
}

func NewFreezer(log *slog.Logger, cfgDir string) (*Freezer, error) {
	r := Freezer{
		Module: modules.NewModule(log, cfgDir),
	}

	db, err := freezerDB_db.NewDB(filepath.Join(cfgDir, "freezer.yaml"))
	if err != nil {
		return nil, err
	}
	r.db = db

	// validation
	if err := r.Validate(); err != nil {
		return nil, err
	}
	if err := r.Module.Validate(); err != nil {
		return nil, err
	}

	return &r, nil
}

func (r *Freezer) Validate() error {
	return nil
}

type Args struct {
	ReportName *struct{} `arg:"subcommand:reportName"`
}

func (r *Freezer) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	var args Args
	parser, err := arg.NewParser(arg.Config{}, &args)
	if err != nil {
		r.Log.Error(fmt.Sprintf("newParser -> %v", err))
		return
	}

	if err := r.Module.Handle(m, signal, virtRcv, parser); err != nil {
		errMsg := err.Error()
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}

	var att []string
	switch {
	case args.ReportName != nil:
		var items freezerDB_models.FreezerItemList
		if err := r.db.DB.Select(&items, "SELECT * FROM freezer_items ORDER BY item_name"); err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}
		ofile, err := dotterFile.CreateFigure(items, 600)
		if err != nil {
			errMsg := fmt.Sprintf("Error: %v", err)
			r.Log.Error(errMsg)
			r.SendError(m, signal, errMsg)
			return
		}
		att = []string{ofile.Path()}
		defer ofile.Close()
	default:
		errMsg := fmt.Sprintf("Error: %v", "unknown/no subcommand")
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
		return
	}
	_, err = signal.Respond("", att, m, true)
	if err != nil {
		errMsg := fmt.Sprintf("Error: %v", err)
		r.Log.Error(errMsg)
		r.SendError(m, signal, errMsg)
	}
}

func (r *Freezer) Close(virtRcv func(*signaldbus.Message)) {
	r.Module.Close(virtRcv)

	r.db.Close()
}
