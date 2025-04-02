package quickbooks

import (
	"errors"
	"strconv"
)

type TaxCode struct {
	PurchaseTaxRateList TaxRateList          `json:",omitempty"`
	SalesTaxRateList    TaxRateList          `json:",omitempty"`
	MetaData            ModificationMetaData `json:",omitempty"`
	Id                  string               `json:",omitempty"`
	Name                string               `json:",omitempty"`
	SyncToken           string               `json:",omitempty"`
	Description         string               `json:",omitempty"`
	TaxCodeConfigType   string               `json:",omitempty"`
	TaxGroup            bool                 `json:",omitempty"`
	Taxable             bool                 `json:",omitempty"`
	Active              bool                 `json:",omitempty"`
	Hidden              bool                 `json:",omitempty"`
}

// FindTaxCodes gets the full list of TaxCodes in the QuickBooks account.
func (c *Client) FindTaxCodes(params RequestParameters) ([]TaxCode, error) {
	var resp struct {
		QueryResponse struct {
			TaxCodes      []TaxCode `json:"TaxCode"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM TaxCode", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	taxCodes := make([]TaxCode, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM TaxCode ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		taxCodes = append(taxCodes, resp.QueryResponse.TaxCodes...)
	}

	return taxCodes, nil
}

func (c *Client) FindTaxCodesByPage(params RequestParameters, startPosition, pageSize int) ([]TaxCode, error) {
	var resp struct {
		QueryResponse struct {
			TaxCodes      []TaxCode `json:"TaxCode"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM TaxCode ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.TaxCodes, nil
}

// FindTaxCodeById finds the taxCode by the given id
func (c *Client) FindTaxCodeById(params RequestParameters, id string) (*TaxCode, error) {
	var resp struct {
		TaxCode TaxCode
		Time    Date
	}

	if err := c.get(params, "taxCode/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TaxCode, nil
}

// QueryTaxCodes accepts an SQL query and returns all taxCodes found using it
func (c *Client) QueryTaxCodes(params RequestParameters, query string) ([]TaxCode, error) {
	var resp struct {
		QueryResponse struct {
			TaxCodes      []TaxCode `json:"TaxCode"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.TaxCodes, nil
}
