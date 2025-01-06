package validec

import "time"

type Time struct {
	string
	time.Time
}

type validTypes struct {
	Time
	int
	string
	bool
}
