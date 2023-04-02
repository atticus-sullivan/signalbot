package show

import (
	"fmt"
	"time"
)

type Show struct {
	Time time.Time `yaml:"time"`
	Name string    `yaml:"name"`
}

func (s *Show) String() string {
	if s == nil {
		return ""
	}
	return fmt.Sprintf("%s -> %s", s.Time.Format("2006-01-02 15:04"), s.Name)
}
