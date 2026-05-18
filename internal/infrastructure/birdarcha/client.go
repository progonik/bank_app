package birdarcha

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ListItem represents a single entity from the paginated list endpoint.
type ListItem struct {
	ID               int    `json:"id"`
	RegistrationDate string `json:"registration_date"`
	TIN              int64  `json:"tin"`
	Name             string `json:"name"`
	BusinessType     struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"business_type"`
	BusinessStatus struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"business_status"`
	Status string `json:"status"`
}

// ListResponse is the API response wrapper for the list endpoint.
type ListResponse struct {
	Status     int        `json:"status"`
	Data       []ListItem `json:"data"`
	TotalPages int        `json:"total_pages"`
	TotalCount int        `json:"total_count"`
	Size       int        `json:"size"`
	Page       int        `json:"page"`
}

// Detail represents the full entity detail from the detail endpoint.
type Detail struct {
	ID          int   `json:"id"`
	TIN         int64 `json:"tin"`
	Name        string `json:"name"`
	BusinessType struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"business_type"`
	BusinessStatus struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"business_status"`
	Location struct {
		ActivityRegion struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"activity_region"`
		ActivitySubRegion struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"activity_sub_region"`
		ActivityAddress string `json:"activity_address"`
		Phone           string `json:"phone"`
		Email           string `json:"email"`
	} `json:"location"`
	OKED struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"oked"`
	Manager *struct {
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		MiddleName string `json:"middle_name"`
	} `json:"manager"`
	Founders *struct {
		FounderIndividualList []struct {
			LastName   string  `json:"last_name"`
			FirstName  string  `json:"first_name"`
			MiddleName string  `json:"middle_name"`
			Percentage string  `json:"percentage"`
			ShareAmount float64 `json:"share_amount"`
		} `json:"founder_individual_response_list"`
		TotalShareAmount float64 `json:"total_share_amount"`
	} `json:"founders"`
	OKPO               string `json:"okpo"`
	RegistrationDate   string `json:"registration_date"`
	RegisterNumber     string `json:"register_number"`
	OwnershipTypeID    int    `json:"ownership_type_id"`
	RegistrationDateFmt string `json:"registration_date_format"`
	Status             string `json:"status,omitempty"`
}

// DetailResponse is the API response wrapper for the detail endpoint.
type DetailResponse struct {
	Status  int     `json:"status"`
	Message *string `json:"message"`
	Data    Detail  `json:"data"`
}

// Client is an HTTP client for the Birdarcha API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a new Birdarcha API client.
func NewClient(baseURL, token string) *Client {
	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken updates the Bearer token.
func (c *Client) SetToken(token string) {
	c.token = token
}

// FetchList fetches a paginated list of legal entities.
func (c *Client) FetchList(ctx context.Context, page, perPage int) (*ListResponse, error) {
	url := fmt.Sprintf("%s/v1/register/legal_entity/office?lang=uz&per_page=%d&page=%d", c.baseURL, perPage, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to create list request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: list request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to read list response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("birdarcha: list returned status %d: %s", resp.StatusCode, string(body))
	}

	var result ListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("birdarcha: failed to parse list response: %w", err)
	}

	return &result, nil
}

// FetchDetail fetches the full details of a single legal entity.
func (c *Client) FetchDetail(ctx context.Context, id int) (*Detail, error) {
	url := fmt.Sprintf("%s/v1/register/legal_entity/%d/office?lang=uz", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to create detail request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: detail request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to read detail response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("birdarcha: detail returned status %d: %s", resp.StatusCode, string(body))
	}

	var result DetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("birdarcha: failed to parse detail response: %w", err)
	}

	return &result.Data, nil
}
