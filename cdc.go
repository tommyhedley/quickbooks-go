package quickbooks

import (
	"fmt"
	"strings"
	"time"
)

type ChangeDataCapture struct {
	CDCResponse []struct {
		QueryResponse []struct {
			Account       []CDCAccount       `json:",omitempty"`
			Bill          []CDCBill          `json:",omitempty"`
			Customer      []CDCCustomer      `json:",omitempty"`
			CustomerType  []CDCCustomerType  `json:",omitempty"`
			Employee      []CDCEmployee      `json:",omitempty"`
			Estimate      []CDCEstimate      `json:",omitempty"`
			Invoice       []CDCInvoice       `json:",omitempty"`
			Item          []CDCItem          `json:",omitempty"`
			PaymentMethod []CDCPaymentMethod `json:",omitempty"`
			Purchase      []CDCPurchase      `json:",omitempty"`
			StartPosition int                `json:"startPosition"`
			MaxResults    int                `json:"maxResults"`
			TotalCount    int                `json:"totalCount,omitempty"`
		} `json:"QueryResponse"`
	} `json:"CDCResponse"`
	Time string `json:"time"`
}

func (c *Client) ChangeDataCapture(entities []string, changedSince time.Time) (ChangeDataCapture, error) {
	var res ChangeDataCapture

	queryParams := map[string]string{
		"entities":     strings.Join(entities, ","),
		"changedSince": changedSince.Format(dateFormat),
	}

	err := c.req("GET", "/cdc", nil, &res, queryParams)
	if err != nil {
		return ChangeDataCapture{}, fmt.Errorf("failed to make change data capture request: %w", err)
	}
	return res, nil
}
