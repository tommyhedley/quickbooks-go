package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Purchase struct {
	Line             []Line
	LinkedTxn        []LinkedTxn          `json:",omitempty"`
	TxnTaxDetail     *TxnTaxDetail        `json:",omitempty"`
	AccountRef       ReferenceType        `json:",omitempty"`
	CurrencyRef      ReferenceType        `json:",omitempty"`
	PaymentMethodRef *ReferenceType       `json:",omitempty"`
	DepartmentRef    *ReferenceType       `json:",omitempty"`
	EntityRef        *ReferenceType       `json:",omitempty"`
	RecurDataRef     *ReferenceType       `json:",omitempty"`
	RemitToAddr      *PhysicalAddress     `json:",omitempty"`
	TxnDate          *Date                `json:",omitempty"`
	MetaData         ModificationMetaData `json:",omitempty"`
	ExchangeRate     json.Number          `json:",omitempty"`
	TotalAmt         json.Number          `json:",omitempty"`
	Id               string               `json:",omitempty"`
	DocNumber        string               `json:",omitempty"`
	PrivateNote      string               `json:",omitempty"`
	SyncToken        string               `json:",omitempty"`
	PaymentType      string               `json:",omitempty"`
	PrintStatus      PrintStatusEnum      `json:",omitempty"`
	Credit           bool                 `json:",omitempty"`
	Domain           string               `json:"domain,omitempty"`
	Status           string               `json:"status,omitempty"`
	// GlobalTaxCalculation
	// TransactionLocationType
	// IncludeInAnnualTPAR
}

// CreatePurchase creates the given Purchase on the QuickBooks server, returning
// the resulting Purchase object.
func (c *Client) CreatePurchase(params RequestParameters, purchase *Purchase) (*Purchase, error) {
	var resp struct {
		Purchase Purchase
		Time     Date
	}

	if err := c.post(params, "purchase", purchase, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Purchase, nil
}

// DeletePurchase deletes the purchase
func (c *Client) DeletePurchase(params RequestParameters, purchase *Purchase) error {
	if purchase.Id == "" || purchase.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "purchase", purchase, nil, map[string]string{"operation": "delete"})
}

// FindPurchases gets the full list of Purchases in the QuickBooks account.
func (c *Client) FindPurchases(params RequestParameters) ([]Purchase, error) {
	var resp struct {
		QueryResponse struct {
			Purchases     []Purchase `json:"Purchase"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Purchase", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no purchases could be found")
	}

	purchases := make([]Purchase, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Purchase ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Purchases == nil {
			return nil, errors.New("no purchases could be found")
		}

		purchases = append(purchases, resp.QueryResponse.Purchases...)
	}

	return purchases, nil
}

func (c *Client) FindPurchasesByPage(params RequestParameters, startPosition, pageSize int) ([]Purchase, error) {
	var resp struct {
		QueryResponse struct {
			Purchases     []Purchase `json:"Purchase"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Purchase ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Purchases == nil {
		return nil, errors.New("no purchases could be found")
	}

	return resp.QueryResponse.Purchases, nil
}

// FindPurchaseById finds the purchase by the given id
func (c *Client) FindPurchaseById(params RequestParameters, id string) (*Purchase, error) {
	var resp struct {
		Purchase Purchase
		Time     Date
	}

	if err := c.get(params, "purchase/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Purchase, nil
}

// QueryPurchases accepts an SQL query and returns all purchases found using it
func (c *Client) QueryPurchases(params RequestParameters, query string) ([]Purchase, error) {
	var resp struct {
		QueryResponse struct {
			Purchases     []Purchase `json:"Purchase"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Purchases == nil {
		return nil, errors.New("could not find any purchases")
	}

	return resp.QueryResponse.Purchases, nil
}

// UpdatePurchase full updates the purchase, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdatePurchase(params RequestParameters, purchase *Purchase) (*Purchase, error) {
	if purchase.Id == "" {
		return nil, errors.New("missing purchase id")
	}

	existingPurchase, err := c.FindPurchaseById(params, purchase.Id)
	if err != nil {
		return nil, err
	}

	purchase.SyncToken = existingPurchase.SyncToken

	payload := struct {
		*Purchase
	}{
		Purchase: purchase,
	}

	var purchaseData struct {
		Purchase Purchase
		Time     Date
	}

	if err = c.post(params, "purchase", payload, &purchaseData, nil); err != nil {
		return nil, err
	}

	return &purchaseData.Purchase, err
}
