package quickbooks

import (
	"context"
	"errors"
	"strconv"
)

type Class struct {
	ParentRef          ReferenceType        `json:",omitempty"`
	MetaDate           ModificationMetaData `json:",omitempty"`
	Id                 string               `json:",omitempty"`
	Name               string               `json:",omitempty"`
	FullyQualifiedName string               `json:",omitempty"`
	SyncToken          string               `json:",omitempty"`
	SubClass           bool                 `json:",omitempty"`
	Active             bool                 `json:",omitempty"`
	Domain             string               `json:"domain,omitempty"`
	Status             string               `json:"status,omitempty"`
}

// CreateClass creates the given Class on the QuickBooks server, returning
// the resulting Class object.
func (c *Client) CreateClass(ctx context.Context, params RequestParameters, class *Class) (*Class, error) {
	var resp struct {
		Class Class
		Time  Date
	}

	if err := c.post(ctx, params, "class", class, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Class, nil
}

// FindClasss gets the full list of Classs in the QuickBooks account.
func (c *Client) FindClasses(ctx context.Context, params RequestParameters) ([]Class, error) {
	var resp struct {
		QueryResponse struct {
			Classes       []Class `json:"Class"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(ctx, params, "SELECT COUNT(*) FROM Class", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	classes := make([]Class, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Class ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(ctx, params, query, &resp); err != nil {
			return nil, err
		}

		classes = append(classes, resp.QueryResponse.Classes...)
	}

	return classes, nil
}

func (c *Client) FindClassesByPage(ctx context.Context, params RequestParameters, startPosition, pageSize int) ([]Class, error) {
	var resp struct {
		QueryResponse struct {
			Classes       []Class `json:"Class"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Class ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(ctx, params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Classes, nil
}

// FindClassById finds the class by the given id
func (c *Client) FindClassById(ctx context.Context, params RequestParameters, id string) (*Class, error) {
	var resp struct {
		Class Class
		Time  Date
	}

	if err := c.get(ctx, params, "class/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Class, nil
}

// QueryClasss accepts an SQL query and returns all classs found using it
func (c *Client) QueryClasses(ctx context.Context, params RequestParameters, query string) ([]Class, error) {
	var resp struct {
		QueryResponse struct {
			Classes       []Class `json:"Class"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(ctx, params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Classes, nil
}

// UpdateClass full updates the class, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateClass(ctx context.Context, params RequestParameters, class *Class) (*Class, error) {
	if class.Id == "" {
		return nil, errors.New("missing class id")
	}

	existingClass, err := c.FindClassById(ctx, params, class.Id)
	if err != nil {
		return nil, err
	}

	class.SyncToken = existingClass.SyncToken

	payload := struct {
		*Class
	}{
		Class: class,
	}

	var classData struct {
		Class Class
		Time  Date
	}

	if err = c.post(ctx, params, "class", payload, &classData, nil); err != nil {
		return nil, err
	}

	return &classData.Class, err
}
