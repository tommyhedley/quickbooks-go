package quickbooks

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type CDCQueryResponse struct {
	Account         []Account         `json:",omitempty"`
	Attachable      []Attachable      `json:",omitempty"`
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
}

type ChangeDataCapture struct {
	CDCResponse []struct {
		QueryResponse []CDCQueryResponse `json:"QueryResponse"`
	} `json:"CDCResponse"`
	Time string `json:"time"`
}

func (c *Client) ChangeDataCapture(ctx context.Context, params RequestParameters, entities []string, changedSince time.Time) (ChangeDataCapture, error) {
	var res ChangeDataCapture

	queryParams := map[string]string{
		"entities":     strings.Join(entities, ","),
		"changedSince": changedSince.Format(dateFormat),
	}

	err := c.req(ctx, params, "GET", "cdc", nil, &res, queryParams)
	if err != nil {
		return ChangeDataCapture{}, fmt.Errorf("failed to make change data capture request: %w", err)
	}
	return res, nil
}

func CDCQueryExtractor[T any](
	res *ChangeDataCapture,
	getSlice func(q CDCQueryResponse) []T,
) []T {
	for _, resp := range res.CDCResponse {
		for _, qr := range resp.QueryResponse {
			if items := getSlice(qr); len(items) > 0 {
				return items
			}
		}
	}
	return nil
}
