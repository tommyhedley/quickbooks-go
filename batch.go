package quickbooks

import (
	"fmt"
	"reflect"
	"strings"
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

type BatchError struct {
	Faults []BatchFault
}

func (e BatchError) Error() string {
	msgs := make([]string, len(e.Faults))
	for i, f := range e.Faults {
		// include code, element, and message
		msgs[i] = fmt.Sprintf("%s/%s: %s", f.Code, f.Element, f.Message)
	}
	return "batch faults: " + strings.Join(msgs, "; ")
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

type BatchQueryResponse struct {
	Account         []Account         `json:",omitempty"`
	Attachable      []Attachable      `json:",omitempty"`
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
}

type BatchItemResponse struct {
	BID             string             `json:"bId"`
	Account         Account            `json:",omitempty"`
	Attachable      Attachable         `json:",omitempty"`
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
	QueryResponse   BatchQueryResponse `json:"QueryResponse,omitempty"`
}

func (c *Client) BatchRequest(params RequestParameters, batchRequests []BatchItemRequest) ([]BatchItemResponse, error) {
	if len(batchRequests) == 0 {
		return nil, nil
	}

	var allResponses []BatchItemResponse

	// each BatchRequest is limited to 30 items
	chunkSize := 30
	for start := 0; start < len(batchRequests); start += chunkSize {
		end := start + chunkSize
		if end > len(batchRequests) {
			end = len(batchRequests)
		}
		batch := batchRequests[start:end]

		var payload struct {
			BatchItemRequest []BatchItemRequest `json:"BatchItemRequest"`
		}

		var res struct {
			BatchItemResponses []BatchItemResponse `json:"BatchItemResponse"`
			Time               time.Time           `json:"time"`
		}

		payload.BatchItemRequest = batch

		err := c.batch(params, payload, &res)
		if err != nil {
			return nil, fmt.Errorf("failed to complete batch request: %w", err)
		}

		allResponses = append(allResponses, res.BatchItemResponses...)
	}

	return allResponses, nil
}

func BatchEntityExtractor[T any](
	resp *BatchItemResponse,
	getEntity func(BatchItemResponse) T,
) (T, bool) {
	var zero T
	entity := getEntity(*resp)
	if !reflect.ValueOf(entity).IsZero() {
		return entity, true
	}
	return zero, false
}

func BatchQueryExtractor[T any](
	resp *BatchItemResponse,
	getSlice func(BatchQueryResponse) []T,
) []T {
	// If QueryResponse is its zero value, nothing to extract
	if reflect.DeepEqual(resp.QueryResponse, BatchQueryResponse{}) {
		return nil
	}
	slice := getSlice(resp.QueryResponse)
	if len(slice) > 0 {
		return slice
	}
	return nil
}
