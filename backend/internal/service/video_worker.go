package service

import (
	"context"
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	videoWorkerLeaderLockKey = "video:generation:worker:leader"
	videoWorkerLeaderLockTTL = 90 * time.Second
)

type VideoWorker struct {
	videoService *VideoService
	lockCache    LeaderLockCache
	db           *sql.DB
	instanceID   string
	interval     time.Duration
	batchSize    int
	stopCh       chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup
}

func NewVideoWorker(videoService *VideoService) *VideoWorker {
	return &VideoWorker{
		videoService: videoService,
		instanceID:   uuid.NewString(),
		interval:     5 * time.Second,
		batchSize:    20,
		stopCh:       make(chan struct{}),
	}
}

func (w *VideoWorker) SetLeaderLock(lockCache LeaderLockCache, db *sql.DB) {
	if w == nil {
		return
	}
	w.lockCache = lockCache
	w.db = db
}

func (w *VideoWorker) Start() {
	if w == nil || w.videoService == nil {
		return
	}
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		w.runOnce()
		for {
			select {
			case <-ticker.C:
				w.runOnce()
			case <-w.stopCh:
				return
			}
		}
	}()
}

func (w *VideoWorker) Stop() {
	if w == nil {
		return
	}
	w.stopOnce.Do(func() { close(w.stopCh) })
	w.wg.Wait()
}

func (w *VideoWorker) runOnce() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	release, ok := tryAcquireSingletonLeaderLock(ctx, w.lockCache, w.db, videoWorkerLeaderLockKey, w.instanceID, videoWorkerLeaderLockTTL)
	if !ok {
		return
	}
	defer release()
	processed, err := w.videoService.ProcessDueTasks(ctx, w.batchSize)
	if err != nil {
		log.Printf("[VideoWorker] process due tasks failed: %v", err)
		return
	}
	if processed > 0 {
		log.Printf("[VideoWorker] processed %d video tasks", processed)
	}
}
