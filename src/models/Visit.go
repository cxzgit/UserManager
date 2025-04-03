package models

import "time"

type Visit struct {
	ID        int
	UserID    int
	VisitTime time.Time
}
