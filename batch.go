package quickbooks

import (
	"fmt"
	"reflect"
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

func BatchEntityExtractor[T any](response *BatchItemResponse) (T, bool) {
	var zero T
	typeName := reflect.TypeOf(zero).Name()

	switch typeName {
	case "Account":
		if !reflect.ValueOf(response.Account).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Account{}) {
			return reflect.ValueOf(response.Account).Interface().(T), true
		}
	case "Bill":
		if !reflect.ValueOf(response.Bill).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Bill{}) {
			return reflect.ValueOf(response.Bill).Interface().(T), true
		}
	case "BillPayment":
		if !reflect.ValueOf(response.BillPayment).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(BillPayment{}) {
			return reflect.ValueOf(response.BillPayment).Interface().(T), true
		}
	case "Class":
		if !reflect.ValueOf(response.Class).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Class{}) {
			return reflect.ValueOf(response.Class).Interface().(T), true
		}
	case "CreditMemo":
		if !reflect.ValueOf(response.CreditMemo).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(CreditMemo{}) {
			return reflect.ValueOf(response.CreditMemo).Interface().(T), true
		}
	case "Customer":
		if !reflect.ValueOf(response.Customer).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Customer{}) {
			return reflect.ValueOf(response.Customer).Interface().(T), true
		}
	case "CustomerType":
		if !reflect.ValueOf(response.CustomerType).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(CustomerType{}) {
			return reflect.ValueOf(response.CustomerType).Interface().(T), true
		}
	case "Deposit":
		if !reflect.ValueOf(response.Deposit).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Deposit{}) {
			return reflect.ValueOf(response.Deposit).Interface().(T), true
		}
	case "Employee":
		if !reflect.ValueOf(response.Employee).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Employee{}) {
			return reflect.ValueOf(response.Employee).Interface().(T), true
		}
	case "Estimate":
		if !reflect.ValueOf(response.Estimate).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Estimate{}) {
			return reflect.ValueOf(response.Estimate).Interface().(T), true
		}
	case "Invoice":
		if !reflect.ValueOf(response.Invoice).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Invoice{}) {
			return reflect.ValueOf(response.Invoice).Interface().(T), true
		}
	case "Item":
		if !reflect.ValueOf(response.Item).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Item{}) {
			return reflect.ValueOf(response.Item).Interface().(T), true
		}
	case "Payment":
		if !reflect.ValueOf(response.Payment).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Payment{}) {
			return reflect.ValueOf(response.Payment).Interface().(T), true
		}
	case "PaymentMethod":
		if !reflect.ValueOf(response.PaymentMethod).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(PaymentMethod{}) {
			return reflect.ValueOf(response.PaymentMethod).Interface().(T), true
		}
	case "Purchase":
		if !reflect.ValueOf(response.Purchase).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Purchase{}) {
			return reflect.ValueOf(response.Purchase).Interface().(T), true
		}
	case "ReimburseCharge":
		if !reflect.ValueOf(response.ReimburseCharge).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(ReimburseCharge{}) {
			return reflect.ValueOf(response.ReimburseCharge).Interface().(T), true
		}
	case "TaxCode":
		if !reflect.ValueOf(response.TaxCode).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(TaxCode{}) {
			return reflect.ValueOf(response.TaxCode).Interface().(T), true
		}
	case "TaxRate":
		if !reflect.ValueOf(response.TaxRate).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(TaxRate{}) {
			return reflect.ValueOf(response.TaxRate).Interface().(T), true
		}
	case "Term":
		if !reflect.ValueOf(response.Term).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Term{}) {
			return reflect.ValueOf(response.Term).Interface().(T), true
		}
	case "TimeActivity":
		if !reflect.ValueOf(response.TimeActivity).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(TimeActivity{}) {
			return reflect.ValueOf(response.TimeActivity).Interface().(T), true
		}
	case "Vendor":
		if !reflect.ValueOf(response.Vendor).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(Vendor{}) {
			return reflect.ValueOf(response.Vendor).Interface().(T), true
		}
	case "VendorCredit":
		if !reflect.ValueOf(response.VendorCredit).IsZero() && reflect.TypeOf(zero) == reflect.TypeOf(VendorCredit{}) {
			return reflect.ValueOf(response.VendorCredit).Interface().(T), true
		}
	}

	return zero, false
}

