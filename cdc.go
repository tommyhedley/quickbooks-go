package quickbooks

import (
	"fmt"
	"reflect"
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

func CDCQueryExtractor[T any](response *ChangeDataCapture) []T {
	var zero T
	typeName := reflect.TypeOf(zero).Name()

	for _, resp := range response.CDCResponse {
		for _, query := range resp.QueryResponse {
			switch typeName {
			case "Account":
				if len(query.Account) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Account{}) {
					return reflect.ValueOf(query.Account).Interface().([]T)
				}
			case "Bill":
				if len(query.Bill) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Bill{}) {
					return reflect.ValueOf(query.Bill).Interface().([]T)
				}
			case "BillPayment":
				if len(query.BillPayment) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(BillPayment{}) {
					return reflect.ValueOf(query.BillPayment).Interface().([]T)
				}
			case "Class":
				if len(query.Class) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Class{}) {
					return reflect.ValueOf(query.Class).Interface().([]T)
				}
			case "Customer":
				if len(query.Customer) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Customer{}) {
					return reflect.ValueOf(query.Customer).Interface().([]T)
				}
			case "CustomerType":
				if len(query.CustomerType) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(CustomerType{}) {
					return reflect.ValueOf(query.CustomerType).Interface().([]T)
				}
			case "Deposit":
				if len(query.Deposit) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Deposit{}) {
					return reflect.ValueOf(query.Deposit).Interface().([]T)
				}
			case "Employee":
				if len(query.Employee) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Employee{}) {
					return reflect.ValueOf(query.Employee).Interface().([]T)
				}
			case "Estimate":
				if len(query.Estimate) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Estimate{}) {
					return reflect.ValueOf(query.Estimate).Interface().([]T)
				}
			case "Invoice":
				if len(query.Invoice) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Invoice{}) {
					return reflect.ValueOf(query.Invoice).Interface().([]T)
				}
			case "Item":
				if len(query.Item) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Item{}) {
					return reflect.ValueOf(query.Item).Interface().([]T)
				}
			case "Payment":
				if len(query.Payment) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Payment{}) {
					return reflect.ValueOf(query.Payment).Interface().([]T)
				}
			case "PaymentMethod":
				if len(query.PaymentMethod) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(PaymentMethod{}) {
					return reflect.ValueOf(query.PaymentMethod).Interface().([]T)
				}
			case "Purchase":
				if len(query.Purchase) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Purchase{}) {
					return reflect.ValueOf(query.Purchase).Interface().([]T)
				}
			case "ReimburseCharge":
				if len(query.ReimburseCharge) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(ReimburseCharge{}) {
					return reflect.ValueOf(query.ReimburseCharge).Interface().([]T)
				}
			case "Term":
				if len(query.Term) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Term{}) {
					return reflect.ValueOf(query.Term).Interface().([]T)
				}
			case "Vendor":
				if len(query.Vendor) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(Vendor{}) {
					return reflect.ValueOf(query.Vendor).Interface().([]T)
				}
			case "VendorCredit":
				if len(query.VendorCredit) > 0 && reflect.TypeOf(zero) == reflect.TypeOf(VendorCredit{}) {
					return reflect.ValueOf(query.VendorCredit).Interface().([]T)
				}
			}
		}
	}
	return nil
}
