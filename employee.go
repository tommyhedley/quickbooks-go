package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

type Employee struct {
	PrimaryAddr      PhysicalAddress      `json:",omitempty"`
	PrimaryEmailAddr *EmailAddress        `json:",omitempty"`
	PrimaryPhone     *TelephoneNumber     `json:",omitempty"`
	Mobile           *TelephoneNumber     `json:",omitempty"`
	BirthDate        *Date                `json:",omitempty"`
	HiredDate        *Date                `json:",omitempty"`
	ReleasedDate     *Date                `json:",omitempty"`
	MetaData         ModificationMetaData `json:",omitempty"`
	CostRate         json.Number          `json:",omitempty"`
	BillRate         json.Number          `json:",omitempty"`
	Id               string               `json:",omitempty"`
	SyncToken        string               `json:",omitempty"`
	Title            string               `json:",omitempty"`
	GivenName        string               `json:",omitempty"`
	MiddleName       string               `json:",omitempty"`
	FamilyName       string               `json:",omitempty"`
	Suffix           string               `json:",omitempty"`
	DisplayName      string               `json:",omitempty"`
	PrintOnCheckName string               `json:",omitempty"`
	Gender           string               `json:",omitempty"`
	EmployeeNumber   string               `json:",omitempty"`
	SSN              string               `json:",omitempty"`
	Active           bool                 `json:",omitempty"`
	BillableTime     bool                 `json:",omitempty"`
	Organization     bool                 `json:",omitempty"`
	Domain           string               `json:"domain,omitempty"`
	Status           string               `json:"status,omitempty"`
}

// CreateEmployee creates the given employee within QuickBooks
func (c *Client) CreateEmployee(params RequestParameters, employee *Employee) (*Employee, error) {
	var resp struct {
		Employee Employee
		Time     Date
	}

	if err := c.post(params, "employee", employee, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

// FindEmployees gets the full list of Employees in the QuickBooks account.
func (c *Client) FindEmployees(params RequestParameters) ([]Employee, error) {
	var resp struct {
		QueryResponse struct {
			Employees     []Employee `json:"Employee"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Employee", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	employees := make([]Employee, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Employee ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		employees = append(employees, resp.QueryResponse.Employees...)
	}

	return employees, nil
}

// FindEmployeeById returns an employee with a given Id.
func (c *Client) FindEmployeeById(params RequestParameters, id string) (*Employee, error) {
	var resp struct {
		Employee Employee
		Time     Date
	}

	if err := c.get(params, "employee/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Employee, nil
}

func (c *Client) FindEmployeesByPage(params RequestParameters, startPosition, pageSize int) ([]Employee, error) {
	var resp struct {
		QueryResponse struct {
			Employees     []Employee `json:"Employee"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Employee ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Employees, nil
}

// QueryEmployees accepts an SQL query and returns all employees found using it
func (c *Client) QueryEmployees(params RequestParameters, query string) ([]Employee, error) {
	var resp struct {
		QueryResponse struct {
			Employees     []Employee `json:"Employee"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Employees, nil
}

// UpdateEmployee updates the employee
func (c *Client) UpdateEmployee(params RequestParameters, employee *Employee) (*Employee, error) {
	if employee.Id == "" {
		return nil, errors.New("missing employee id")
	}

	existingEmployee, err := c.FindEmployeeById(params, employee.Id)
	if err != nil {
		return nil, err
	}

	employee.SyncToken = existingEmployee.SyncToken

	payload := struct {
		*Employee
		Sparse bool `json:"sparse"`
	}{
		Employee: employee,
		Sparse:   true,
	}

	var employeeData struct {
		Employee Employee
		Time     Date
	}

	if err = c.post(params, "employee", payload, &employeeData, nil); err != nil {
		return nil, err
	}

	return &employeeData.Employee, err
}