func BatchQueryExtractor[T any](response *BatchItemResponse) []T {
	if reflect.ValueOf(response.QueryResponse).IsZero() {
		return nil
	}

	var zero T
	typeName := reflect.TypeOf(zero).Name()

	switch typeName {
	case "Account":
		if len(response.QueryResponse.Account) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Account{}) {
			return reflect.ValueOf(response.QueryResponse.Account).Interface().([]T)
		}
	case "Bill":
		if len(response.QueryResponse.Bill) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Bill{}) {
			return reflect.ValueOf(response.QueryResponse.Bill).Interface().([]T)
		}
	case "BillPayment":
		if len(response.QueryResponse.BillPayment) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(BillPayment{}) {
			return reflect.ValueOf(response.QueryResponse.BillPayment).Interface().([]T)
		}
	case "Class":
		if len(response.QueryResponse.Class) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Class{}) {
			return reflect.ValueOf(response.QueryResponse.Class).Interface().([]T)
		}
	case "CreditMemo":
		if len(response.QueryResponse.CreditMemo) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(CreditMemo{}) {
			return reflect.ValueOf(response.QueryResponse.CreditMemo).Interface().([]T)
		}
	case "Customer":
		if len(response.QueryResponse.Customer) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Customer{}) {
			return reflect.ValueOf(response.QueryResponse.Customer).Interface().([]T)
		}
	case "CustomerType":
		if len(response.QueryResponse.CustomerType) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(CustomerType{}) {
			return reflect.ValueOf(response.QueryResponse.CustomerType).Interface().([]T)
		}
	case "Deposit":
		if len(response.QueryResponse.Deposit) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Deposit{}) {
			return reflect.ValueOf(response.QueryResponse.Deposit).Interface().([]T)
		}
	case "Employee":
		if len(response.QueryResponse.Employee) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Employee{}) {
			return reflect.ValueOf(response.QueryResponse.Employee).Interface().([]T)
		}
	case "Estimate":
		if len(response.QueryResponse.Estimate) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Estimate{}) {
			return reflect.ValueOf(response.QueryResponse.Estimate).Interface().([]T)
		}
	case "Invoice":
		if len(response.QueryResponse.Invoice) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Invoice{}) {
			return reflect.ValueOf(response.QueryResponse.Invoice).Interface().([]T)
		}
	case "Item":
		if len(response.QueryResponse.Item) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Item{}) {
			return reflect.ValueOf(response.QueryResponse.Item).Interface().([]T)
		}
	case "Payment":
		if len(response.QueryResponse.Payment) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Payment{}) {
			return reflect.ValueOf(response.QueryResponse.Payment).Interface().([]T)
		}
	case "PaymentMethod":
		if len(response.QueryResponse.PaymentMethod) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(PaymentMethod{}) {
			return reflect.ValueOf(response.QueryResponse.PaymentMethod).Interface().([]T)
		}
	case "Purchase":
		if len(response.QueryResponse.Purchase) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Purchase{}) {
			return reflect.ValueOf(response.QueryResponse.Purchase).Interface().([]T)
		}
	case "ReimburseCharge":
		if len(response.QueryResponse.ReimburseCharge) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(ReimburseCharge{}) {
			return reflect.ValueOf(response.QueryResponse.ReimburseCharge).Interface().([]T)
		}
	case "TaxCode":
		if len(response.QueryResponse.TaxCode) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(TaxCode{}) {
			return reflect.ValueOf(response.QueryResponse.TaxCode).Interface().([]T)
		}
	case "TaxRate":
		if len(response.QueryResponse.TaxRate) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(TaxRate{}) {
			return reflect.ValueOf(response.QueryResponse.TaxRate).Interface().([]T)
		}
	case "Term":
		if len(response.QueryResponse.Term) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Term{}) {
			return reflect.ValueOf(response.QueryResponse.Term).Interface().([]T)
		}
	case "TimeActivity":
		if len(response.QueryResponse.TimeActivity) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(TimeActivity{}) {
			return reflect.ValueOf(response.QueryResponse.TimeActivity).Interface().([]T)
		}
	case "Vendor":
		if len(response.QueryResponse.Vendor) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Vendor{}) {
			return reflect.ValueOf(response.QueryResponse.Vendor).Interface().([]T)
		}
	case "VendorCredit":
		if len(response.QueryResponse.VendorCredit) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(VendorCredit{}) {
			return reflect.ValueOf(response.QueryResponse.VendorCredit).Interface().([]T)
		}
	}

	return nil
}
