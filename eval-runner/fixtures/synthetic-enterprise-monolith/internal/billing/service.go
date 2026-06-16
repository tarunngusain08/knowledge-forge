package billing

type PaymentGateway interface {
	Charge(customerID string, cents int) error
}

type BillingService struct {
	gateway PaymentGateway
}

func NewBillingService(gateway PaymentGateway) *BillingService {
	return &BillingService{gateway: gateway}
}

func (s *BillingService) ChargeInvoice(customerID string, cents int) error {
	return s.gateway.Charge(customerID, cents)
}
