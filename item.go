// Copyright (c) 2018, Randy Westlund. All rights reserved.
// This code is under the BSD-2-Clause license.

package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

// Item represents a QuickBooks Item object (a product type).
type Item struct {
	AssetAccountRef      ReferenceType        `json:",omitempty"`
	IncomeAccountRef     ReferenceType        `json:",omitempty"`
	ExpenseAccountRef    *ReferenceType       `json:",omitempty"`
	SalesTaxCodeRef      *ReferenceType       `json:",omitempty"`
	PurchaseTaxCodeRef   *ReferenceType       `json:",omitempty"`
	TaxClassificationRef *ReferenceType       `json:",omitempty"`
	ClassRef             *ReferenceType       `json:",omitempty"`
	PrefVendorRef        *ReferenceType       `json:",omitempty"`
	ParentRef            *ReferenceType       `json:",omitempty"`
	InvStartDate         Date                 `json:",omitempty"`
	MetaData             ModificationMetaData `json:",omitempty"`
	QtyOnHand            json.Number          `json:",omitempty"`
	ReorderPoint         json.Number          `json:",omitempty"`
	PurchaseCost         json.Number          `json:",omitempty"`
	UnitPrice            json.Number          `json:",omitempty"`
	Level                json.Number          `json:",omitempty"`
	Id                   string               `json:",omitempty"`
	SyncToken            string               `json:",omitempty"`
	Name                 string               `json:",omitempty"`
	FullyQualifiedName   string               `json:",omitempty"`
	SKU                  string               `json:"Sku,omitempty"`
	Description          string               `json:",omitempty"`
	PurchaseDesc         string               `json:",omitempty"`
	Type                 string               `json:",omitempty"`
	TrackQtyOnHand       bool                 `json:",omitempty"`
	Active               bool                 `json:",omitempty"`
	Taxable              bool                 `json:",omitempty"`
	SalesTaxIncluded     bool                 `json:",omitempty"`
	PurchaseTaxIncluded  bool                 `json:",omitempty"`
	SubItem              bool                 `json:",omitempty"`
	// ItemCategoryType
	// AbatementRate
	// UQCDisplayText
	// UQCId
	// ReverseChargeRate
	// ServiceType
}

type CDCItem struct {
	Item
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

func (c *Client) CreateItem(params RequestParameters, item *Item) (*Item, error) {
	var resp struct {
		Item Item
		Time Date
	}

	if err := c.post(params, "item", item, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// FindItems gets the full list of Items in the QuickBooks account.
func (c *Client) FindItems(params RequestParameters) ([]Item, error) {
	var resp struct {
		QueryResponse struct {
			Items         []Item `json:"Item"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Item", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no items could be found")
	}

	items := make([]Item, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Item ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Items == nil {
			return nil, errors.New("no items could be found")
		}

		items = append(items, resp.QueryResponse.Items...)
	}

	return items, nil
}

func (c *Client) FindItemsByPage(params RequestParameters, startPosition, pageSize int) ([]Item, error) {
	var resp struct {
		QueryResponse struct {
			Items         []Item `json:"Item"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Item ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Items == nil {
		return nil, errors.New("no items could be found")
	}

	return resp.QueryResponse.Items, nil
}

// FindItemById returns an item with a given Id.
func (c *Client) FindItemById(params RequestParameters, id string) (*Item, error) {
	var resp struct {
		Item Item
		Time Date
	}

	if err := c.get(params, "item/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Item, nil
}

// QueryItems accepts an SQL query and returns all items found using it
func (c *Client) QueryItems(params RequestParameters, query string) ([]Item, error) {
	var resp struct {
		QueryResponse struct {
			Items         []Item `json:"Item"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Items == nil {
		return nil, errors.New("could not find any items")
	}

	return resp.QueryResponse.Items, nil
}

// UpdateItem full updates the item, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateItem(params RequestParameters, item *Item) (*Item, error) {
	if item.Id == "" {
		return nil, errors.New("missing item id")
	}

	existingItem, err := c.FindItemById(params, item.Id)
	if err != nil {
		return nil, err
	}

	item.SyncToken = existingItem.SyncToken

	payload := struct {
		*Item
	}{
		Item: item,
	}

	var itemData struct {
		Item Item
		Time Date
	}

	if err = c.post(params, "item", payload, &itemData, nil); err != nil {
		return nil, err
	}

	return &itemData.Item, err
}
