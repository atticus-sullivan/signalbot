package show

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
