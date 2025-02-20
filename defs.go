// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"encoding/json"
	"time"
)

type CustomField struct {
	DefinitionId string `json:"DefinitionId,omitempty"`
	StringValue  string `json:"StringValue,omitempty"`
	Type         string `json:"Type,omitempty"`
	Name         string `json:"Name,omitempty"`
}

// Date represents a Quickbooks date
type Date struct {
	time.Time `json:",omitempty"`
}

// UnmarshalJSON removes time from parsed date
func (d *Date) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' && b[len(b)-1] == '"' {
		b = b[1 : len(b)-1]
	}

	d.Time, err = time.Parse(dateFormat, string(b))
	if err != nil {
		d.Time, err = time.Parse(dayFormat, string(b))
	}

	return err
}

func (d Date) String() string {
	return d.Format(dateFormat)
}

// EmailAddress represents a QuickBooks email address.
type EmailAddress struct {
	Address string `json:",omitempty"`
}

const (
	QueryPageSize = 1000
	dateFormat    = "2006-01-02T15:04:05-07:00"
	dayFormat     = "2006-01-02"
)

// MemoRef represents a QuickBooks MemoRef object.
type MemoRef struct {
	Value string `json:"value,omitempty"`
}

// MetaData is a timestamp of genesis and last change of a Quickbooks object
type MetaData struct {
	CreateTime      Date `json:",omitempty"`
	LastUpdatedTime Date `json:",omitempty"`
}

// PhysicalAddress represents a QuickBooks address.
type PhysicalAddress struct {
	Id string `json:"Id,omitempty"`
	// These lines are context-dependent! Read the QuickBooks API carefully.
	Line1   string `json:",omitempty"`
	Line2   string `json:",omitempty"`
	Line3   string `json:",omitempty"`
	Line4   string `json:",omitempty"`
	Line5   string `json:",omitempty"`
	City    string `json:",omitempty"`
	Country string `json:",omitempty"`
	// A.K.A. State.
	CountrySubDivisionCode string `json:",omitempty"`
	PostalCode             string `json:",omitempty"`
	Lat                    string `json:",omitempty"`
	Long                   string `json:",omitempty"`
}

// ReferenceType represents a QuickBooks reference to another object.
type ReferenceType struct {
	Value string `json:"value,omitempty"`
	Name  string `json:"name,omitempty"`
	Type  string `json:"type,omitempty"`
}

// TelephoneNumber represents a QuickBooks phone number.
type TelephoneNumber struct {
	FreeFormNumber string `json:",omitempty"`
}

// WebSiteAddress represents a Quickbooks Website
type WebSiteAddress struct {
	URI string `json:",omitempty"`
}

type MarkupInfo struct {
	PriceLevelRef          ReferenceType `json:",omitempty"`
	Percent                json.Number   `json:",omitempty"`
	MarkUpIncomeAccountRef ReferenceType `json:",omitempty"`
}

type DeliveryInfo struct {
	DeliveryType string
	DeliveryTime Date
}

type LinkedTxn struct {
	TxnID     string
	TxnType   string
	TxnLineId string `json:",omitempty"`
}

type TxnTaxDetail struct {
	TxnTaxCodeRef ReferenceType `json:",omitempty"`
	TotalTax      json.Number   `json:",omitempty"`
	TaxLine       []Line        `json:",omitempty"`
}

type Line struct {
	Id                            string `json:",omitempty"`
	LineNum                       int    `json:",omitempty"`
	Description                   string `json:",omitempty"`
	Amount                        json.Number
	DetailType                    string
	LinkedTxn                     []LinkedTxn                   `json:",omitempty"`
	AccountBasedExpenseLineDetail AccountBasedExpenseLineDetail `json:",omitempty"`
	ItemBasedExpenseLineDetail    ItemBasedExpenseLineDetail    `json:",omitempty"`
	SalesItemLineDetail           SalesItemLineDetail           `json:",omitempty"`
	GroupLineDetail               GroupLineDetail               `json:",omitempty"`
	DescriptionLineDetail         DescriptionLineDetail         `json:",omitempty"`
	DiscountLineDetail            DiscountLineDetail            `json:",omitempty"`
	SubTotalLineDetail            SubTotalLineDetail            `json:",omitempty"`
	TaxLineDetail                 TaxLineDetail                 `json:",omitempty"`
}

