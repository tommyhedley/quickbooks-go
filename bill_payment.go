package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type BillPaymentTypeEnum string

const (
	CheckPaymentType      BillPaymentTypeEnum = "Check"
	CreditCardPaymentType BillPaymentTypeEnum = "CreditCard"
)

type PrintStatusEnum string

const (
	NotSetStatus        PrintStatusEnum = "NotSet"
	NeedToPrintStatus   PrintStatusEnum = "NeedToPrint"
	PrintCompleteStatus PrintStatusEnum = "PrintComplete"
)

type BillPaymentCheck struct {
	BankAccountRef ReferenceType   `json:",omitempty"`
	PrintStatus    PrintStatusEnum `json:",omitempty"`
}

type BillPaymentCreditCard struct {
	CCAccountRef ReferenceType `json:",omitempty"`
}

type BillPayment struct {
	Line               []Line
	LinkedTxn          []LinkedTxn           `json:",omitempty"`
	VendorRef          ReferenceType         `json:",omitempty"`
	CurrencyRef        ReferenceType         `json:",omitempty"`
	APAccountRef       ReferenceType         `json:",omitempty"`
	DepartmentRef      ReferenceType         `json:",omitempty"`
	CheckPayment       BillPaymentCheck      `json:",omitempty"`
	CreditCardPayment  BillPaymentCreditCard `json:",omitempty"`
	TxnDate            Date                  `json:",omitempty"`
	MetaData           MetaData              `json:",omitempty"`
	TotalAmt           json.Number           `json:",omitempty"`
	ExchangeRate       json.Number           `json:",omitempty"`
	PayType            BillPaymentTypeEnum   `json:",omitempty"`
	Id                 string                `json:",omitempty"`
	SyncToken          string                `json:",omitempty"`
	DocNumber          string                `json:",omitempty"`
	PrivateNote        string                `json:",omitempty"`
	ProcessBillPayment bool                  `json:",omitempty"`
}

type CDCBillPayment struct {
	BillPayment
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateBillPayment creates the given Bill on the QuickBooks server, returning
// the resulting Bill object.
func (c *Client) CreateBillPayment(billPayment *BillPayment) (*BillPayment, error) {
	var resp struct {
		BillPayment BillPayment
		Time        Date
	}

	if err := c.post("billpayment", billPayment, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.BillPayment, nil
}

// DeleteBill deletes the bill
func (c *Client) DeleteBillPayment(billPayment *BillPayment) error {
	if billPayment.Id == "" || billPayment.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post("billpayment", billPayment, nil, map[string]string{"operation": "delete"})
}

// FindBills gets the full list of Bills in the QuickBooks account.
func (c *Client) FindBillPayments() ([]BillPayment, error) {
	var resp struct {
		QueryResponse struct {
			BillPayments  []BillPayment `json:"BillPayment"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM BillPayments", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no bill payments could be found")
	}

	billPayments := make([]BillPayment, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM BillPayment ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.BillPayments == nil {
			return nil, errors.New("no bill payments could be found")
		}

		billPayments = append(billPayments, resp.QueryResponse.BillPayments...)
	}

	return billPayments, nil
}

func (c *Client) FindBillPaymentsByPage(startPosition int, pageSize int) ([]BillPayment, error) {
	var resp struct {
		QueryResponse struct {
			BillPayments  []BillPayment `json:"BillPayment"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM BillPayment ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.BillPayments == nil {
		return nil, errors.New("no bill payments could be found")
	}

	return resp.QueryResponse.BillPayments, nil
}

// FindBillById finds the bill by the given id
func (c *Client) FindBillPaymentById(id string) (*BillPayment, error) {
	var resp struct {
		BillPayment BillPayment
		Time        Date
	}

	if err := c.get("billpayment/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.BillPayment, nil
}

// QueryBills accepts an SQL query and returns all bills found using it
func (c *Client) QueryBillPayments(query string) ([]BillPayment, error) {
	var resp struct {
		QueryResponse struct {
			BillPayments  []BillPayment `json:"BillPayment"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.BillPayments == nil {
		return nil, errors.New("could not find any bill payments")
	}

	return resp.QueryResponse.BillPayments, nil
}

// UpdateBill updates the bill
func (c *Client) UpdateBillPayment(billPayment *BillPayment) (*BillPayment, error) {
	if billPayment.Id == "" {
		return nil, errors.New("missing bill payment id")
	}

	existingBillPayment, err := c.FindBillPaymentById(billPayment.Id)
	if err != nil {
		return nil, err
	}

	billPayment.SyncToken = existingBillPayment.SyncToken

	payload := struct {
		*BillPayment
		Sparse bool `json:"sparse"`
	}{
		BillPayment: billPayment,
		Sparse:      true,
	}

	var billPaymentData struct {
		BillPayment BillPayment
		Time        Date
	}

	if err = c.post("billpayment", payload, &billPaymentData, nil); err != nil {
		return nil, err
	}

	return &billPaymentData.BillPayment, err
}

func (c *Client) VoidBillPayment(billPayment BillPayment) error {
	if billPayment.Id == "" {
		return errors.New("missing bill payment id")
	}

	existingBillPayment, err := c.FindBillPaymentById(billPayment.Id)
	if err != nil {
		return err
	}

	billPayment.SyncToken = existingBillPayment.SyncToken

	return c.post("billpayment", billPayment, nil, map[string]string{"operation": "void"})
}
