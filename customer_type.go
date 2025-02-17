package quickbooks

import (
	"errors"
	"strconv"
)

type CustomerType struct {
	Id        string   `json:",omitempty"`
	Name      string   `json:",omitempty"`
	SyncToken string   `json:",omitempty"`
	Active    bool     `json:",omitempty"`
	MetaData  MetaData `json:",omitempty"`
}

type CDCCustomerType struct {
	CustomerType
	Domain string `json:"domain,omitempty"`
	Status string `json:"status,omitempty"`
}

// FindCustomerTypeById returns a customerType with a given Id.
func (c *Client) FindCustomerTypeById(id string) (*CustomerType, error) {
	var r struct {
		CustomerType CustomerType
		Time         Date
	}

	if err := c.get("customertype/"+id, &r, nil); err != nil {
		return nil, err
	}

	return &r.CustomerType, nil
}

func (c *Client) FindCustomerTypesByPage(startPosition int, pageSize int) ([]CustomerType, error) {
	var resp struct {
		QueryResponse struct {
			CustomerTypes []CustomerType `json:"CustomerType"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM CustomerType ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.CustomerTypes == nil {
		return nil, errors.New("no customer types could be found")
	}

	return resp.QueryResponse.CustomerTypes, nil
}

// QueryCustomerTypes accepts an SQL query and returns all customerTypes found using it
func (c *Client) QueryCustomerTypes(query string) ([]CustomerType, error) {
	var resp struct {
		QueryResponse struct {
			CustomerTypes []CustomerType `json:"CustomerType"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.CustomerTypes == nil {
		return nil, errors.New("could not find any customer types")
	}

	return resp.QueryResponse.CustomerTypes, nil
}