type BillableStatusEnum string

const (
	BillableStatusType      BillableStatusEnum = "Billable"
	NotBillableStatusType   BillableStatusEnum = "NotBillable"
	HasBeenBilledStatusType BillableStatusEnum = "HasBeenBilled"
)

// AccountBasedExpenseLineDetail ...
type AccountBasedExpenseLineDetail struct {
	AccountRef ReferenceType
	TaxAmount  json.Number `json:",omitempty"`
	// TaxInclusiveAmt json.Number              `json:",omitempty"`
	ClassRef       ReferenceType      `json:",omitempty"`
	TaxCodeRef     ReferenceType      `json:",omitempty"`
	MarkupInfo     MarkupInfo         `json:",omitempty"`
	BillableStatus BillableStatusEnum `json:",omitempty"`
	CustomerRef    ReferenceType      `json:",omitempty"`
}

// ItemBasedExpenseLineDetail ...
type ItemBasedExpenseLineDetail struct {
	ItemRef ReferenceType
	// TaxInclusiveAmt json.Number              `json:",omitempty"`
	// PriceLevelRef ReferenceType `json:",omitempty"`
	ClassRef       ReferenceType      `json:",omitempty"`
	TaxCodeRef     ReferenceType      `json:",omitempty"`
	MarkupInfo     MarkupInfo         `json:",omitempty"`
	BillableStatus BillableStatusEnum `json:",omitempty"`
	CustomerRef    ReferenceType      `json:",omitempty"`
	Qty            json.Number
	UnitPrice      json.Number
}

// SalesItemLineDetail ...
type SalesItemLineDetail struct {
	ItemRef         ReferenceType `json:",omitempty"`
	ClassRef        ReferenceType `json:",omitempty"`
	UnitPrice       json.Number   `json:",omitempty"`
	MarkupInfo      MarkupInfo    `json:",omitempty"`
	Qty             json.Number   `json:",omitempty"`
	ItemAccountRef  ReferenceType `json:",omitempty"`
	TaxCodeRef      ReferenceType `json:",omitempty"`
	ServiceDate     Date          `json:",omitempty"`
	TaxInclusiveAmt json.Number   `json:",omitempty"`
	DiscountRate    json.Number   `json:",omitempty"`
	DiscountAmt     json.Number   `json:",omitempty"`
}

// GroupLineDetail ...
type GroupLineDetail struct {
	Quantity     float32       `json:",omitempty"`
	GroupItemRef ReferenceType `json:",omitempty"`
	Line         []Line        `json:",omitempty"`
}

// DescriptionLineDetail ...
type DescriptionLineDetail struct {
	TaxCodeRef  ReferenceType `json:",omitempty"`
	ServiceDate Date          `json:",omitempty"`
}

// DiscountLineDetail ...
type DiscountLineDetail struct {
	PercentBased    bool
	DiscountPercent json.Number `json:",omitempty"`
}

// SubTotalLineDetail ...
type SubTotalLineDetail struct {
	ItemRef ReferenceType `json:",omitempty"`
}

// TaxLineDetail ...
type TaxLineDetail struct {
	PercentBased     bool        `json:",omitempty"`
	NetAmountTaxable json.Number `json:",omitempty"`
	// TaxInclusiveAmount json.Number `json:",omitempty"`
	// OverrideDeltaAmount
	TaxPercent json.Number `json:",omitempty"`
	TaxRateRef ReferenceType
}
