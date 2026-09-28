package cache

import (
	"context"
	"testing"
	"time"

	config2 "github.com/chuccp/go-web-frame/config"
)

// 撞一撞 SetNX 的显式过期到底算不算数。
//
// 应用场景：**会增减的那种配置/清单数据**——内容变了必须到点重取，不能靠"没人读"来兜底。
// 这个方法给了 expire 参数，文档也写着 "stores a value with an expiration"，
// 那就得真的过期——**包括中途被读过**（读一次就把过期时间顶回默认值的话，
// 高频访问的键永不过期，正是最要命的场景）。
func TestSetNXExpiryHonored(t *testing.T) {
	c := newTestCache(t)

	if _, ok := c.SetNX("k", "v", 200*time.Millisecond); !ok {
		t.Fatal("首次 SetNX 应当成功")
	}
	if _, ok := c.Get("k"); !ok {
		t.Fatal("刚写进去就该读得到")
	}

	time.Sleep(400 * time.Millisecond)

	if v, ok := c.Get("k"); ok {
		t.Errorf("显式给的 200ms 过期没生效：400ms 后还读得到（%v）", v)
	}
}

// 同上，但一次都不读（排除"是读把它续上了"这个解释）。
func TestSetNXExpiryHonoredWithoutReads(t *testing.T) {
	c := newTestCache(t)
	c.SetNX("k2", "v", 200*time.Millisecond)
	time.Sleep(400 * time.Millisecond)
	if v, ok := c.Get("k2"); ok {
		t.Errorf("没人读过也没过期（%v）", v)
	}
}

// 没配 cache: 段时 UnmarshalKey 会不会报错——报错的话 Init 起不来，
// 整个框架挂在一个可选的配置段上。
func TestConfigMissingSection(t *testing.T) {
	cfg := config2.NewConfig()
	lConfig := &Config{MaxSize: 1000_000, Expiry: 3600}
	err := cfg.UnmarshalKey(ConfigKey, lConfig)
	t.Logf("缺 cache: 段时 UnmarshalKey 返回 err=%v，解出来是 %+v", err, lConfig)
	if err != nil {
		t.Errorf("没配 cache: 段就返回错误，Init 会直接失败：%v", err)
	}
	if lConfig.MaxSize != 1000_000 || lConfig.Expiry != 3600 {
		t.Errorf("缺配置段时默认值被冲掉了：%+v", lConfig)
	}
}

// Set 走的是默认过期（ExpiryAccessing）——写进去的东西能不能读到、能不能过期。
func TestSetUsesDefaultExpiry(t *testing.T) {
	c := newTestCache(t)
	c.Set("k3", "v")
	if _, ok := c.Get("k3"); !ok {
		t.Error("Set 之后读不到")
	}
}

// ComputeIfAbsent 的 cancel=true 应当"不写入"（文档这么写的），验一下。
func TestComputeIfAbsentCancel(t *testing.T) {
	c := newTestCache(t)
	v, ok := c.ComputeIfAbsent("k4", func() (any, bool) { return "nope", true })
	if _, exists := c.Get("k4"); exists {
		t.Errorf("cancel=true 还是写进去了：ComputeIfAbsent 返回 (%v,%v)", v, ok)
	}
}

// 过期时间给 0 / 负数时不该炸，也不该变成"立刻过期"。
func TestSetNXZeroAndNegativeExpiry(t *testing.T) {
	c := newTestCache(t)
	if _, ok := c.SetNX("k5", "v", 0); !ok {
		t.Error("expire=0 时 SetNX 应当照常写入（用默认过期）")
	}
	if _, ok := c.Get("k5"); !ok {
		t.Error("expire=0 写进去的立刻就读不到了")
	}
	if _, ok := c.SetNX("k6", "v", -time.Second); !ok {
		t.Error("expire 负数时 SetNX 也不该失败")
	}
	if _, ok := c.Get("k6"); !ok {
		t.Error("expire 负数写进去的立刻就读不到了")
	}
}

// Invalidate 之后应当读不到。
func TestInvalidateThenGet(t *testing.T) {
	c := newTestCache(t)
	c.Set("k7", "v")
	if _, ok := c.Invalidate("k7"); !ok {
		t.Error("Invalidate 一个存在的键应当返回 true")
	}
	if _, ok := c.Get("k7"); ok {
		t.Error("Invalidate 之后还读得到")
	}
}

var _ = context.Background
