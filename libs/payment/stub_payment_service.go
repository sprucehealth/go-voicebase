package payment

import "github.com/sprucehealth/backend/common"

type StubPaymentService struct {
	CustomerToReturn  *Customer
	CardToReturnOnAdd *common.Card
	CardsToReturn     []*common.Card
}

func (s *StubPaymentService) CreateCustomerWithDefaultCard(token string) (*Customer, error) {
	return s.CustomerToReturn, nil
}

func (s *StubPaymentService) AddCardForCustomer(cardToken string, customerId string) (*common.Card, error) {
	return s.CardToReturnOnAdd, nil
}

func (s *StubPaymentService) MakeCardDefaultForCustomer(cardId string, customerId string) error {
	return nil
}

func (s *StubPaymentService) GetCardsForCustomer(customerId string) ([]*common.Card, error) {
	return s.CardsToReturn, nil
}

func (s *StubPaymentService) DeleteCardForCustomer(customerId string, cardId string) error {
	return nil
}
