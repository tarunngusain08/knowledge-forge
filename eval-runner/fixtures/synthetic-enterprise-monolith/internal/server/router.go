package server

import (
	"example.com/monolith/internal/auth"
	"example.com/monolith/internal/billing"
	"example.com/monolith/internal/notifications"
	"example.com/monolith/internal/orders"
)

type Route struct {
	Method  string
	Path    string
	Handler string
}

type Router struct {
	routes []Route
}

func NewRouter(authService *auth.AuthService, billingService *billing.BillingService, notificationsService *notifications.NotificationService, orderService *orders.OrderService) *Router {
	_ = authService
	_ = billingService
	_ = notificationsService
	_ = orderService
	return &Router{routes: []Route{
		{Method: "POST", Path: "/login", Handler: "AuthService.Login"},
		{Method: "POST", Path: "/invoices/charge", Handler: "BillingService.ChargeInvoice"},
		{Method: "POST", Path: "/orders", Handler: "OrderService.PlaceOrder"},
		{Method: "POST", Path: "/welcome-email", Handler: "NotificationService.SendWelcomeEmail"},
	}}
}

func (r *Router) Routes() []Route {
	return append([]Route{}, r.routes...)
}
