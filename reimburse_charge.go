package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type ReimburseCharge struct {
	Line            []Line
	LinkedTxn       []LinkedTxn          `json:",omitempty"`
	CustomerRef     ReferenceType        `json:",omitempty"`
	CurrencyRef     ReferenceType        `json:",omitempty"`
	MetaData        ModificationMetaData `json:",omitempty"`
	Amount          json.Number          `json:",omitempty"`
	ExchangeRate    json.Number          `json:",omitempty"`
	HomeTotalAmt    json.Number          `json:",omitempty"`
	Id              string               `json:",omitempty"`
	SyncToken       string               `json:",omitempty"`
	PrivateNote     string               `json:",omitempty"`
	HasBeenInvoiced bool                 `json:",omitempty"`
	Domain          string               `json:"domain,omitempty"`
	Status          string               `json:"status,omitempty"`
}

// FindReimburseCharges gets the full list of ReimburseCharges in the QuickBooks account.
func (c *Client) FindReimburseCharges(params RequestParameters) ([]ReimburseCharge, error) {
	var resp struct {
		QueryResponse struct {
			ReimburseCharges []ReimburseCharge `json:"ReimburseCharge"`
			MaxResults       int
			StartPosition    int
			TotalCount       int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM ReimburseCharge", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no reimburse charges could be found")
	}

	reimburseCharges := make([]ReimburseCharge, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM ReimburseCharge ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.ReimburseCharges == nil {
			return nil, errors.New("no reimburse charges could be found")
		}

		reimburseCharges = append(reimburseCharges, resp.QueryResponse.ReimburseCharges...)
	}

	return reimburseCharges, nil
}

func (c *Client) FindReimburseChargesByPage(params RequestParameters, startPosition, pageSize int) ([]ReimburseCharge, error) {
	var resp struct {
		QueryResponse struct {
			ReimburseCharges []ReimburseCharge `json:"ReimburseCharge"`
			MaxResults       int
			StartPosition    int
			TotalCount       int
		}
	}

	query := "SELECT * FROM ReimburseCharge ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.ReimburseCharges == nil {
		return nil, errors.New("no reimburse charges could be found")
	}

	return resp.QueryResponse.ReimburseCharges, nil
}

// FindReimburseChargeById finds the reimburseCharge by the given id
func (c *Client) FindReimburseChargeById(params RequestParameters, id string) (*ReimburseCharge, error) {
	var resp struct {
		ReimburseCharge ReimburseCharge
		Time            Date
	}

	if err := c.get(params, "reimburseCharge/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.ReimburseCharge, nil
}

// QueryReimburseCharges accepts an SQL query and returns all reimburseCharges found using it
func (c *Client) QueryReimburseCharges(params RequestParameters, query string) ([]ReimburseCharge, error) {
	var resp struct {
		QueryResponse struct {
			ReimburseCharges []ReimburseCharge `json:"ReimburseCharge"`
			StartPosition    int
			MaxResults       int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.ReimburseCharges == nil {
		return nil, errors.New("could not find any reimburse charges")
	}

	return resp.QueryResponse.ReimburseCharges, nil
}
