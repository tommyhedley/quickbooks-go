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
	StartTime      *DateTime            `json:",omitempty"`
	EndTime        *DateTime            `json:",omitempty"`
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
func (c *Client) CreateTimeActivity(params RequestParameters, timeActivity *TimeActivity) (*TimeActivity, error) {
	var resp struct {
		TimeActivity TimeActivity
		Time         Date
	}

	if err := c.post(params, "timeactivity", timeActivity, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TimeActivity, nil
}

// DeleteTimeActivity deletes the timeActivity
func (c *Client) DeleteTimeActivity(params RequestParameters, timeActivity *TimeActivity) error {
	if timeActivity.Id == "" || timeActivity.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "timeactivity", timeActivity, nil, map[string]string{"operation": "delete"})
}

// FindTimeActivitys gets the full list of TimeActivitys in the QuickBooks account.
func (c *Client) FindTimeActivities(params RequestParameters) ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivities []TimeActivity `json:"TimeActivity"`
			MaxResults     int
			StartPosition  int
			TotalCount     int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM TimeActivity", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	timeActivities := make([]TimeActivity, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM TimeActivity ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		timeActivities = append(timeActivities, resp.QueryResponse.TimeActivities...)
	}

	return timeActivities, nil
}

func (c *Client) FindTimeActivitiesByPage(params RequestParameters, startPosition, pageSize int) ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivities []TimeActivity `json:"TimeActivity"`
			MaxResults     int
			StartPosition  int
			TotalCount     int
		}
	}

	query := "SELECT * FROM TimeActivity ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.TimeActivities, nil
}

// FindTimeActivityById finds the timeActivity by the given id
func (c *Client) FindTimeActivityById(params RequestParameters, id string) (*TimeActivity, error) {
	var resp struct {
		TimeActivity TimeActivity
		Time         Date
	}

	if err := c.get(params, "timeactivity/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.TimeActivity, nil
}

// QueryTimeActivitys accepts an SQL query and returns all timeActivitys found using it
func (c *Client) QueryTimeActivities(params RequestParameters, query string) ([]TimeActivity, error) {
	var resp struct {
		QueryResponse struct {
			TimeActivities []TimeActivity `json:"TimeActivity"`
			StartPosition  int
			MaxResults     int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.TimeActivities, nil
}

// UpdateTimeActivity full updates the time activity, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateTimeActivity(params RequestParameters, timeActivity *TimeActivity) (*TimeActivity, error) {
	if timeActivity.Id == "" {
		return nil, errors.New("missing time activity id")
	}

	existingTimeActivity, err := c.FindTimeActivityById(params, timeActivity.Id)
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

	if err = c.post(params, "timeactivity", payload, &timeActivityData, nil); err != nil {
		return nil, err
	}

	return &timeActivityData.TimeActivity, err
}
