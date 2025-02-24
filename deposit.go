package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Deposit struct {
	Line                []Line
	TxnTaxDetail        *TxnTaxDetail `json:",omitempty"`
	DepositToAccountRef ReferenceType
	CurrencyRef         ReferenceType        `json:",omitempty"`
	DepartmentRef       *ReferenceType       `json:",omitempty"`
	RecurDataRef        *ReferenceType       `json:",omitempty"`
	TxnDate             *Date                `json:",omitempty"`
	MetaData            ModificationMetaData `json:",omitempty"`
	ExchangeRate        json.Number          `json:",omitempty"`
	TotalAmt            json.Number          `json:",omitempty"`
	HomeTotalAmt        json.Number          `json:",omitempty"`
	Id                  string               `json:",omitempty"`
	SyncToken           string               `json:",omitempty"`
	PrivateNote         string               `json:",omitempty"`
	// GlobalTaxCalculation
	// CashBackInfo
	// TransactionLocationType
}

type CDCDeposit struct {
	Deposit
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// CreateDeposit creates the given deposit within QuickBooks
func (c *Client) CreateDeposit(params RequestParameters, deposit *Deposit) (*Deposit, error) {
	var resp struct {
		Deposit Deposit
		Time    Date
	}

	if err := c.post(params, "deposit", deposit, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Deposit, nil
}

func (c *Client) DeleteDeposit(params RequestParameters, deposit *Deposit) error {
	if deposit.Id == "" || deposit.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "deposit", deposit, nil, map[string]string{"operation": "delete"})
}

// FindDeposits gets the full list of Deposits in the QuickBooks account.
func (c *Client) FindDeposits(params RequestParameters) ([]Deposit, error) {
	var resp struct {
		QueryResponse struct {
			Deposits      []Deposit `json:"Deposit"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Deposit", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no deposits could be found")
	}

	deposits := make([]Deposit, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Deposit ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Deposits == nil {
			return nil, errors.New("no deposits could be found")
		}

		deposits = append(deposits, resp.QueryResponse.Deposits...)
	}

	return deposits, nil
}

// FindDepositById returns an deposit with a given Id.
func (c *Client) FindDepositById(params RequestParameters, id string) (*Deposit, error) {
	var resp struct {
		Deposit Deposit
		Time    Date
	}

	if err := c.get(params, "deposit/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Deposit, nil
}

// QueryDeposits accepts an SQL query and returns all deposits found using it
func (c *Client) QueryDeposits(params RequestParameters, query string) ([]Deposit, error) {
	var resp struct {
		QueryResponse struct {
			Deposits      []Deposit `json:"Deposit"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Deposits == nil {
		return nil, errors.New("could not find any deposits")
	}

	return resp.QueryResponse.Deposits, nil
}

// UpdateDeposit full updates the deposit, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateDeposit(params RequestParameters, deposit *Deposit) (*Deposit, error) {
	if deposit.Id == "" {
		return nil, errors.New("missing deposit id")
	}

	existingDeposit, err := c.FindDepositById(params, deposit.Id)
	if err != nil {
		return nil, err
	}

	deposit.SyncToken = existingDeposit.SyncToken

	payload := struct {
		*Deposit
	}{
		Deposit: deposit,
	}

	var depositData struct {
		Deposit Deposit
		Time    Date
	}

	if err = c.post(params, "deposit", payload, &depositData, nil); err != nil {
		return nil, err
	}

	return &depositData.Deposit, err
}

// SparseUpdateDeposit updates only fields included in the deposit struct, other fields are left unmodified
func (c *Client) SparseUpdateDeposit(params RequestParameters, deposit *Deposit) (*Deposit, error) {
	if deposit.Id == "" {
		return nil, errors.New("missing deposit id")
	}

	existingDeposit, err := c.FindDepositById(params, deposit.Id)
	if err != nil {
		return nil, err
	}

	deposit.SyncToken = existingDeposit.SyncToken

	payload := struct {
		*Deposit
		Sparse bool `json:"sparse"`
	}{
		Deposit: deposit,
		Sparse:  true,
	}

	var depositData struct {
		Deposit Deposit
		Time    Date
	}

	if err = c.post(params, "deposit", payload, &depositData, nil); err != nil {
		return nil, err
	}

	return &depositData.Deposit, err
}
