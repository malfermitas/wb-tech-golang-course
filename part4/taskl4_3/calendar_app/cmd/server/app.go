package main

import (
	"calendar_app/internal/delivery/http/handlers"
	"calendar_app/internal/repository/in_memory"
	"calendar_app/internal/repository/interfaces"
	"calendar_app/internal/service/cleanup"
	"calendar_app/internal/service/event"
	"calendar_app/internal/service/notification"
	"context"
	"time"
)

type App struct {
	EventRepo     interfaces.EventRepository
	EventNotifier *notification.EventNotificationService
	CleanupWorker *cleanup.CleanupWorker
	EventService  *event.Service
	EventHandler  *handlers.EventHandler
	ctx           context.Context
	cancel        context.CancelFunc
}

func NewApp() *App {
	ctx, cancel := context.WithCancel(context.Background())

	eventRepo := inmemory.NewEventRepository()
	notifier := notification.NewEventNotifier(eventRepo, ctx)

	cleanupWorker := cleanup.NewCleanupWorker(
		eventRepo,
		ctx,
		10*time.Minute,
		30*24*time.Hour, // 30 days
	)

	eventService := event.NewService(eventRepo, notifier)
	eventHandler := handlers.NewEventHandler(eventService)

	return &App{
		EventRepo:     eventRepo,
		EventNotifier: notifier,
		CleanupWorker: cleanupWorker,
		EventService:  eventService,
		EventHandler:  eventHandler,
		ctx:           ctx,
		cancel:        cancel,
	}
}

func (a *App) Start() {
	a.EventNotifier.Start()
	a.CleanupWorker.Start()
}

func (a *App) Stop() {
	a.cancel()
	a.CleanupWorker.Stop()
}
