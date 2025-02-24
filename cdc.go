package quickbooks

import (
	"fmt"
	"strings"
	"time"
)

type ChangeDataCapture struct {
	CDCResponse []struct {
		QueryResponse []struct {
			Account         []CDCAccount         `json:",omitempty"`
			Bill            []CDCBill            `json:",omitempty"`
			BillPayment     []CDCBillPayment     `json:",omitempty"`
			Class           []CDCClass           `json:",omitempty"`
			Customer        []CDCCustomer        `json:",omitempty"`
			CustomerType    []CDCCustomerType    `json:",omitempty"`
			Deposit         []Deposit            `json:",omitempty"`
			Employee        []CDCEmployee        `json:",omitempty"`
			Estimate        []CDCEstimate        `json:",omitempty"`
			Invoice         []CDCInvoice         `json:",omitempty"`
			Item            []CDCItem            `json:",omitempty"`
			Payment         []CDCPayment         `json:",omitempty"`
			PaymentMethod   []CDCPaymentMethod   `json:",omitempty"`
			Purchase        []CDCPurchase        `json:",omitempty"`
			ReimburseCharge []CDCReimburseCharge `json:",omitempty"`
			Term            []CDCTerm            `json:",omitempty"`
			Vendor          []CDCVendor          `json:",omitempty"`
			VendorCredit    []CDCVendorCredit    `json:",omitempty"`
			StartPosition   int                  `json:"startPosition"`
			MaxResults      int                  `json:"maxResults"`
			TotalCount      int                  `json:"totalCount,omitempty"`
		} `json:"QueryResponse"`
	} `json:"CDCResponse"`
	Time string `json:"time"`
}

func (c *Client) ChangeDataCapture(params RequestParameters, entities []string, changedSince time.Time) (ChangeDataCapture, error) {
	var res ChangeDataCapture

	queryParams := map[string]string{
		"entities":     strings.Join(entities, ","),
		"changedSince": changedSince.Format(dateFormat),
	}

	err := c.req(params, "GET", "/cdc", nil, &res, queryParams)
	if err != nil {
		return ChangeDataCapture{}, fmt.Errorf("failed to make change data capture request: %w", err)
	}
	return res, nil
}
