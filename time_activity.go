package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type TimeActivity struct {
	VendorRef      ReferenceType        `json:",omitempty"`
	EmployeeRef    ReferenceType        `json:",omitempty"`
	CustomerRef    ReferenceType        `json:",omitempty"`
	ProjectRef     ReferenceType        `json:",omitempty"`
	ClassRef       *ReferenceType       `json:",omitempty"`
	ItemRef        *ReferenceType       `json:",omitempty"`
	DepartmentRef  *ReferenceType       `json:",omitempty"`
	PayrollItemRef *ReferenceType       `json:",omitempty"`
	TxnDate        Date                 `json:",omitempty"`
	StartTime      DateTime             `json:",omitempty"`
	EndTime        DateTime             `json:",omitempty"`
	MetaData       ModificationMetaData `json:",omitempty"`
	BillableStatus BillableStatusEnum   `json:",omitempty"`
	BreakHours     json.Number          `json:",omitempty"`
	BreakMinutes   json.Number          `json:",omitempty"`
	BreakSeconds   json.Number          `json:",omitempty"`
	Hours          json.Number          `json:",omitempty"`
	Minutes        json.Number          `json:",omitempty"`
	Seconds        json.Number          `json:",omitempty"`
	HourlyRate     json.Number          `json:",omitempty"`
	CostRate       json.Number          `json:",omitempty"`
	Id             string               `json:",omitempty"`
	NameOf         string               `json:",omitempty"`
	SyncToken      string               `json:",omitempty"`
	Description    string               `json:",omitempty"`
	Taxable        bool                 `json:",omitempty"`
	// TransactionLocationType
}

// CreateTimeActivity creates the given TimeActivity on the QuickBooks server, returning
// the resulting TimeActivity object.
func (c *Client) CreateTimeActivity(req RequestParameters, timeActivity *TimeActivity) (*TimeActivity, error) {
	var resp struct {
		TimeActivity TimeActivity
		Time         Date
	}

	if err := c.post(req, "timeactivity", timeActivity, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TimeActivity, nil
}

// DeleteTimeActivity deletes the timeActivity
func (c *Client) DeleteTimeActivity(req RequestParameters, timeActivity *TimeActivity) error {
	if timeActivity.Id == "" || timeActivity.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(req, "timeactivity", timeActivity, nil, map[string]string{"operation": "delete"})
}

// FindTimeActivitys gets the full list of TimeActivitys in the QuickBooks account.
func (c *Client) FindTimeActivities(req RequestParameters) ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivitys []TimeActivity `json:"TimeActivity"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(req, "SELECT COUNT(*) FROM TimeActivity", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no time activitys could be found")
	}

	timeActivitys := make([]TimeActivity, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM TimeActivity ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(req, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.TimeActivitys == nil {
			return nil, errors.New("no time activitys could be found")
		}

		timeActivitys = append(timeActivitys, resp.QueryResponse.TimeActivitys...)
	}

	return timeActivitys, nil
}

func (c *Client) FindTimeActivitiesByPage(req RequestParameters, startPosition, pageSize int) ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivitys []TimeActivity `json:"TimeActivity"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM TimeActivity ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(req, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TimeActivitys == nil {
		return nil, errors.New("no time activitys could be found")
	}

	return resp.QueryResponse.TimeActivitys, nil
}

// FindTimeActivityById finds the timeActivity by the given id
func (c *Client) FindTimeActivityById(req RequestParameters, id string) (*TimeActivity, error) {
	var resp struct {
		TimeActivity TimeActivity
		Time         Date
	}

	if err := c.get(req, "timeactivity/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TimeActivity, nil
}

// QueryTimeActivitys accepts an SQL query and returns all timeActivitys found using it
func (c *Client) QueryTimeActivities(req RequestParameters, query string) ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivitys []TimeActivity `json:"TimeActivity"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(req, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TimeActivitys == nil {
		return nil, errors.New("could not find any time activitys")
	}

	return resp.QueryResponse.TimeActivitys, nil
}

// UpdateTimeActivity full updates the time activity, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateTimeActivity(req RequestParameters, timeActivity *TimeActivity) (*TimeActivity, error) {
	if timeActivity.Id == "" {
		return nil, errors.New("missing time activity id")
	}

	existingTimeActivity, err := c.FindTimeActivityById(req, timeActivity.Id)
	if err != nil {
		return nil, err
	}

	timeActivity.SyncToken = existingTimeActivity.SyncToken

	payload := struct {
		*TimeActivity
	}{
		TimeActivity: timeActivity,
	}

	var timeActivityData struct {
		TimeActivity TimeActivity
		Time         Date
	}

	if err = c.post(req, "timeactivity", payload, &timeActivityData, nil); err != nil {
		return nil, err
	}

	return &timeActivityData.TimeActivity, err
}
