package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"financial-risk-server/internal/config"
	"financial-risk-server/internal/service"
)

// MarketDataScheduler управляет фоновой синхронизацией рыночных данных
type MarketDataScheduler struct {
	syncService  *service.MarketDataSyncService
	config       *config.MarketDataConfig
	ticker       *time.Ticker
	stopChan     chan struct{}
	wg           sync.WaitGroup
	isRunning    bool
	mu           sync.RWMutex
}

// NewMarketDataScheduler создаёт новый планировщик
func NewMarketDataScheduler(
	syncService *service.MarketDataSyncService,
	config *config.MarketDataConfig,
) *MarketDataScheduler {
	return &MarketDataScheduler{
		syncService: syncService,
		config:      config,
		stopChan:    make(chan struct{}),
	}
}

// Start запускает планировщик
func (s *MarketDataScheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.isRunning {
		return fmt.Errorf("scheduler is already running")
	}

	if !s.config.Enabled {
		log.Println("⚠️  [Scheduler] Market data sync is disabled by config")
		return nil
	}

	interval := time.Duration(s.config.IntervalHours) * time.Hour
	s.ticker = time.NewTicker(interval)

	s.isRunning = true
	s.wg.Add(1)

	log.Printf("⏰ [Scheduler] Market data sync started (interval: %v, initial delay: %vs)",
		interval, s.config.InitialDelaySeconds)

	// Первая синхронизация с задержкой
	go func() {
		defer s.wg.Done()

		// Ждём initial delay
		select {
		case <-time.After(time.Duration(s.config.InitialDelaySeconds) * time.Second):
			s.runSync("initial")
		case <-s.stopChan:
			return
		}

		// Затем по расписанию
		for {
			select {
			case <-s.ticker.C:
				s.runSync("scheduled")
			case <-s.stopChan:
				s.ticker.Stop()
				log.Println("⏰ [Scheduler] Market data sync stopped")
				return
			}
		}
	}()

	return nil
}

// Stop останавливает планировщик
func (s *MarketDataScheduler) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
}

// SyncNow запускает синхронизацию вручную (для админ-эндпоинта)
func (s *MarketDataScheduler) SyncNow(ctx context.Context) error {
	log.Println("🔄 [Scheduler] Manual sync triggered")
	return s.runSync("manual")
}

// runSync выполняет синхронизацию с логированием
func (s *MarketDataScheduler) runSync(trigger string) error {
	start := time.Now()
	log.Printf("🚀 [Scheduler] Running %s market data sync", trigger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	if err := s.syncService.SyncAll(ctx); err != nil {
		log.Printf("❌ [Scheduler] %s sync failed after %v: %v",
			trigger, time.Since(start).Round(time.Second), err)
		return err
	}

	log.Printf("✅ [Scheduler] %s sync completed successfully in %v",
		trigger, time.Since(start).Round(time.Second))
	return nil
}

// IsRunning возвращает статус планировщика
func (s *MarketDataScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.isRunning
}

// GetLastSyncTime возвращает время последней успешной синхронизации
// (можно расширить, храня метаданные в БД или памяти)
func (s *MarketDataScheduler) GetLastSyncTime() time.Time {
	// TODO: реализовать при необходимости
	return time.Time{}
}