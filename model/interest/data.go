package interest

import (
	"github.com/awakari/interests/model/condition"
	"time"
)

type Data struct {

	// Description is human readable interest description
	Description string

	// Enabled defines whether the interest may be used for the matching
	Enabled bool

	// Expires defines a deadline when the interest is treated as Enabled
	Expires time.Time

	// Created represents the interest creation time.
	Created time.Time

	// Updated represents the interest last update time.
	Updated time.Time

	// Result represents the last read result time.
	Result time.Time

	Public bool

	Followers int64

	// Condition represents the certain criteria to select the Interest for the further routing.
	// It's immutable once the Interest is created.
	Condition condition.Condition
}
