package quickbooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strconv"
)

type ContentType string

const (
	AI   ContentType = "application/postscript"
	CSV  ContentType = "text/csv"
	DOC  ContentType = "application/msword"
	DOCX ContentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	EPS  ContentType = "application/postscript"
	GIF  ContentType = "image/gif"
	JPEG ContentType = "image/jpeg"
	JPG  ContentType = "image/jpg"
	ODS  ContentType = "application/vnd.oasis.opendocument.spreadsheet"
	PDF  ContentType = "application/pdf"
	PNG  ContentType = "image/png"
	RTF  ContentType = "text/rtf"
	TIF  ContentType = "image/tif"
	TXT  ContentType = "text/plain"
	XLS  ContentType = "application/vnd/ms-excel"
	XLSX ContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	XML  ContentType = "text/xml"
)

type Attachable struct {
	Id                       string               `json:"Id,omitempty"`
	SyncToken                string               `json:",omitempty"`
	FileName                 string               `json:",omitempty"`
	Note                     string               `json:",omitempty"`
	Category                 string               `json:",omitempty"`
	ContentType              ContentType          `json:",omitempty"`
	PlaceName                string               `json:",omitempty"`
	AttachableRef            []AttachableRef      `json:",omitempty"`
	Long                     string               `json:",omitempty"`
	Tag                      string               `json:",omitempty"`
	Lat                      string               `json:",omitempty"`
	MetaData                 ModificationMetaData `json:",omitempty"`
	FileAccessUri            string               `json:",omitempty"`
	Size                     json.Number          `json:",omitempty"`
	ThumbnailFileAccessUri   string               `json:",omitempty"`
	TempDownloadUri          string               `json:",omitempty"`
	ThumbnailTempDownloadUri string               `json:",omitempty"`
}

type AttachableRef struct {
	IncludeOnSend bool   `json:",omitempty"`
	LineInfo      string `json:",omitempty"`
	NoRefOnly     bool   `json:",omitempty"`
	// CustomField[0..n]
	Inactive  bool          `json:",omitempty"`
	EntityRef ReferenceType `json:",omitempty"`
}

