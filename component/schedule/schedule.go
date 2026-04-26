package schedule

import (
	"sync"
	"time"

	"github.com/chuccp/go-web-frame/core"
	"github.com/robfig/cron/v3"
)

type Info struct {
	entryID    cron.EntryID
	UpdateTime time.Time
	Key        string
	id         uint
}

type Schedule struct {
	cron      *cron.Cron
	infoMap   map[string]*Info
	lock      *sync.RWMutex
	idInfoMap map[uint]*Info
	ctx       *core.Context
}

func NewSchedule(opts ...cron.Option) *Schedule {
	return &Schedule{
		cron:      cron.New(opts...),
		infoMap:   make(map[string]*Info),
		idInfoMap: make(map[uint]*Info),
		lock:      new(sync.RWMutex),
	}
}
func NewScheduleWithSeconds() *Schedule {
	return &Schedule{
		cron:      cron.New(cron.WithSeconds()),
		infoMap:   make(map[string]*Info),
		idInfoMap: make(map[uint]*Info),
		lock:      new(sync.RWMutex),
	}
}
func (c *Schedule) AddFunc(spec string, cmd func(context2 *core.Context)) (cron.EntryID, error) {
	return c.cron.AddFunc(spec, func() {
		c.ctx.Go(cmd)
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
func (c *Schedule) ReplaceKeyFunc(key string, spec string, cmd func(context2 *core.Context)) (cron.EntryID, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	v, err := c.cron.AddFunc(spec, func() {
		c.ctx.Go(cmd)
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
func (c *Schedule) AddKeyFunc(key string, spec string, cmd func(context2 *core.Context)) (cron.EntryID, bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	_, ok := c.infoMap[key]
	if ok {
		return 0, ok, nil
	}
	v, err := c.cron.AddFunc(spec, func() {
		c.ctx.Go(cmd)
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
func (c *Schedule) GetIds() []uint {
	ids := make([]uint, 0)
	for _, info := range c.idInfoMap {
		ids = append(ids, info.id)
	}
	return ids
}
func (c *Schedule) AddIdOrReplaceKeyFunc(id uint, key string, spec string, cmd func(context2 *core.Context)) (cron.EntryID, bool, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	info, ok := c.idInfoMap[id]
	if ok {
		if info.Key == key {
			return info.entryID, ok, nil
		}
	}
	v, err := c.cron.AddFunc(spec, func() {
		c.ctx.Go(cmd)
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

func (c *Schedule) Init(ctx *core.Context) error {
	c.ctx = ctx
	return nil
}
func (c *Schedule) Run() error {
	go func() {
		<-c.ctx.Done()
		c.cron.Stop()
	}()
	c.cron.Start()
	return nil
}
