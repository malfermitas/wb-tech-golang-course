package cleanup

import (
	"calendar_app/internal/repository/interfaces"
	"context"
	"log"
	"sync"
	"time"
)

type CleanupWorker struct {
	repo      interfaces.EventRepository
	ctx       context.Context
	interval  time.Duration
	olderThan time.Duration
	logger    *log.Logger
	wg        sync.WaitGroup
}

func NewCleanupWorker(repo interfaces.EventRepository, ctx context.Context, interval, olderThan time.Duration) *CleanupWorker {
	return &CleanupWorker{
		repo:      repo,
		ctx:       ctx,
		interval:  interval,
		olderThan: olderThan,
		logger:    log.New(log.Writer(), "[cleanup] ", log.LstdFlags),
	}
}

func (c *CleanupWorker) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		ticker := time.NewTicker(c.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-c.ctx.Done():
				c.logger.Println("Stopping cleanup worker")
				return
			}
		}
	}()
}

func (c *CleanupWorker) cleanup() {
	cutoff := time.Now().Add(-c.olderThan)

	count, err := c.repo.DeleteOldEvents(cutoff)
	if err != nil {
		c.logger.Printf("Error cleaning up events: %v", err)
		return
	}

	if count > 0 {
		c.logger.Printf("Deleted %d old events (older than %v)", count, c.olderThan)
	}
}

func (c *CleanupWorker) Stop() {
	c.wg.Wait()
}