// CreateAttachable creates the given Attachable on the QuickBooks server,
// returning the resulting Attachable object.
func (c *Client) CreateAttachable(params RequestParameters, attachable *Attachable) (*Attachable, error) {
	var resp struct {
		Attachable Attachable
		Time       Date
	}

	if err := c.post(params, "attachable", attachable, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Attachable, nil
}

// DeleteAttachable deletes the attachable
func (c *Client) DeleteAttachable(params RequestParameters, attachable *Attachable) error {
	if attachable.Id == "" || attachable.SyncToken == "" {
		return errors.New("missing id/sync token")
	}

	return c.post(params, "attachable", attachable, nil, map[string]string{"operation": "delete"})
}

// DownloadAttachable downloads the attachable
func (c *Client) DownloadAttachable(params RequestParameters, id string) (*url.URL, error) {
	var urlString string
	var url *url.URL
	var err error
	if err = c.get(params, "download/"+id, &urlString, nil); err != nil {
		return nil, err
	}
	if url, err = url.Parse(urlString); err != nil {
		return nil, err
	}
	return url, nil
}

// FindAttachables gets the full list of Attachables in the QuickBooks attachable.
func (c *Client) FindAttachables(params RequestParameters) ([]Attachable, error) {
	var resp struct {
		QueryResponse struct {
			Attachables   []Attachable `json:"Attachable"`
			MaxResults    int
			StartPosition int
			TotalCount    int
		}
	}

	if err := c.query(params, "SELECT COUNT(*) FROM Attachable", &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.TotalCount == 0 {
		return nil, errors.New("no attachables could be found")
	}

	attachables := make([]Attachable, 0, resp.QueryResponse.TotalCount)

	for i := 0; i < resp.QueryResponse.TotalCount; i += QueryPageSize {
		query := "SELECT * FROM Attachable ORDERBY Id STARTPOSITION " + strconv.Itoa(i+1) + " MAXRESULTS " + strconv.Itoa(QueryPageSize)

		if err := c.query(params, query, &resp); err != nil {
			return nil, err
		}

		if resp.QueryResponse.Attachables == nil {
			return nil, errors.New("no attachables could be found")
		}

		attachables = append(attachables, resp.QueryResponse.Attachables...)
	}

	return attachables, nil
}

// FindAttachableById finds the attachable by the given id
func (c *Client) FindAttachableById(params RequestParameters, id string) (*Attachable, error) {
	var resp struct {
		Attachable Attachable
		Time       Date
	}

	if err := c.get(params, "attachable/"+id, &resp, nil); err != nil {
		return nil, err
	}

	return &resp.Attachable, nil
}

// QueryAttachables accepts an SQL query and returns all attachables found using it
func (c *Client) QueryAttachables(params RequestParameters, query string) ([]Attachable, error) {
	var resp struct {
		QueryResponse struct {
			Attachables   []Attachable `json:"Attachable"`
			StartPosition int
			MaxResults    int
		}
	}

	if err := c.query(params, query, &resp); err != nil {
		return nil, err
	}

	if resp.QueryResponse.Attachables == nil {
		return nil, errors.New("could not find any attachables")
	}

	return resp.QueryResponse.Attachables, nil
}

// UpdateAttachable updates the attachable
func (c *Client) UpdateAttachable(params RequestParameters, attachable *Attachable) (*Attachable, error) {
	if attachable.Id == "" {
		return nil, errors.New("missing attachable id")
	}

	existingAttachable, err := c.FindAttachableById(params, attachable.Id)
	if err != nil {
		return nil, err
	}

	attachable.SyncToken = existingAttachable.SyncToken

	payload := struct {
		*Attachable
		Sparse bool `json:"sparse"`
	}{
		Attachable: attachable,
		Sparse:     true,
	}

	var attachableData struct {
		Attachable Attachable
		Time       Date
	}

	if err = c.post(params, "attachable", payload, &attachableData, nil); err != nil {
		return nil, err
	}

	return &attachableData.Attachable, err
}

// UploadAttachable uploads the attachable
func (c *Client) UploadAttachable(realmId string, attachable *Attachable, data io.Reader) (*Attachable, error) {
	endpointUrl := *c.baseEndpoint
	endpointUrl.Path += realmId + "/upload"

	urlValues := url.Values{}
	urlValues.Add("minorversion", c.minorVersion)
	endpointUrl.RawQuery = urlValues.Encode()

	var buffer bytes.Buffer
	mWriter := multipart.NewWriter(&buffer)

	// Add file metadata
	metadataHeader := make(textproto.MIMEHeader)
	metadataHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file_metadata_01", "attachment.json"))
	metadataHeader.Set("Content-Type", "application/json")

	metadataContent, err := mWriter.CreatePart(metadataHeader)
	if err != nil {
		return nil, err
	}

	j, err := json.Marshal(attachable)
	if err != nil {
		return nil, err
	}

	if _, err = metadataContent.Write(j); err != nil {
		return nil, err
	}

	// Add file content
	fileHeader := make(textproto.MIMEHeader)
	fileHeader.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file_content_01", attachable.FileName))
	fileHeader.Set("Content-Type", string(attachable.ContentType))

	fileContent, err := mWriter.CreatePart(fileHeader)
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(fileContent, data); err != nil {
		return nil, err
	}

	mWriter.Close()

	req, err := http.NewRequest("POST", endpointUrl.String(), &buffer)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", mWriter.FormDataContentType())
	req.Header.Add("Accept", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, parseFailure(resp)
	}

	var r struct {
		AttachableResponse []struct {
			Attachable Attachable
		}
		Time Date
	}

	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}

	return &r.AttachableResponse[0].Attachable, nil
}
