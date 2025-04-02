package quickbooks

import (
	"encoding/json"
	"errors"
	"strconv"
)

// Vendor describes a vendor.
type Vendor struct {
	TermRef             *ReferenceType       `json:",omitempty"`
	CurrencyRef         *ReferenceType       `json:",omitempty"`
	PrimaryPhone        *TelephoneNumber     `json:",omitempty"`
	AlternatePhone      *TelephoneNumber     `json:",omitempty"`
	Mobile              *TelephoneNumber     `json:",omitempty"`
	Fax                 *TelephoneNumber     `json:",omitempty"`
	PrimaryEmailAddr    *EmailAddress        `json:",omitempty"`
	WebAddr             *WebSiteAddress      `json:",omitempty"`
	BillAddr            *PhysicalAddress     `json:",omitempty"`
	OtherContactInfo    *ContactInfo         `json:",omitempty"`
	MetaData            ModificationMetaData `json:",omitempty"`
	CostRate            json.Number          `json:",omitempty"`
	BillRate            json.Number          `json:",omitempty"`
	Balance             json.Number          `json:",omitempty"`
	Id                  string               `json:",omitempty"`
	SyncToken           string               `json:",omitempty"`
	Title               string               `json:",omitempty"`
	GivenName           string               `json:",omitempty"`
	MiddleName          string               `json:",omitempty"`
	Suffix              string               `json:",omitempty"`
	FamilyName          string               `json:",omitempty"`
	DisplayName         string               `json:",omitempty"`
	CompanyName         string               `json:",omitempty"`
	TaxIdentifier       string               `json:",omitempty"`
	AcctNum             string               `json:",omitempty"`
	GSTRegistrationType string               `json:",omitempty"`
	PrintOnCheckName    string               `json:",omitempty"`
	Active              bool                 `json:",omitempty"`
	Vendor1099          bool                 `json:",omitempty"`
	Domain              string               `json:"domain,omitempty"`
	Status              string               `json:"status,omitempty"`
	// Source
	// APAccountRef
	// GSTIN
	// GSTRegistrationType
	// T4AEligible
	// T5018Eligible
	// BusinessNumber
	// HasTPAR
	// TaxReportingBasis
	// VendorPaymentBankDetail
}

// CreateVendor creates the given Vendor on the QuickBooks server, returning
// the resulting Vendor object.
func (c *Client) CreateVendor(params RequestParameters, vendor *Vendor) (*Vendor, error) {
	var resp struct {
		Vendor Vendor
		Time   Date
	}

	if err := c.post(params, "vendor", vendor, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Vendor, nil
}

// FindVendors gets the full list of Vendors in the QuickBooks account.
func (c *Client) FindVendors(params RequestParameters) ([]Vendor, error) {
	var resp struct {
		QueryResponse struct {
			Vendors       []Vendor `json:"Vendor"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Vendor", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, nil
	}

	vendors := make([]Vendor, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Vendor ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		vendors = append(vendors, resp.QueryResponse.Vendors...)
	}

	return vendors, nil
}

func (c *Client) FindVendorsByPage(params RequestParameters, startPosition, pageSize int) ([]Vendor, error) {
	var resp struct {
		QueryResponse struct {
			Vendors       []Vendor `json:"Vendor"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	query := "SELECT * FROM Vendor ORDERBY Id STARTPOSITION " + strconv.Itoa(startPosition) + " MAXRESULTS " + strconv.Itoa(pageSize)

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Vendors, nil
}

// FindVendorById finds the vendor by the given id
func (c *Client) FindVendorById(params RequestParameters, id string) (*Vendor, error) {
	var resp struct {
		Vendor Vendor
		Time   Date
	}

	if err := c.get(params, "vendor/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Vendor, nil
}

// QueryVendors accepts an SQL query and returns all vendors found using it
func (c *Client) QueryVendors(params RequestParameters, query string) ([]Vendor, error) {
	var resp struct {
		QueryResponse struct {
			Vendors       []Vendor `json:"Vendor"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	return resp.QueryResponse.Vendors, nil
}

// UpdateVendor full updates the vendor, meaning that missing writable fields will be set to nil/null
func (c *Client) UpdateVendor(params RequestParameters, vendor *Vendor) (*Vendor, error) {
	if vendor.Id == "" {
		return nil, errors.New("missing vendor id")
	}

	existingVendor, err := c.FindVendorById(params, vendor.Id)
	if err != nil {
		return nil, err
	}

	vendor.SyncToken = existingVendor.SyncToken

	payload := struct {
		*Vendor
	}{
		Vendor: vendor,
	}

	var vendorData struct {
		Vendor Vendor
		Time   Date
	}

	if err = c.post(params, "vendor", payload, &vendorData, nil); err != nil {
		return nil, err
	}

	return &vendorData.Vendor, err
}
