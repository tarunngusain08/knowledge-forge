package main

import (
	"example.com/monolith/internal/audit"
	"example.com/monolith/internal/auth"
	"example.com/monolith/internal/billing"
	"example.com/monolith/internal/notifications"
	"example.com/monolith/internal/orders"
	"example.com/monolith/internal/server"
)

type App struct {
	Auth          *auth.AuthService
	Billing       *billing.BillingService
	Notifications *notifications.NotificationService
	Orders        *orders.OrderService
	Router        *server.Router
	Audit         *audit.Logger
}

func NewApp(users auth.UserStore, payments billing.PaymentGateway, sender notifications.Sender, orderStore orders.Store) App {
	auditLog := audit.NewLogger("api")
	authService := auth.NewAuthService(users)
	billingService := billing.NewBillingService(payments)
	notificationService := notifications.NewNotificationService(sender)
	orderService := orders.NewOrderService(orderStore, billingService, notificationService, auditLog)
	router := server.NewRouter(authService, billingService, notificationService, orderService)
	return App{
		Auth:          authService,
		Billing:       billingService,
		Notifications: notificationService,
		Orders:        orderService,
		Router:        router,
		Audit:         auditLog,
	}
}
