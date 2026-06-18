package orders

import (
	"testing"

	"example.com/monolith/internal/audit"
	"example.com/monolith/internal/billing"
	"example.com/monolith/internal/notifications"
)

func TestPlaceOrderChargesAndNotifies(t *testing.T) {
	store := &fakeStore{}
	payments := &fakePayments{}
	sender := &fakeSender{}
	service := NewOrderService(store, billing.NewBillingService(payments), notifications.NewNotificationService(sender), audit.NewLogger("test"))

	if err := service.PlaceOrder(Order{ID: "o-1", CustomerID: "c-1", Email: "buyer@example.com", Cents: 4200}); err != nil {
		t.Fatalf("err = %v", err)
	}
	if !store.saved || !payments.charged || !sender.sent {
		t.Fatalf("workflow did not complete: store=%v payments=%v sender=%v", store.saved, payments.charged, sender.sent)
	}
}

type fakeStore struct {
	saved bool
}

func (f *fakeStore) Save(order Order) error {
	f.saved = true
	return nil
}

type fakePayments struct {
	charged bool
}

func (f *fakePayments) Charge(customerID string, cents int) error {
	f.charged = true
	return nil
}

type fakeSender struct {
	sent bool
}

func (f *fakeSender) Send(to string, subject string, body string) error {
	f.sent = true
	return nil
}
