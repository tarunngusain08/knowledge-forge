package orders

import (
	"example.com/monolith/internal/audit"
	"example.com/monolith/internal/billing"
	"example.com/monolith/internal/notifications"
)

type Store interface {
	Save(order Order) error
}

type Order struct {
	ID         string
	CustomerID string
	Email      string
	Cents      int
}

type OrderService struct {
	store         Store
	billing       *billing.BillingService
	notifications *notifications.NotificationService
	audit         *audit.Logger
}

func NewOrderService(store Store, billing *billing.BillingService, notifications *notifications.NotificationService, audit *audit.Logger) *OrderService {
	return &OrderService{store: store, billing: billing, notifications: notifications, audit: audit}
}

func (s *OrderService) PlaceOrder(order Order) error {
	if err := s.store.Save(order); err != nil {
		return err
	}
	if err := s.billing.ChargeInvoice(order.CustomerID, order.Cents); err != nil {
		return err
	}
	s.audit.Record("order_placed")
	return s.notifications.SendWelcomeEmail(order.Email)
}
