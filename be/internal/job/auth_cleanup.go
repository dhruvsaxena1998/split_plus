package job

import (
	"context"
	"log"
	"time"

	"github.com/dhruvsaxena1998/splitplus/internal/service"
)

type AuthCleanup struct {
	authService service.AuthService
	ticker      *time.Ticker
	done        chan bool
}

func NewAuthCleanup(authService service.AuthService) *AuthCleanup {
	return &AuthCleanup{
		authService: authService,
		done:        make(chan bool),
	}
}

func (c *AuthCleanup) Start(ctx context.Context) {
	// Run cleanup every hour
	c.ticker = time.NewTicker(1 * time.Hour)

	go func() {
		// Run immediately on startup
		c.cleanup(ctx)

		// Then run on ticker
		for {
			select {
			case <-c.ticker.C:
				c.cleanup(ctx)
			case <-ctx.Done():
				return
			case <-c.done:
				return
			}
		}
	}()
}

func (c *AuthCleanup) Stop() {
	if c.ticker != nil {
		c.ticker.Stop()
	}
	close(c.done)
}

func (c *AuthCleanup) cleanup(ctx context.Context) {
	log.Println("Running auth cleanup: removing expired sessions and blacklisted tokens...")

	err := c.authService.CleanupExpiredSessions(ctx)
	if err != nil {
		log.Printf("Error during auth cleanup: %v", err)
		return
	}

	log.Println("Auth cleanup completed successfully")
}
