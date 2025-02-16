// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

// CDCInvoice represents an invoice object returned as part of a Change Data Capture response
type CDCInvoice struct {
	Invoice
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// Invoice represents a QuickBooks Invoice object.
type Invoice struct {
	Id            string        `json:"Id,omitempty"`
	SyncToken     string        `json:",omitempty"`
	MetaData      MetaData      `json:",omitempty"`
	CustomField   []CustomField `json:",omitempty"`
	DocNumber     string        `json:",omitempty"`
	TxnDate       Date          `json:",omitempty"`
	DepartmentRef ReferenceType `json:",omitempty"`
	PrivateNote   string        `json:",omitempty"`
	LinkedTxn     []LinkedTxn   `json:"LinkedTxn"`
	Line          []Line
	TxnTaxDetail  TxnTaxDetail `json:",omitempty"`
	CustomerRef   ReferenceType
	CustomerMemo  MemoRef         `json:",omitempty"`
	BillAddr      PhysicalAddress `json:",omitempty"`
	ShipAddr      PhysicalAddress `json:",omitempty"`
	ClassRef      ReferenceType   `json:",omitempty"`
	SalesTermRef  ReferenceType   `json:",omitempty"`
	DueDate       Date            `json:",omitempty"`
	// GlobalTaxCalculation
	ShipMethodRef                ReferenceType `json:",omitempty"`
	ShipDate                     Date          `json:",omitempty"`
	TrackingNum                  string        `json:",omitempty"`
	TotalAmt                     json.Number   `json:",omitempty"`
	CurrencyRef                  ReferenceType `json:",omitempty"`
	ExchangeRate                 json.Number   `json:",omitempty"`
	HomeAmtTotal                 json.Number   `json:",omitempty"`
	HomeBalance                  json.Number   `json:",omitempty"`
	ApplyTaxAfterDiscount        bool          `json:",omitempty"`
	PrintStatus                  string        `json:",omitempty"`
	EmailStatus                  string        `json:",omitempty"`
	BillEmail                    EmailAddress  `json:",omitempty"`
	BillEmailCC                  EmailAddress  `json:"BillEmailCc,omitempty"`
	BillEmailBCC                 EmailAddress  `json:"BillEmailBcc,omitempty"`
	DeliveryInfo                 *DeliveryInfo `json:",omitempty"`
	TaxExemptionRef              ReferenceType `json:",omitempty"`
	Balance                      json.Number   `json:",omitempty"`
	TxnSource                    string        `json:",omitempty"`
	AllowOnlineCreditCardPayment bool          `json:",omitempty"`
	AllowOnlineACHPayment        bool          `json:",omitempty"`
	Deposit                      json.Number   `json:",omitempty"`
	DepositToAccountRef          ReferenceType `json:",omitempty"`
}

// CreateInvoice creates the given Invoice on the QuickBooks server, returning
// the resulting Invoice object.
func (c *Client) CreateInvoice(invoice *Invoice) (*Invoice, error) {
	var resp struct {
		Invoice Invoice
		Time    Date
	}

	if err := c.post("invoice", invoice, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Invoice, nil
}

// DeleteInvoice deletes the invoice
//
// If the invoice was already deleted, QuickBooks returns 400 :(
// The response looks like this:
// {"Fault":{"Error":[{"Message":"Object Not Found","Detail":"Object Not Found : Something you're trying to use has been made inactive. Check the fields with accounts, invoices, items, vendors or employees.","code":"610","element":""}],"type":"ValidationFault"},"time":"2018-03-20T20:15:59.571-07:00"}
//
// This is slightly horrifying and not documented in their API. When this
// happens we just return success; the goal of deleting it has been
// accomplished, just not by us.
func (c *Client) DeleteInvoice(invoice *Invoice) error {
	if invoice.Id == "" || invoice.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post("invoice", invoice, nil, map[string]string{"operation": "delete"})
}

// FindInvoices gets the full list of Invoices in the QuickBooks account.
func (c *Client) FindInvoices() ([]Invoice, error) {
	var resp struct {
		QueryResponse struct {
			Invoices      []Invoice `json:"Invoice"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query("SELECT COUNT(*) FROM Invoice", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no invoices could be found")
	}

	invoices := make([]Invoice, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Invoice ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Invoices == nil {
			return nil, errors.New("no invoices could be found")
		}

		invoices = append(invoices, resp.QueryResponse.Invoices...)
	}

	return invoices, nil
}

// FindInvoiceById finds the invoice by the given id
func (c *Client) FindInvoiceById(id string) (*Invoice, error) {
	var resp struct {
		Invoice Invoice
		Time    Date
	}

	if err := c.get("invoice/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Invoice, nil
}

// FindInvoicesByPage gets a page of invoices from the QuickBooks account at the current max results size.
func (c *Client) FindInvoicesByPage(StartPosition int) ([]Invoice, error) {
	var resp struct {
		QueryResponse struct {
			Invoices      []Invoice `json:"Invoice"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Invoice ORDERBY Id STARTPOSITION " + strconv.Itoa(StartPosition) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Invoices == nil {
		return nil, errors.New("no invoices could be found")
	}

	return resp.QueryResponse.Invoices, nil
}

// QueryInvoices accepts an SQL query and returns all invoices found using it
func (c *Client) QueryInvoices(query string) ([]Invoice, error) {
	var resp struct {
		QueryResponse struct {
			Invoices      []Invoice `json:"Invoice"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Invoices == nil {
		return nil, errors.New("could not find any invoices")
	}

	return resp.QueryResponse.Invoices, nil
}

// SendInvoice sends the invoice to the Invoice.BillEmail if emailAddress is left empty
func (c *Client) SendInvoice(invoiceId string, emailAddress string) error {
	queryParameters := make(map[string]string)

	if emailAddress != "" {
		queryParameters["sendTo"] = emailAddress
	}

	return c.post("invoice/"+invoiceId+"/send", nil, nil, queryParameters)
}

// UpdateInvoice updates the invoice
func (c *Client) UpdateInvoice(invoice *Invoice) (*Invoice, error) {
	if invoice.Id == "" {
		return nil, errors.New("missing invoice id")
	}

	existingInvoice, err := c.FindInvoiceById(invoice.Id)
	if err != nil {
		return nil, err
	}

	invoice.SyncToken = existingInvoice.SyncToken

	payload := struct {
		*Invoice
		Sparse bool `json:"sparse"`
	}{
		Invoice: invoice,
		Sparse:  true,
	}

	var invoiceData struct {
		Invoice Invoice
		Time    Date
	}

	if err = c.post("invoice", payload, &invoiceData, nil); err != nil {
		return nil, err
	}

	return &invoiceData.Invoice, err
}

func (c *Client) VoidInvoice(invoice Invoice) error {
	if invoice.Id == "" {
		return errors.New("missing invoice id")
	}

	existingInvoice, err := c.FindInvoiceById(invoice.Id)
	if err != nil {
		return err
	}

	invoice.SyncToken = existingInvoice.SyncToken

	return c.post("invoice", invoice, nil, map[string]string{"operation": "void"})
}
