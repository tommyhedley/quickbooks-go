package quickbooks

import (
	"encoding/json"
	"strconv"
)

type TaxRate struct {
	// EffectiveTaxRate EffectiveTaxRateData `json:",omitempty"`
	// AgencyRef        ReferenceType        `json:",omitempty"`
	// TaxReturnLineRef ReferenceType        `json:",omitempty"`
	MetaData       ModificationMetaData `json:",omitempty"`
	RateValue      json.Number          `json:",omitempty"`
	Id             string               `json:",omitempty"`
	SyncToken      string               `json:",omitempty"`
	Name           string               `json:",omitempty"`
	Description    string               `json:",omitempty"`
	SpecialTaxType string               `json:",omitempty"`
	DisplayType    string               `json:",omitempty"`
	Active         bool                 `json:",omitempty"`
}

// FindTaxRates gets the full list of TaxRates in the QuickBooks account.
func (c *Client) FindTaxRates(params RequestParameters) ([]TaxRate, error) {
	var resp struct {
		QueryResponse struct {
			TaxRates      []TaxRate `json:"TaxRate"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM TaxRate", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	taxRates := make([]TaxRate, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM TaxRate ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		taxRates = append(taxRates, resp.QueryResponse.TaxRates...)
	}

	return taxRates, nil
}

func (c *Client) FindTaxRatesByPage(params RequestParameters, startPosition, pageSize int) ([]TaxRate, error) {
	var resp struct {
		QueryResponse struct {
			TaxRates      []TaxRate `json:"TaxRate"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM TaxRate ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.TaxRates, nil
}

// FindTaxRateById finds the taxRate by the given id
func (c *Client) FindTaxRateById(params RequestParameters, id string) (*TaxRate, error) {
	var resp struct {
		TaxRate TaxRate
		Time    Date
	}

	if err := c.get(params, "taxRate/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TaxRate, nil
}

// QueryTaxRates accepts an SQL query and returns all taxRates found using it
func (c *Client) QueryTaxRates(params RequestParameters, query string) ([]TaxRate, error) {
	var resp struct {
		QueryResponse struct {
			TaxRates      []TaxRate `json:"TaxRate"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.TaxRates, nil
}
