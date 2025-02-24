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
func (c *Client) FindTaxCodes(req RequestParameters) ([]TaxCode, error) {
	var resp struct {
		QueryResponse struct {
			TaxCodes      []TaxCode `json:"TaxCode"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(req, "SELECT COUNT(*) FROM TaxCode", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no tax codes could be found")
	}

	taxCodes := make([]TaxCode, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM TaxCode ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(req, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.TaxCodes == nil {
			return nil, errors.New("no tax codes could be found")
		}

		taxCodes = append(taxCodes, resp.QueryResponse.TaxCodes...)
	}

	return taxCodes, nil
}

func (c *Client) FindTaxCodesByPage(req RequestParameters, startPosition, pageSize int) ([]TaxCode, error) {
	var resp struct {
		QueryResponse struct {
			TaxCodes      []TaxCode `json:"TaxCode"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM TaxCode ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(req, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TaxCodes == nil {
		return nil, errors.New("no tax codes could be found")
	}

	return resp.QueryResponse.TaxCodes, nil
}

// FindTaxCodeById finds the taxCode by the given id
func (c *Client) FindTaxCodeById(req RequestParameters, id string) (*TaxCode, error) {
	var resp struct {
		TaxCode TaxCode
		Time    Date
	}

	if err := c.get(req, "taxCode/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TaxCode, nil
}

// QueryTaxCodes accepts an SQL query and returns all taxCodes found using it
func (c *Client) QueryTaxCodes(req RequestParameters, query string) ([]TaxCode, error) {
	var resp struct {
		QueryResponse struct {
			TaxCodes      []TaxCode `json:"TaxCode"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(req, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TaxCodes == nil {
		return nil, errors.New("could not find any tax codes")
	}

	return resp.QueryResponse.TaxCodes, nil
}
