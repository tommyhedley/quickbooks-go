package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Bill struct {
	Line                    []Line
	LinkedTxn               []LinkedTxn `json:",omitempty"`
	VendorRef               ReferenceType
	CurrencyRef             ReferenceType        `json:",omitempty"`
	APAccountRef            *ReferenceType       `json:",omitempty"`
	SalesTermRef            *ReferenceType       `json:",omitempty"`
	DepartmentRef           *ReferenceType       `json:",omitempty"`
	RecurDataRef            *ReferenceType       `json:",omitempty"`
	TxnTaxDetail            *TxnTaxDetail        `json:",omitempty"`
	MetaData                ModificationMetaData `json:",omitempty"`
	TxnDate                 Date                 `json:",omitempty"`
	DueDate                 Date                 `json:",omitempty"`
	TotalAmt                json.Number          `json:",omitempty"`
	ExchangeRate            json.Number          `json:",omitempty"`
	HomeBalance             json.Number          `json:",omitempty"`
	Balance                 json.Number          `json:",omitempty"`
	Id                      string               `json:",omitempty"`
	SyncToken               string               `json:",omitempty"`
	TransactionLocationType string               `json:",omitempty"`
	DocNumber               string               `json:",omitempty"`
	PrivateNote             string               `json:",omitempty"`
	// IncludeInAnnualTPAR  bool          `json:",omitempty"`
	// GlobalTaxCalculation
	// TransactionLocationType
}

type CDCBill struct {
	Bill
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateBill creates the given Bill on the QuickBooks server, returning
// the resulting Bill object.
func (c *Client) CreateBill(params RequestParameters, bill *Bill) (*Bill, error) {
	var resp struct {
		Bill Bill
		Time Date
	}

	if err := c.post(params, "bill", bill, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Bill, nil
}

// DeleteBill deletes the bill
func (c *Client) DeleteBill(params RequestParameters, bill *Bill) error {
	if bill.Id == "" || bill.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "bill", bill, nil, map[string]string{"operation": "delete"})
}

// FindBills gets the full list of Bills in the QuickBooks account.
func (c *Client) FindBills(params RequestParameters) ([]Bill, error) {
	var resp struct {
		QueryResponse struct {
			Bills         []Bill `json:"Bill"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Bill", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no bills could be found")
	}

	bills := make([]Bill, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Bill ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Bills == nil {
			return nil, errors.New("no bills could be found")
		}

		bills = append(bills, resp.QueryResponse.Bills...)
	}

	return bills, nil
}

func (c *Client) FindBillsByPage(params RequestParameters, startPosition, pageSize int) ([]Bill, error) {
	var resp struct {
		QueryResponse struct {
			Bills         []Bill `json:"Bill"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Bill ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Bills == nil {
		return nil, errors.New("no bills could be found")
	}

	return resp.QueryResponse.Bills, nil
}

// FindBillById finds the bill by the given id
func (c *Client) FindBillById(params RequestParameters, id string) (*Bill, error) {
	var resp struct {
		Bill Bill
		Time Date
	}

	if err := c.get(params, "bill/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Bill, nil
}

// QueryBills accepts an SQL query and returns all bills found using it
func (c *Client) QueryBills(params RequestParameters, query string) ([]Bill, error) {
	var resp struct {
		QueryResponse struct {
			Bills         []Bill `json:"Bill"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Bills == nil {
		return nil, errors.New("could not find any bills")
	}

	return resp.QueryResponse.Bills, nil
}

// UpdateBill full updates the bill, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateBill(params RequestParameters, bill *Bill) (*Bill, error) {
	if bill.Id == "" {
		return nil, errors.New("missing bill id")
	}

	existingBill, err := c.FindBillById(params, bill.Id)
	if err != nil {
		return nil, err
	}

	bill.SyncToken = existingBill.SyncToken

	payload := struct {
		*Bill
	}{
		Bill: bill,
	}

	var billData struct {
		Bill Bill
		Time Date
	}

	if err = c.post(params, "bill", payload, &billData, nil); err != nil {
		return nil, err
	}

	return &billData.Bill, err
}
