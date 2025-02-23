package quickbooks

import (
	"fmt"
	"time"
)

type BatchOperations string

const (
	Create BatchOperations = "create"
	Update BatchOperations = "update"
	Delete BatchOperations = "delete"
)

type BatchOptions string

const Void BatchOptions = "void"

type BatchFault struct {
	Message string
	Code    string `json:"code"`
	Detail  string
	Element string `json:"element"`
}

type BatchItemRequest struct {
	BID         string          `json:"bId"`
	OptionsData BatchOptions    `json:"optionsData,omitempty"`
	Operation   BatchOperations `json:"operation,omitempty"`
	Query       string          `json:",omitempty"`
}

type BatchFaultResponse struct {
	FaultType string       `json:"type"`
	Faults    []BatchFault `json:"Error"`
}

type BatchItemResponse struct {
	BID             string             `json:"bId"`
	Account         Account            `json:",omitempty"`
	Bill            Bill               `json:",omitempty"`
	BillPayment     BillPayment        `json:",omitempty"`
	Class           Class              `json:",omitempty"`
	CreditMemo      CreditMemo         `json:",omitempty"`
	Customer        Customer           `json:",omitempty"`
	CustomerType    CustomerType       `json:",omitempty"`
	Deposit         Deposit            `json:",omitempty"`
	Employee        Employee           `json:",omitempty"`
	Estimate        Estimate           `json:",omitempty"`
	Invoice         Invoice            `json:",omitempty"`
	Item            Item               `json:",omitempty"`
	Payment         Payment            `json:",omitempty"`
	PaymentMethod   PaymentMethod      `json:",omitempty"`
	Purchase        Purchase           `json:",omitempty"`
	ReimburseCharge ReimburseCharge    `json:",omitempty"`
	TaxCode         TaxCode            `json:",omitempty"`
	TaxRate         TaxRate            `json:",omitempty"`
	Term            Term               `json:",omitempty"`
	TimeActivity    TimeActivity       `json:",omitempty"`
	Vendor          Vendor             `json:",omitempty"`
	VendorCredit    VendorCredit       `json:",omitempty"`
	Fault           BatchFaultResponse `json:",omitempty"`
	QueryResponse   struct {
		Account         []Account         `json:",omitempty"`
		Bill            []Bill            `json:",omitempty"`
		BillPayment     []BillPayment     `json:",omitempty"`
		Class           []Class           `json:",omitempty"`
		CreditMemo      []CreditMemo      `json:",omitempty"`
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
		TaxCode         []TaxCode         `json:",omitempty"`
		TaxRate         []TaxRate         `json:",omitempty"`
		Term            []Term            `json:",omitempty"`
		TimeActivity    []TimeActivity    `json:",omitempty"`
		Vendor          []Vendor          `json:",omitempty"`
		VendorCredit    []VendorCredit    `json:",omitempty"`
		StartPosition   int               `json:"startPosition"`
		MaxResults      int               `json:"maxResults"`
		TotalCount      int               `json:"totalCount,omitempty"`
	} `json:"QueryResponse,omitempty"`
}

func (c *Client) BatchRequest(items []BatchItemRequest) ([]BatchItemResponse, error) {
	if len(items) == 0 {
		return nil, nil
	}

	var allResponses []BatchItemResponse

	// each BatchRequest is limited to 30 items
	chunkSize := 30
	for start := 0; start < len(items); start += chunkSize {
		end := start + chunkSize
		if end > len(items) {
			end = len(items)
		}
		batch := items[start:end]

		var req struct {
			BatchItemRequest []BatchItemRequest `json:"BatchItemRequest"`
		}

		var res struct {
			BatchItemResponses []BatchItemResponse `json:"BatchItemResponse"`
			Time               time.Time           `json:"time"`
		}

		req.BatchItemRequest = batch

		err := c.req("POST", "/batch", req, &res, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to complete batch request: %w", err)
		}

		allResponses = append(allResponses, res.BatchItemResponses...)
	}

	return allResponses, nil
}
