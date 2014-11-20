package cost

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/sprucehealth/backend/Godeps/_workspace/src/github.com/samuel/go-metrics/metrics"
	"github.com/sprucehealth/backend/analytics"
	"github.com/sprucehealth/backend/api"
	"github.com/sprucehealth/backend/apiservice"
	"github.com/sprucehealth/backend/common"
	"github.com/sprucehealth/backend/email"
	"github.com/sprucehealth/backend/libs/dispatch"
	"github.com/sprucehealth/backend/libs/golog"
	"github.com/sprucehealth/backend/libs/stripe"
	"github.com/sprucehealth/backend/sku"
)

const (
	batchSize               = 1
	visibilityTimeout       = 60 * 5
	waitTimeSeconds         = 20
	timeBetweenEmailRetries = 10
	receiptNumberMax        = 99999
	receiptNumDigits        = 5
	defaultTimePeriod       = 60
)

type Worker struct {
	dataAPI             api.DataAPI
	analyticsLogger     analytics.Logger
	dispatcher          *dispatch.Dispatcher
	stripeCli           apiservice.StripeClient
	emailService        email.Service
	supportEmail        string
	queue               *common.SQSQueue
	chargeSuccess       *metrics.Counter
	chargeFailure       *metrics.Counter
	receiptSendSuccess  *metrics.Counter
	receiptSendFailure  *metrics.Counter
	timePeriodInSeconds int
	stopChan            chan bool
}

func StartWorker(dataAPI api.DataAPI, analyticsLogger analytics.Logger, dispatcher *dispatch.Dispatcher,
	stripeCli apiservice.StripeClient, emailService email.Service,
	queue *common.SQSQueue, metricsRegistry metrics.Registry,
	timePeriodInSeconds int, supportEmail string) *Worker {
	if timePeriodInSeconds == 0 {
		timePeriodInSeconds = defaultTimePeriod
	}

	chargeSuccess := metrics.NewCounter()
	chargeFailure := metrics.NewCounter()
	receiptSendSuccess := metrics.NewCounter()
	receiptSendFailure := metrics.NewCounter()

	metricsRegistry.Add("case_charge/success", chargeSuccess)
	metricsRegistry.Add("case_charge/failure", chargeFailure)
	metricsRegistry.Add("receipt_send/success", receiptSendSuccess)
	metricsRegistry.Add("receipt_send/failure", receiptSendFailure)

	w := &Worker{
		dataAPI:             dataAPI,
		analyticsLogger:     analyticsLogger,
		dispatcher:          dispatcher,
		stripeCli:           stripeCli,
		emailService:        emailService,
		supportEmail:        supportEmail,
		queue:               queue,
		chargeSuccess:       chargeSuccess,
		chargeFailure:       chargeFailure,
		receiptSendSuccess:  receiptSendSuccess,
		receiptSendFailure:  receiptSendFailure,
		timePeriodInSeconds: timePeriodInSeconds,
		stopChan:            make(chan bool),
	}

	w.start()

	return w
}

