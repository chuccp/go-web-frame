package core

import "github.com/robfig/cron/v3"

type Schedule struct {
	cron *cron.Cron
}

func NewSchedule() *Schedule {
	return &Schedule{
		cron: cron.New(),
	}
}
func (c *Schedule) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	return c.cron.AddFunc(spec, cmd)
}
