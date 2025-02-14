package interest

import "time"

type Cursor struct {
	Id        string
	Followers int64
	CreatedAt time.Time
}
