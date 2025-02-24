package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type VendorCredit struct {
	Line          []Line
	LinkedTxn     []LinkedTxn `json:",omitempty"`
	VendorRef     ReferenceType
	CurrencyRef   ReferenceType        `json:",omitempty"`
	APAccountRef  *ReferenceType       `json:",omitempty"`
	DepartmentRef *ReferenceType       `json:",omitempty"`
	RecurDataRef  *ReferenceType       `json:",omitempty"`
	TxnDate       *Date                `json:",omitempty"`
	MetaData      ModificationMetaData `json:",omitempty"`
	TotalAmt      json.Number          `json:",omitempty"`
	Balance       json.Number          `json:",omitempty"`
	ExchangeRate  json.Number          `json:",omitempty"`
	Id            string               `json:",omitempty"`
	SyncToken     string               `json:",omitempty"`
	DocNumber     string               `json:",omitempty"`
	PrivateNote   string               `json:",omitempty"`
	// ClobalTaxCalculation
}

type CDCVendorCredit struct {
	VendorCredit
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateVendorCredit creates the given VendorCredit on the QuickBooks server, returning
// the resulting VendorCredit object.
func (c *Client) CreateVendorCredit(params RequestParameters, vendorCredit *VendorCredit) (*VendorCredit, error) {
	var resp struct {
		VendorCredit VendorCredit
		Time         Date
	}

	if err := c.post(params, "vendorcredit", vendorCredit, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.VendorCredit, nil
}

// DeleteVendorCredit deletes the vendorCredit
func (c *Client) DeleteVendorCredit(params RequestParameters, vendorCredit *VendorCredit) error {
	if vendorCredit.Id == "" || vendorCredit.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "vendorcredit", vendorCredit, nil, map[string]string{"operation": "delete"})
}

// FindVendorCredits gets the full list of VendorCredits in the QuickBooks account.
func (c *Client) FindVendorCredits(params RequestParameters) ([]VendorCredit, error) {
	var resp struct {
		QueryResponse struct {
			VendorCredits []VendorCredit `json:"VendorCredit"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM VendorCredit", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no vendor credits could be found")
	}

	vendorCredits := make([]VendorCredit, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM VendorCredit ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.VendorCredits == nil {
			return nil, errors.New("no vendor credits could be found")
		}

		vendorCredits = append(vendorCredits, resp.QueryResponse.VendorCredits...)
	}

	return vendorCredits, nil
}

func (c *Client) FindVendorCreditsByPage(params RequestParameters, startPosition, pageSize int) ([]VendorCredit, error) {
	var resp struct {
		QueryResponse struct {
			VendorCredits []VendorCredit `json:"VendorCredit"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM VendorCredit ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.VendorCredits == nil {
		return nil, errors.New("no vendor credits could be found")
	}

	return resp.QueryResponse.VendorCredits, nil
}

// FindVendorCreditById finds the vendorCredit by the given id
func (c *Client) FindVendorCreditById(params RequestParameters, id string) (*VendorCredit, error) {
	var resp struct {
		VendorCredit VendorCredit
		Time         Date
	}

	if err := c.get(params, "vendorcredit/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.VendorCredit, nil
}

// QueryVendorCredits accepts an SQL query and returns all vendorCredits found using it
func (c *Client) QueryVendorCredits(params RequestParameters, query string) ([]VendorCredit, error) {
	var resp struct {
		QueryResponse struct {
			VendorCredits []VendorCredit `json:"VendorCredit"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.VendorCredits == nil {
		return nil, errors.New("could not find any vendor credits")
	}

	return resp.QueryResponse.VendorCredits, nil
}

// UpdateVendorCredit full updates the vendor credit, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateVendorCredit(params RequestParameters, vendorCredit *VendorCredit) (*VendorCredit, error) {
	if vendorCredit.Id == "" {
		return nil, errors.New("missing vendorCredit id")
	}

	existingVendorCredit, err := c.FindVendorCreditById(params, vendorCredit.Id)
	if err != nil {
		return nil, err
	}

	vendorCredit.SyncToken = existingVendorCredit.SyncToken

	payload := struct {
		*VendorCredit
	}{
		VendorCredit: vendorCredit,
	}

	var vendorCreditData struct {
		VendorCredit VendorCredit
		Time         Date
	}

	if err = c.post(params, "vendorcredit", payload, &vendorCreditData, nil); err != nil {
		return nil, err
	}

	return &vendorCreditData.VendorCredit, err
}
