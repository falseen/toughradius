package toughradius

import (
	"sync"
	"sync/atomic"
	"time"
)

type RejectItem struct {
	Rejects    int64
	LastReject time.Time
	Lock       sync.RWMutex
}

func (ri *RejectItem) Incr() {
	ri.Lock.Lock()
	defer ri.Lock.Unlock()
	atomic.AddInt64(&ri.Rejects, 1)
	ri.LastReject = time.Now()
}

func (ri *RejectItem) IsOver(max int64) bool {
	// 读取阶段先加读锁
	ri.Lock.RLock()

	// 如果距离上次拒绝超过 10 秒，需要清零计数。
	if time.Since(ri.LastReject).Seconds() > 10 {
		// 先释放读锁，再加写锁做修改
		ri.Lock.RUnlock()

		ri.Lock.Lock()
		if time.Since(ri.LastReject).Seconds() > 10 {
			atomic.StoreInt64(&ri.Rejects, 0)
		}
		ri.Lock.Unlock()
		return false
	}

	// 正常路径：直接判断次数是否超限
	over := atomic.LoadInt64(&ri.Rejects) > max
	ri.Lock.RUnlock()
	return over
}

type RejectCache struct {
	Items map[string]*RejectItem
	Lock  sync.RWMutex
}

func (rc *RejectCache) GetItem(username string) *RejectItem {
	rc.Lock.RLock()

	// 如果缓存过大，进行清理，需要升级为写锁
	if len(rc.Items) >= 65535 {
		rc.Lock.RUnlock()

		rc.Lock.Lock()
		if len(rc.Items) >= 65535 {
			rc.Items = make(map[string]*RejectItem, 0)
		}
		rc.Lock.Unlock()
		return nil
	}

	item := rc.Items[username]
	rc.Lock.RUnlock()
	return item
}

func (rc *RejectCache) SetItem(username string) {
	rc.Lock.Lock()
	defer rc.Lock.Unlock()
	if _, ok := rc.Items[username]; !ok {
		rc.Items[username] = &RejectItem{
			Rejects:    1,
			LastReject: time.Now(),
		}
	}
}
