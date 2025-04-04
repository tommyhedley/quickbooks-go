package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Term struct {
	MetaData           ModificationMetaData `json:",omitempty"`
	DiscountPercent    json.Number          `json:",omitempty"`
	DiscountDays       json.Number          `json:",omitempty"`
	DayOfMonthDue      json.Number          `json:",omitempty"`
	DiscountDayOfMonth json.Number          `json:",omitempty"`
	DueNextMonthDays   json.Number          `json:",omitempty"`
	DueDays            json.Number          `json:",omitempty"`
	Id                 string               `json:",omitempty"`
	Name               string               `json:",omitempty"`
	SyncToken          string               `json:",omitempty"`
	Type               string               `json:",omitempty"`
	Active             bool                 `json:",omitempty"`
	Domain             string               `json:"domain,omitempty"`
	Status             string               `json:"status,omitempty"`
}

// CreateTerm creates the given Term on the QuickBooks server, returning
// the resulting Term object.
func (c *Client) CreateTerm(params RequestParameters, term *Term) (*Term, error) {
	var resp struct {
		Term Term
		Time Date
	}

	if err := c.post(params, "term", term, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Term, nil
}

// FindTerms gets the full list of Terms in the QuickBooks account.
func (c *Client) FindTerms(params RequestParameters) ([]Term, error) {
	var resp struct {
		QueryResponse struct {
			Terms         []Term `json:"Term"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Term", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	terms := make([]Term, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Term ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		terms = append(terms, resp.QueryResponse.Terms...)
	}

	return terms, nil
}

func (c *Client) FindTermsByPage(params RequestParameters, startPosition, pageSize int) ([]Term, error) {
	var resp struct {
		QueryResponse struct {
			Terms         []Term `json:"Term"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Term ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Terms, nil
}

// FindTermById finds the term by the given id
func (c *Client) FindTermById(params RequestParameters, id string) (*Term, error) {
	var resp struct {
		Term Term
		Time Date
	}

	if err := c.get(params, "term/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Term, nil
}

// QueryTerms accepts an SQL query and returns all terms found using it
func (c *Client) QueryTerms(params RequestParameters, query string) ([]Term, error) {
	var resp struct {
		QueryResponse struct {
			Terms         []Term `json:"Term"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Terms, nil
}

// UpdateTerm full updates the term, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateTerm(params RequestParameters, term *Term) (*Term, error) {
	if term.Id == "" {
		return nil, errors.New("missing term id")
	}

	existingTerm, err := c.FindTermById(params, term.Id)
	if err != nil {
		return nil, err
	}

	term.SyncToken = existingTerm.SyncToken

	payload := struct {
		*Term
	}{
		Term: term,
	}

	var termData struct {
		Term Term
		Time Date
	}

	if err = c.post(params, "term", payload, &termData, nil); err != nil {
		return nil, err
	}

	return &termData.Term, err
}
