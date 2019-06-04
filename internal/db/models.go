package db

import "time"

// ExtSession extends the Session proto struct with calculated information
type ExtSession struct {
	Session
	Key           string
	Begin         time.Time
	PageviewCount int
}

// ExtPageview extends the Pageview proto struct with calculated information
type ExtPageview struct {
	Pageview
	Time time.Time
}