func (w *Worker) start() {
	go func() {
		for {
			select {
			case <-w.stopChan:
				return
			default:
			}

			msgConsumed, err := w.consumeMessage()
			if err != nil {
				golog.Errorf(err.Error())
			}
			if !msgConsumed {
				time.Sleep(time.Duration(w.timePeriodInSeconds) * time.Second)
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.stopChan)
}

func (w *Worker) consumeMessage() (bool, error) {
	msgs, err := w.queue.QueueService.ReceiveMessage(w.queue.QueueUrl, nil, batchSize, visibilityTimeout, waitTimeSeconds)
	if err != nil {
		return false, err
	}

	allMsgsConsumed := len(msgs) > 0

	for _, m := range msgs {
		v := &VisitMessage{}
		if err := json.Unmarshal([]byte(m.Body), v); err != nil {
			return false, err
		}

		if err := w.processMessage(v); err != nil {
			golog.Errorf(err.Error())
			allMsgsConsumed = false
		} else {
			if err := w.queue.QueueService.DeleteMessage(w.queue.QueueUrl, m.ReceiptHandle); err != nil {
				golog.Errorf(err.Error())
				allMsgsConsumed = false
			}
		}
	}

	return allMsgsConsumed, nil
}

func (w *Worker) processMessage(m *VisitMessage) error {
	patient, err := w.dataAPI.GetPatientFromPatientVisitId(m.PatientVisitID)
	if err != nil {
		return err
	} else if patient.Training {
		return nil
	}

	// get the cost of the visit
	costBreakdown, err := totalCostForItems([]sku.SKU{m.ItemType}, m.AccountID, true, w.dataAPI, w.analyticsLogger)
	if err != nil {
		return err
	}

	pReceipt, err := w.retrieveOrCreatePatientReceipt(m.PatientID,
		m.PatientVisitID,
		costBreakdown.ItemCosts[0].ID,
		m.ItemType,
		costBreakdown)
	if err != nil {
		return err
	}

	currentStatus := pReceipt.Status
	nextStatus := common.PRCharged
	patientReceiptUpdate := &api.PatientReceiptUpdate{Status: &nextStatus}

	if costBreakdown.TotalCost.Amount > 0 && currentStatus == common.PRChargePending {
		// check if the charge already exists for the customer
		var charge *stripe.Charge
		charges, err := w.stripeCli.ListAllCustomerCharges(patient.PaymentCustomerId)
		if err != nil {
			return err
		}
		for _, cItem := range charges {
			if refNum, ok := cItem.Metadata["receipt_ref_num"]; ok && refNum == pReceipt.ReferenceNumber {
				charge = cItem
				break
			}
		}

		// if a charge exists, get the card used for the charge, else get the default card for the customer
		var card *common.Card
		if charge != nil {
			card, err = w.dataAPI.GetCardFromThirdPartyID(charge.Card.ID)
			if err != nil && err != api.NoRowsError {
				return err
			}
		} else if m.CardID != 0 {
			card, err = w.dataAPI.GetCardFromId(m.CardID)
			if err != nil {
				return err
			}
		} else {
			// get the default card of the patient from the visit that we are going to charge
			card, err = w.dataAPI.GetDefaultCardForPatient(m.PatientID)
			if err == api.NoRowsError {
				return errors.New("No default card for patient")
			} else if err != nil {
				return err
			}
		}

		// only create a charge if one doesn't already exist for the customer
		if charge == nil {
			charge, err = w.stripeCli.CreateChargeForCustomer(&stripe.CreateChargeRequest{
				Amount:       costBreakdown.TotalCost.Amount,
				CurrencyCode: costBreakdown.TotalCost.Currency,
				CustomerID:   patient.PaymentCustomerId,
				Description:  fmt.Sprintf("Spruce Visit for %s %s", patient.FirstName, patient.LastName),
				CardToken:    card.ThirdPartyID,
				ReceiptEmail: patient.Email,
				Metadata: map[string]string{
					"receipt_ref_num": pReceipt.ReferenceNumber,
				},
			})
			if err != nil {
				w.chargeFailure.Inc(1)
				return err
			}
			w.chargeSuccess.Inc(1)
			defaultCardId := card.ID.Int64()
			patientReceiptUpdate.CreditCardID = &defaultCardId
		}

		patientReceiptUpdate.StripeChargeID = &charge.ID
	}

	if currentStatus == common.PRChargePending {
		// update receipt to indicate that any payment was successfully charged to the customer
		if err := w.dataAPI.UpdatePatientReceipt(pReceipt.ID, patientReceiptUpdate); err != nil {
			return err
		}
		currentStatus = common.PRCharged
	}

	// update the patient visit to indicate that it was successfully charged
	pvStatus := common.PVStatusCharged
	if err := w.dataAPI.UpdatePatientVisit(m.PatientVisitID, &api.PatientVisitUpdate{Status: &pvStatus}); err != nil {
		return err
	}

	// first publish the charged event before sending the email so that we are not waiting too long
	// before routing the case (say, in the event that email service is down)
	w.publishVisitChargedEvent(m)

	return nil
}

func (w *Worker) retrieveOrCreatePatientReceipt(patientID, patientVisitID, itemCostId int64,
	itemType sku.SKU, costBreakdown *common.CostBreakdown) (*common.PatientReceipt, error) {
	// check if a receipt exists in the databse
	var pReceipt *common.PatientReceipt
	var err error
	pReceipt, err = w.dataAPI.GetPatientReceipt(patientID, patientVisitID, itemType, false)
	if err != api.NoRowsError && err != nil {
		return nil, err
	} else if err != api.NoRowsError {
		return pReceipt, nil
	}

	// generate a random receipt number
	refNum, err := common.GenerateRandomNumber(receiptNumberMax, receiptNumDigits)
	if err != nil {
		return nil, err
	}

	// append the itemID to ensure that the number is unique
	refNum += strconv.FormatInt(patientVisitID, 10)

	pReceipt = &common.PatientReceipt{
		ReferenceNumber: refNum,
		ItemType:        itemType,
		ItemID:          patientVisitID,
		PatientID:       patientID,
		Status:          common.PRChargePending,
		CostBreakdown:   costBreakdown,
		ItemCostID:      itemCostId,
	}

	if err := w.dataAPI.CreatePatientReceipt(pReceipt); err != nil {
		return nil, err
	}

	return pReceipt, nil
}

func (w *Worker) publishVisitChargedEvent(m *VisitMessage) error {
	if err := w.dispatcher.Publish(&VisitChargedEvent{
		PatientID:     m.PatientID,
		AccountID:     m.AccountID,
		VisitID:       m.PatientVisitID,
		PatientCaseID: m.PatientCaseID,
	}); err != nil {
		return err
	}
	return nil
}
