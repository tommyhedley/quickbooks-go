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
	BID           string `json:"bId"`
	Invoice       `json:",omitempty"`
	Fault         BatchFaultResponse `json:",omitempty"`
	QueryResponse struct {
		Invoice       []Invoice `json:",omitempty"`
		StartPosition int       `json:"startPosition"`
		MaxResults    int       `json:"maxResults"`
		TotalCount    int       `json:"totalCount,omitempty"`
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
