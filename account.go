package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type AccountTypeEnum string

const (
	BankAccountType                  AccountTypeEnum = "Bank"
	OtherCurrentAssetAccountType     AccountTypeEnum = "Other Current Asset"
	FixedAssetAccountType            AccountTypeEnum = "Fixed Asset"
	OtherAssetAccountType            AccountTypeEnum = "Other Asset"
	AccountsReceivableAccountType    AccountTypeEnum = "Accounts Receivable"
	EquityAccountType                AccountTypeEnum = "Equity"
	ExpenseAccountType               AccountTypeEnum = "Expense"
	OtherExpenseAccountType          AccountTypeEnum = "Other Expense"
	CostOfGoodsSoldAccountType       AccountTypeEnum = "Cost of Goods Sold"
	AccountsPayableAccountType       AccountTypeEnum = "Accounts Payable"
	CreditCardAccountType            AccountTypeEnum = "Credit Card"
	LongTermLiabilityAccountType     AccountTypeEnum = "Long Term Liability"
	OtherCurrentLiabilityAccountType AccountTypeEnum = "Other Current Liability"
	IncomeAccountType                AccountTypeEnum = "Income"
	OtherIncomeAccountType           AccountTypeEnum = "Other Income"
)

type Account struct {
	CurrencyRef                   *ReferenceType       `json:",omitempty"`
	ParentRef                     *ReferenceType       `json:",omitempty"`
	TaxCodeRef                    *ReferenceType       `json:",omitempty"`
	MetaData                      ModificationMetaData `json:",omitempty"`
	CurrentBalanceWithSubAccounts json.Number          `json:",omitempty"`
	CurrentBalance                json.Number          `json:",omitempty"`
	AccountType                   AccountTypeEnum      `json:",omitempty"`
	Id                            string               `json:",omitempty"`
	Name                          string
	SyncToken                     string `json:",omitempty"`
	AcctNum                       string `json:",omitempty"`
	Description                   string `json:",omitempty"`
	Classification                string `json:",omitempty"`
	FullyQualifiedName            string `json:",omitempty"`
	TxnLocationType               string `json:",omitempty"`
	AccountSubType                string `json:",omitempty"`
	Active                        bool   `json:",omitempty"`
	SubAccount                    bool   `json:",omitempty"`
	Domain                        string `json:"domain,omitempty"`
	Status                        string `json:"status,omitempty"`
	// AccountAlias                  string               `json:",omitempty"`
	// TxnLocationType
}

// CreateAccount creates the given account within QuickBooks
func (c *Client) CreateAccount(params RequestParameters, account *Account) (*Account, error) {
	var resp struct {
		Account Account
		Time    Date
	}

	if err := c.post(params, "account", account, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Account, nil
}

// FindAccounts gets the full list of Accounts in the QuickBooks account.
func (c *Client) FindAccounts(params RequestParameters) ([]Account, error) {
	var resp struct {
		QueryResponse struct {
			Accounts      []Account `json:"Account"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Account", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	accounts := make([]Account, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Account ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		accounts = append(accounts, resp.QueryResponse.Accounts...)
	}

	return accounts, nil
}

func (c *Client) FindAccountsByPage(params RequestParameters, startPosition, pageSize int) ([]Account, error) {
	var resp struct {
		QueryResponse struct {
			Accounts      []Account `json:"Account"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Account ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Accounts, nil
}

// FindAccountById returns an account with a given Id.
func (c *Client) FindAccountById(params RequestParameters, id string) (*Account, error) {
	var resp struct {
		Account Account
		Time    Date
	}

	if err := c.get(params, "account/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Account, nil
}

// QueryAccounts accepts an SQL query and returns all accounts found using it
func (c *Client) QueryAccounts(params RequestParameters, query string) ([]Account, error) {
	var resp struct {
		QueryResponse struct {
			Accounts      []Account `json:"Account"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Accounts, nil
}

// UpdateAccount full updates the account, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateAccount(params RequestParameters, account *Account) (*Account, error) {
	if account.Id == "" {
		return nil, errors.New("missing account id")
	}

	existingAccount, err := c.FindAccountById(params, account.Id)
	if err != nil {
		return nil, err
	}

	account.SyncToken = existingAccount.SyncToken

	payload := struct {
		*Account
	}{
		Account: account,
	}

	var accountData struct {
		Account Account
		Time    Date
	}

	if err = c.post(params, "account", payload, &accountData, nil); err != nil {
		return nil, err
	}

	return &accountData.Account, err
}
