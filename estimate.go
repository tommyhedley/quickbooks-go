package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Estimate struct {
	Line                  []Line
	LinkedTxn             []LinkedTxn          `json:",omitempty"`
	CustomField           []CustomField        `json:",omitempty"`
	TxnTaxDetail          *TxnTaxDetail        `json:",omitempty"`
	CustomerRef           ReferenceType        `json:",omitempty"`
	ClassRef              *ReferenceType       `json:",omitempty"`
	SalesTermRef          *ReferenceType       `json:",omitempty"`
	DepartmentRef         *ReferenceType       `json:",omitempty"`
	ShipMethodRef         *ReferenceType       `json:",omitempty"`
	RecurDataRef          *ReferenceType       `json:",omitempty"`
	TaxExemptionRef       *ReferenceType       `json:",omitempty"`
	CurrencyRef           ReferenceType        `json:",omitempty"`
	ProjectRef            ReferenceType        `json:",omitempty"`
	ShipFromAddr          PhysicalAddress      `json:",omitempty"`
	ShipAddr              *PhysicalAddress     `json:",omitempty"`
	BillAddr              *PhysicalAddress     `json:",omitempty"`
	BillEmail             EmailAddress         `json:",omitempty"`
	BillEmailCC           *EmailAddress        `json:"BillEmailCc,omitempty"`
	BillEmailBCC          *EmailAddress        `json:"BillEmailBcc,omitempty"`
	DeliveryInfo          *DeliveryInfo        `json:",omitempty"`
	TxnDate               *Date                `json:",omitempty"`
	ShipDate              *Date                `json:",omitempty"`
	AcceptedDate          *Date                `json:",omitempty"`
	ExpirationDate        *Date                `json:",omitempty"`
	DueDate               *Date                `json:",omitempty"`
	CustomerMemo          MemoRef              `json:",omitempty"`
	MetaData              ModificationMetaData `json:",omitempty"`
	ExchangeRate          json.Number          `json:",omitempty"`
	TotalAmt              json.Number          `json:",omitempty"`
	HomeTotalAmt          json.Number          `json:",omitempty"`
	Id                    string               `json:",omitempty"`
	DocNumber             string               `json:",omitempty"`
	SyncToken             string               `json:",omitempty"`
	TxnStatus             string               `json:",omitempty"`
	PrintStatus           string               `json:",omitempty"`
	EmailStatus           string               `json:",omitempty"`
	PrivateNote           string               `json:",omitempty"`
	AcceptedBy            string               `json:",omitempty"`
	ApplyTaxAfterDiscount bool                 `json:",omitempty"`
	FreeFormAddress       bool                 `json:",omitempty"`
	Domain                string               `json:"domain,omitempty"`
	Status                string               `json:"status,omitempty"`
	// GlobalTaxCalculation
	// TransactionLocationType
}

// CreateEstimate creates the given Estimate on the QuickBooks server, returning
// the resulting Estimate object.
func (c *Client) CreateEstimate(params RequestParameters, estimate *Estimate) (*Estimate, error) {
	var resp struct {
		Estimate Estimate
		Time     Date
	}

	if err := c.post(params, "estimate", estimate, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Estimate, nil
}

// DeleteEstimate deletes the estimate
func (c *Client) DeleteEstimate(params RequestParameters, estimate *Estimate) error {
	if estimate.Id == "" || estimate.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "estimate", estimate, nil, map[string]string{"operation": "delete"})
}

// FindEstimates gets the full list of Estimates in the QuickBooks account.
func (c *Client) FindEstimates(params RequestParameters) ([]Estimate, error) {
	var resp struct {
		QueryResponse struct {
			Estimates     []Estimate `json:"Estimate"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Estimate", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	estimates := make([]Estimate, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Estimate ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		estimates = append(estimates, resp.QueryResponse.Estimates...)
	}

	return estimates, nil
}

func (c *Client) FindEstimatesByPage(params RequestParameters, startPosition, pageSize int) ([]Estimate, error) {
	var resp struct {
		QueryResponse struct {
			Estimates     []Estimate `json:"Estimate"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Estimate ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Estimates, nil
}

// FindEstimateById finds the estimate by the given id
func (c *Client) FindEstimateById(params RequestParameters, id string) (*Estimate, error) {
	var resp struct {
		Estimate Estimate
		Time     Date
	}

	if err := c.get(params, "estimate/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Estimate, nil
}

// QueryEstimates accepts an SQL query and returns all estimates found using it
func (c *Client) QueryEstimates(params RequestParameters, query string) ([]Estimate, error) {
	var resp struct {
		QueryResponse struct {
			Estimates     []Estimate `json:"Estimate"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Estimates, nil
}

// SendEstimate sends the estimate to the Estimate.BillEmail if emailAddress is left empty
func (c *Client) SendEstimate(params RequestParameters, estimateId, emailAddress string) error {
	queryParameters := make(map[string]string)

	if emailAddress != "" {
		queryParameters["sendTo"] = emailAddress
	}

	return c.post(params, "estimate/"+estimateId+"/send", nil, nil, queryParameters)
}

// UpdateEstimate full updates the estimate, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateEstimate(params RequestParameters, estimate *Estimate) (*Estimate, error) {
	if estimate.Id == "" {
		return nil, errors.New("missing estimate id")
	}

	existingEstimate, err := c.FindEstimateById(params, estimate.Id)
	if err != nil {
		return nil, err
	}

	estimate.SyncToken = existingEstimate.SyncToken

	payload := struct {
		*Estimate
	}{
		Estimate: estimate,
	}

	var estimateData struct {
		Estimate Estimate
		Time     Date
	}

	if err = c.post(params, "estimate", payload, &estimateData, nil); err != nil {
		return nil, err
	}

	return &estimateData.Estimate, err
}

// SparseUpdateEstimate updates only fields included in the estimate struct, other fields are left unmodified
func (c *Client) SparseUpdateEstimate(params RequestParameters, estimate *Estimate) (*Estimate, error) {
	if estimate.Id == "" {
		return nil, errors.New("missing estimate id")
	}

	existingEstimate, err := c.FindEstimateById(params, estimate.Id)
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

	if err = c.post(params, "estimate", payload, &estimateData, nil); err != nil {
		return nil, err
	}

	return &estimateData.Estimate, err
}

func (c *Client) VoidEstimate(params RequestParameters, estimate Estimate) error {
	if estimate.Id == "" {
		return errors.New("missing estimate id")
	}

	existingEstimate, err := c.FindEstimateById(params, estimate.Id)
	if err != nil {
		return err
	}

	estimate.SyncToken = existingEstimate.SyncToken

	return c.post(params, "estimate", estimate, nil, map[string]string{"operation": "void"})
}
