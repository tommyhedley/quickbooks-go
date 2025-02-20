package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Estimate struct {
	Line                  []Line
	LinkedTxn             []LinkedTxn     `json:",omitempty"`
	CustomField           []CustomField   `json:",omitempty"`
	TxnTaxDetail          TxnTaxDetail    `json:",omitempty"`
	CustomerRef           ReferenceType   `json:",omitempty"`
	ClassRef              ReferenceType   `json:",omitempty"`
	SalesTermRef          ReferenceType   `json:",omitempty"`
	DepartmentRef         ReferenceType   `json:",omitempty"`
	ShipMethodRef         ReferenceType   `json:",omitempty"`
	RecurDatRef           ReferenceType   `json:",omitempty"`
	TaxExemptionRef       ReferenceType   `json:",omitempty"`
	ShipFromAddr          PhysicalAddress `json:",omitempty"`
	ShipAddr              PhysicalAddress `json:",omitempty"`
	BillAddr              PhysicalAddress `json:",omitempty"`
	BillEmail             EmailAddress    `json:",omitempty"`
	BillEmailCC           EmailAddress    `json:"BillEmailCc,omitempty"`
	BillEmailBCC          EmailAddress    `json:"BillEmailBcc,omitempty"`
	TxnDate               Date            `json:",omitempty"`
	ShipDate              Date            `json:",omitempty"`
	AcceptedDate          Date            `json:",omitempty"`
	ExpirationDate        Date            `json:",omitempty"`
	DueDate               Date            `json:",omitempty"`
	CustomerMemo          MemoRef         `json:",omitempty"`
	MetaData              MetaData        `json:",omitempty"`
	Id                    string          `json:",omitempty"`
	DocNumber             string          `json:",omitempty"`
	SyncToken             string          `json:",omitempty"`
	TxnStatus             string          `json:",omitempty"`
	PrintStatus           string          `json:",omitempty"`
	EmailStatus           string          `json:",omitempty"`
	PrivateNote           string          `json:",omitempty"`
	AcceptedBy            string          `json:",omitempty"`
	ExchangeRate          json.Number     `json:",omitempty"`
	TotalAmt              json.Number     `json:",omitempty"`
	HomeTotalAmt          json.Number     `json:",omitempty"`
	ApplyTaxAfterDiscount bool            `json:",omitempty"`
	FreeFormAddress       bool            `json:",omitempty"`
	// GlobalTaxCalculation
	// CurrencyRef
	// ProjectRef
}

type CDCEstimate struct {
	Estimate
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateEstimate creates the given Estimate on the QuickBooks server, returning
// the resulting Estimate object.
func (c *Client) CreateEstimate(estimate *Estimate) (*Estimate, error) {
	var resp struct {
		Estimate Estimate
		Time     Date
	}

	if err := c.post("estimate", estimate, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Estimate, nil
}

// DeleteEstimate deletes the estimate
func (c *Client) DeleteEstimate(estimate *Estimate) error {
	if estimate.Id == "" || estimate.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post("estimate", estimate, nil, map[string]string{"operation": "delete"})
}

// FindEstimates gets the full list of Estimates in the QuickBooks account.
func (c *Client) FindEstimates() ([]Estimate, error) {
	var resp struct {
		QueryResponse struct {
			Estimates     []Estimate `json:"Estimate"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM Estimate", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no estimates could be found")
	}

	estimates := make([]Estimate, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Estimate ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Estimates == nil {
			return nil, errors.New("no estimates could be found")
		}

		estimates = append(estimates, resp.QueryResponse.Estimates...)
	}

	return estimates, nil
}

// FindEstimateById finds the estimate by the given id
func (c *Client) FindEstimateById(id string) (*Estimate, error) {
	var resp struct {
		Estimate Estimate
		Time     Date
	}

	if err := c.get("estimate/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Estimate, nil
}

// QueryEstimates accepts an SQL query and returns all estimates found using it
func (c *Client) QueryEstimates(query string) ([]Estimate, error) {
	var resp struct {
		QueryResponse struct {
			Estimates     []Estimate `json:"Estimate"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Estimates == nil {
		return nil, errors.New("could not find any estimates")
	}

	return resp.QueryResponse.Estimates, nil
}

// SendEstimate sends the estimate to the Estimate.BillEmail if emailAddress is left empty
func (c *Client) SendEstimate(estimateId string, emailAddress string) error {
	queryParameters := make(map[string]string)

	if emailAddress != "" {
		queryParameters["sendTo"] = emailAddress
	}

	return c.post("estimate/"+estimateId+"/send", nil, nil, queryParameters)
}

// UpdateEstimate updates the estimate
func (c *Client) UpdateEstimate(estimate *Estimate) (*Estimate, error) {
	if estimate.Id == "" {
		return nil, errors.New("missing estimate id")
	}

	existingEstimate, err := c.FindEstimateById(estimate.Id)
	if err != nil {
		return nil, err
	}

	estimate.SyncToken = existingEstimate.SyncToken

	payload := struct {
		*Estimate
		Sparse bool `json:"sparse"`
	}{
		Estimate: estimate,
		Sparse:   true,
	}

	var estimateData struct {
		Estimate Estimate
		Time     Date
	}

	if err = c.post("estimate", payload, &estimateData, nil); err != nil {
		return nil, err
	}

	return &estimateData.Estimate, err
}

func (c *Client) VoidEstimate(estimate Estimate) error {
	if estimate.Id == "" {
		return errors.New("missing estimate id")
	}

	existingEstimate, err := c.FindEstimateById(estimate.Id)
	if err != nil {
		return err
	}

	estimate.SyncToken = existingEstimate.SyncToken

	return c.post("estimate", estimate, nil, map[string]string{"operation": "void"})
}
