package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Payment struct {
	Line                []Line
	CustomerRef         ReferenceType        `json:",omitempty"`
	DepositToAccountRef *ReferenceType       `json:",omitempty"`
	CurrencyRef         ReferenceType        `json:",omitempty"`
	ProjectRef          ReferenceType        `json:",omitempty"`
	PaymentMethodRef    *ReferenceType       `json:",omitempty"`
	TaxExemptionRef     *ReferenceType       `json:",omitempty"`
	TxnDate             Date                 `json:",omitempty"`
	MetaData            ModificationMetaData `json:",omitempty"`
	ExchangeRate        json.Number          `json:",omitempty"`
	UnappliedAmt        json.Number          `json:",omitempty"`
	TotalAmt            json.Number          `json:",omitempty"`
	Id                  string               `json:",omitempty"`
	SyncToken           string               `json:",omitempty"`
	PrivateNote         string               `json:",omitempty"`
	ProcessPayment      bool                 `json:",omitempty"`
	// CreditCardPayment
	// TransactionLocationType
	// PaymentRefNum
}

type CDCPayment struct {
	Payment
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreatePayment creates the given payment within QuickBooks.
func (c *Client) CreatePayment(params RequestParameters, payment *Payment) (*Payment, error) {
	var resp struct {
		Payment Payment
		Time    Date
	}

	if err := c.post(params, "payment", payment, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Payment, nil
}

// DeletePayment deletes the given payment from QuickBooks.
func (c *Client) DeletePayment(params RequestParameters, payment *Payment) error {
	if payment.Id == "" || payment.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "payment", payment, nil, map[string]string{"operation": "delete"})
}

// FindPayments gets the full list of Payments in the QuickBooks account.
func (c *Client) FindPayments(params RequestParameters) ([]Payment, error) {
	var resp struct {
		QueryResponse struct {
			Payments      []Payment `json:"Payment"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Payment", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no payments could be found")
	}

	payments := make([]Payment, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Payment ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Payments == nil {
			return nil, errors.New("no payments could be found")
		}

		payments = append(payments, resp.QueryResponse.Payments...)
	}

	return payments, nil
}

func (c *Client) FindPaymentsByPage(params RequestParameters, startPosition, pageSize int) ([]Payment, error) {
	var resp struct {
		QueryResponse struct {
			Payments      []Payment `json:"Payment"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Payment ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Payments == nil {
		return nil, errors.New("no payments could be found")
	}

	return resp.QueryResponse.Payments, nil
}

// FindPaymentById returns an payment with a given Id.
func (c *Client) FindPaymentById(params RequestParameters, id string) (*Payment, error) {
	var resp struct {
		Payment Payment
		Time    Date
	}

	if err := c.get(params, "payment/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Payment, nil
}

// QueryPayments accepts a SQL query and returns all payments found using it.
func (c *Client) QueryPayments(params RequestParameters, query string) ([]Payment, error) {
	var resp struct {
		QueryResponse struct {
			Payments      []Payment `json:"Payment"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Payments == nil {
		return nil, errors.New("could not find any payments")
	}

	return resp.QueryResponse.Payments, nil
}

// UpdatePayment full updates the payment, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdatePayment(params RequestParameters, payment *Payment) (*Payment, error) {
	if payment.Id == "" {
		return nil, errors.New("missing payment id")
	}

	existingPayment, err := c.FindPaymentById(params, payment.Id)
	if err != nil {
		return nil, err
	}

	payment.SyncToken = existingPayment.SyncToken

	payload := struct {
		*Payment
	}{
		Payment: payment,
	}

	var paymentData struct {
		Payment Payment
		Time    Date
	}

	if err = c.post(params, "payment", payload, &paymentData, nil); err != nil {
		return nil, err
	}

	return &paymentData.Payment, err
}

// VoidPayment voids the given payment in QuickBooks.
func (c *Client) VoidPayment(params RequestParameters, payment Payment) error {
	if payment.Id == "" {
		return errors.New("missing payment id")
	}

	existingPayment, err := c.FindPaymentById(params, payment.Id)
	if err != nil {
		return err
	}

	payment.SyncToken = existingPayment.SyncToken

	return c.post(params, "payment", payment, nil, map[string]string{"operation": "update", "include": "void"})
}
