package core

import (
	"sync"
	"time"

	"emperror.dev/errors"

	config2 "github.com/chuccp/go-web-frame/config"
	"github.com/chuccp/go-web-frame/log"
	"github.com/robfig/cron/v3"
	"github.com/sourcegraph/conc/panics"
)

type Info struct {
	entryID    cron.EntryID
	UpdateTime time.Time
	Key        string
	id         uint
}

type ScheduleConfig struct {
	Enable bool
}

func (c *ScheduleConfig) Key() string {
	return "web.schedule"
}

type Schedule struct {
	cron      *cron.Cron
	infoMap   map[string]*Info
	lock      *sync.RWMutex
	config    *ScheduleConfig
	idInfoMap map[uint]*Info
}

func NewSchedule() *Schedule {
	return &Schedule{
		cron:      cron.New(),
		infoMap:   make(map[string]*Info),
		idInfoMap: make(map[uint]*Info),
		lock:      new(sync.RWMutex),
		config:    &ScheduleConfig{Enable: false},
	}
}
func (c *Schedule) AddFunc(spec string, cmd func()) (cron.EntryID, error) {
	if !c.config.Enable {
		return 0, errors.New("schedule is not enable")
	}
	return c.cron.AddFunc(spec, func() {
		var catcher panics.Catcher
		catcher.Try(cmd)
		log.Errors(spec, catcher.Recovered().AsError())
	})
}
func (c *Schedule) StopKeyFunc(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if info, ok := c.infoMap[key]; ok {
		delete(c.infoMap, key)
		c.cron.Remove(info.entryID)
	}
}
func (c *Schedule) ReplaceKeyFunc(key string, spec string, cmd func()) (cron.EntryID, error) {
	if !c.config.Enable {
		return 0, errors.New("schedule is not enable")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	v, err := c.cron.AddFunc(spec, func() {
		var catcher panics.Catcher
		catcher.Try(cmd)
		log.Errors(spec, catcher.Recovered().AsError())
	})
	if err != nil {
		return 0, err
	}
	if preInfo, ok := c.infoMap[key]; ok {
		delete(c.infoMap, key)
		c.cron.Remove(preInfo.entryID)
	}
	info := &Info{
		entryID:    v,
		UpdateTime: time.Now(),
		Key:        key,
	}
	c.infoMap[key] = info
	return v, err
}
func (c *Schedule) AddKeyFunc(key string, spec string, cmd func()) (cron.EntryID, bool, error) {
	if !c.config.Enable {
		return 0, false, errors.New("schedule is not enable")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.infoMap[key]
	if ok {
		return 0, ok, nil
	}
	v, err := c.cron.AddFunc(spec, func() {
		var catcher panics.Catcher
		catcher.Try(cmd)
		log.Errors(spec, catcher.Recovered().AsError())
	})
	if err != nil {
		return 0, ok, err
	}
	info := &Info{
		entryID:    v,
		UpdateTime: time.Now(),
		Key:        key,
	}
	c.infoMap[key] = info
	return v, ok, err
}
func (c *Schedule) StopIdFunc(id uint) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if info, ok := c.idInfoMap[id]; ok {
		delete(c.idInfoMap, id)
		c.cron.Remove(info.entryID)
	}
}
func (c *Schedule) AddIdOrReplaceKeyFunc(id uint, key string, spec string, cmd func()) (cron.EntryID, bool, error) {
	if !c.config.Enable {
		return 0, false, errors.New("schedule is not enable")
	}
	c.lock.Lock()
	defer c.lock.Unlock()
	info, ok := c.idInfoMap[id]
	if ok {
		if info.Key == key {
			return info.entryID, ok, nil
		}
	}
	v, err := c.cron.AddFunc(spec, func() {
		var catcher panics.Catcher
		catcher.Try(cmd)
		log.Errors(spec, catcher.Recovered().AsError())
	})
	if err != nil {
		return 0, ok, err
	}
	info = &Info{
		entryID:    v,
		UpdateTime: time.Now(),
		Key:        key,
		id:         id,
	}
	c.idInfoMap[id] = info
	return v, ok, err
}

func (c *Schedule) Init(config config2.IConfig) error {

	err := config.Unmarshal(c.config.Key(), c.config)
	if err != nil {
		return errors.WithStackIf(err)
	}
	if c.config.Enable {
		c.Start()
	}
	return nil
}
func (c *Schedule) Name() string {
	return "schedule"
}

func (c *Schedule) Start() {
	c.cron.Start()
}
func (c *Schedule) Stop() {
	if c.config.Enable {
		c.cron.Stop()
	}
}
