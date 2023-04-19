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

// freezer module. Should be instanciated with `NewFreezer`.
// data members are only global to be able to unmarshal them
type Freezer struct {
	modules.Module
	db *freezerDB_db.DB `yaml:"-"`
}

// instanciates a new Freezer from a configuration file
// (cfgDir/freezer.yaml)
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

	return &r, nil
}

// can be used to validate the freezer struct
func (r *Freezer) Validate() error {
	// validate the generic module first
	if err := r.Module.Validate(); err != nil {
		return err
	}
	return nil
}

// specifies the arguments when handling a request to this module
type Args struct {
	ReportName *struct{} `arg:"subcommand:reportName"`
}

// Handle a message from the signaldbus. Parses the message, executes the query
// and responds to signal.
func (r *Freezer) Handle(m *signaldbus.Message, signal signalsender.SignalSender, virtRcv func(*signaldbus.Message)) {
	// parse the message
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

	// execute the query
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

	// respond
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
