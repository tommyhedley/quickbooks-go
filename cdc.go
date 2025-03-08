package quickbooks

import (
	"fmt"
	"strings"
	"time"
)

type ChangeDataCapture struct {
	CDCResponse []struct {
		QueryResponse []struct {
			Account         []Account         `json:",omitempty"`
			Bill            []Bill            `json:",omitempty"`
			BillPayment     []BillPayment     `json:",omitempty"`
			Class           []Class           `json:",omitempty"`
			Customer        []Customer        `json:",omitempty"`
			CustomerType    []CustomerType    `json:",omitempty"`
			Deposit         []Deposit         `json:",omitempty"`
			Employee        []Employee        `json:",omitempty"`
			Estimate        []Estimate        `json:",omitempty"`
			Invoice         []Invoice         `json:",omitempty"`
			Item            []Item            `json:",omitempty"`
			Payment         []Payment         `json:",omitempty"`
			PaymentMethod   []PaymentMethod   `json:",omitempty"`
			Purchase        []Purchase        `json:",omitempty"`
			ReimburseCharge []ReimburseCharge `json:",omitempty"`
			Term            []Term            `json:",omitempty"`
			Vendor          []Vendor          `json:",omitempty"`
			VendorCredit    []VendorCredit    `json:",omitempty"`
			StartPosition   int               `json:"startPosition"`
			MaxResults      int               `json:"maxResults"`
			TotalCount      int               `json:"totalCount,omitempty"`
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
