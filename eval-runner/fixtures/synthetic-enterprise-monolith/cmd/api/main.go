package main

import (
	"example.com/monolith/internal/auth"
	"example.com/monolith/internal/billing"
	"example.com/monolith/internal/notifications"
)

type App struct {
	Auth          *auth.AuthService
	Billing       *billing.BillingService
	Notifications *notifications.NotificationService
}

func NewApp(users auth.UserStore, payments billing.PaymentGateway, sender notifications.Sender) App {
	return App{
		Auth:          auth.NewAuthService(users),
		Billing:       billing.NewBillingService(payments),
		Notifications: notifications.NewNotificationService(sender),
	}
}
