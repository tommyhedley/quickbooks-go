// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/guregu/null.v4"
)

// Customer represents a QuickBooks Customer object.
type Customer struct {
	CustomerTypeRef      *ReferenceType       `json:",omitempty"`
	ParentRef            *ReferenceType       `json:",omitempty"`
	CurrencyRef          *ReferenceType       `json:",omitempty"`
	DefaultTaxCodeRef    *ReferenceType       `json:",omitempty"`
	SalesTermRef         *ReferenceType       `json:",omitempty"`
	PaymentMethodRef     *ReferenceType       `json:",omitempty"`
	PrimaryPhone         *TelephoneNumber     `json:",omitempty"`
	AlternatePhone       *TelephoneNumber     `json:",omitempty"`
	Mobile               *TelephoneNumber     `json:",omitempty"`
	Fax                  *TelephoneNumber     `json:",omitempty"`
	PrimaryEmailAddr     *EmailAddress        `json:",omitempty"`
	WebAddr              *WebSiteAddress      `json:",omitempty"`
	BillAddr             *PhysicalAddress     `json:",omitempty"`
	ShipAddr             *PhysicalAddress     `json:",omitempty"`
	OpenBalanceDate      *Date                `json:",omitempty"`
	Job                  null.Bool            `json:",omitempty"`
	MetaData             ModificationMetaData `json:",omitempty"`
	Balance              json.Number          `json:",omitempty"`
	BalanceWithJobs      json.Number          `json:",omitempty"`
	Id                   string               `json:",omitempty"`
	SyncToken            string               `json:",omitempty"`
	Title                string               `json:",omitempty"`
	GivenName            string               `json:",omitempty"`
	MiddleName           string               `json:",omitempty"`
	FamilyName           string               `json:",omitempty"`
	Suffix               string               `json:",omitempty"`
	DisplayName          string               `json:",omitempty"`
	FullyQualifiedName   string               `json:",omitempty"`
	CompanyName          string               `json:",omitempty"`
	PrintOnCheckName     string               `json:",omitempty"`
	TaxExemptionReasonId string               `json:",omitempty"`
	Notes                string               `json:",omitempty"`
	ResaleNum            string               `json:",omitempty"`
	Level                int                  `json:",omitempty"`
	Active               bool                 `json:",omitempty"`
	Taxable              bool                 `json:",omitempty"`
	BillWithParent       bool                 `json:",omitempty"`
	Domain               string               `json:"domain,omitempty"`
	Status               string               `json:"status,omitempty"`
	// Source
	// PrimaryTaxIdentifier
	// SecondaryTaxIdentifier
	// ARAccountRef
	// GSTRegistrationType
	// GSTIN
	// BusinessNumber
}

// CreateCustomer creates the given Customer on the QuickBooks server,
// returning the resulting Customer object.
func (c *Client) CreateCustomer(params RequestParameters, customer *Customer) (*Customer, error) {
	var resp struct {
		Customer Customer
		Time     Date
	}

	if err := c.post(params, "customer", customer, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Customer, nil
}

// FindCustomers gets the full list of Customers in the QuickBooks account.
func (c *Client) FindCustomers(params RequestParameters) ([]Customer, error) {
	var resp struct {
		QueryResponse struct {
			Customers     []Customer `json:"Customer"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Customer", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	customers := make([]Customer, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Customer ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		customers = append(customers, resp.QueryResponse.Customers...)
	}

	return customers, nil
}

func (c *Client) FindCustomersByPage(params RequestParameters, startPosition, pageSize int) ([]Customer, error) {
	var resp struct {
		QueryResponse struct {
			Customers     []Customer `json:"Customer"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Customer ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Customers, nil
}

// FindCustomerById returns a customer with a given Id.
func (c *Client) FindCustomerById(params RequestParameters, id string) (*Customer, error) {
	var r struct {
		Customer Customer
		Time     Date
	}

	if err := c.get(params, "customer/"+id, &r, nil); err != nil {
		return nil, err
	}

	return &r.Customer, nil
}

// FindCustomerByName gets a customer with a given name.
func (c *Client) FindCustomerByName(params RequestParameters, name string) (*Customer, error) {
	var resp struct {
		QueryResponse struct {
			Customer   []Customer
			TotalCount int
		}
	}

	query := "SELECT * FROM Customer WHERE DisplayName = '" + strings.Replace(name, "'", "''", -1) + "'"

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return &resp.QueryResponse.Customer[0], nil
}

// QueryCustomers accepts an SQL query and returns all customers found using it
func (c *Client) QueryCustomers(params RequestParameters, query string) ([]Customer, error) {
	var resp struct {
		QueryResponse struct {
			Customers     []Customer `json:"Customer"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Customers, nil
}

// UpdateCustomer full updates the customer, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateCustomer(params RequestParameters, customer *Customer) (*Customer, error) {
	if customer.Id == "" {
		return nil, errors.New("missing customer id")
	}

	existingCustomer, err := c.FindCustomerById(params, customer.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing customer: %v", err)
	}

	customer.SyncToken = existingCustomer.SyncToken

	payload := struct {
		*Customer
	}{
		Customer: customer,
	}

	var customerData struct {
		Customer Customer
		Time     Date
	}

	if err = c.post(params, "customer", payload, &customerData, nil); err != nil {
		return nil, err
	}

	return &customerData.Customer, nil
}

// SparseUpdateCustomer updates only fields included in the customer struct, other fields are left unmodified
func (c *Client) SparseUpdateCustomer(params RequestParameters, customer *Customer) (*Customer, error) {
	if customer.Id == "" {
		return nil, errors.New("missing customer id")
	}

	existingCustomer, err := c.FindCustomerById(params, customer.Id)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing customer: %v", err)
	}

	customer.SyncToken = existingCustomer.SyncToken

	payload := struct {
		*Customer
		Sparse bool `json:"sparse"`
	}{
		Customer: customer,
		Sparse:   true,
	}

	var customerData struct {
		Customer Customer
		Time     Date
	}

	if err = c.post(params, "customer", payload, &customerData, nil); err != nil {
		return nil, err
	}

	return &customerData.Customer, nil
}
