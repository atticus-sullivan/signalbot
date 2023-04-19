package show

import (
	"fmt"
	"time"
)

// represents a tv show
type Show struct {
	Date time.Time `yaml:"time"`
	Name string    `yaml:"name"`
}

// stringer
func (s *Show) String() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("%s -> %s", s.Date.Format("2006-01-02 15:04"), s.Name)
}
