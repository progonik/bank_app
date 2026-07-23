package birdarcha

import (
	"bytes"
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
	ActivityRegion struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"activity_region"`
	ActivitySubRegion struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"activity_sub_region"`
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

// IndividualListItem represents a single individual entrepreneur from the list endpoint.
type IndividualListItem struct {
	ID               int    `json:"id"`
	RegistrationDate string `json:"registration_date"`
	FullName         string `json:"full_name"`
	PIN              string `json:"pin"`
	BusinessType     struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"business_type"`
	ActivityRegion struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"activity_region"`
	ActivitySubRegion struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"activity_sub_region"`
	BusinessStatus string `json:"business_status"`
}

// IndividualListResponse is the API response wrapper for the individual list endpoint.
type IndividualListResponse struct {
	Status     int                  `json:"status"`
	Data       []IndividualListItem `json:"data"`
	TotalPages int                  `json:"total_pages"`
	TotalCount int                  `json:"total_count"`
	Size       int                  `json:"size"`
	Page       int                  `json:"page"`
}

// Detail represents the full entity detail from the detail endpoint.
type Detail struct {
	ID           int    `json:"id"`
	TIN          int64  `json:"tin"`
	Name         string `json:"name"`
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
			LastName    string  `json:"last_name"`
			FirstName   string  `json:"first_name"`
			MiddleName  string  `json:"middle_name"`
			Percentage  string  `json:"percentage"`
			ShareAmount float64 `json:"share_amount"`
		} `json:"founder_individual_response_list"`
		TotalShareAmount float64 `json:"total_share_amount"`
	} `json:"founders"`
	OKPO                string `json:"okpo"`
	RegistrationDate    string `json:"registration_date"`
	RegisterNumber      string `json:"register_number"`
	OwnershipTypeID     int    `json:"ownership_type_id"`
	RegistrationDateFmt string `json:"registration_date_format"`
	Status              string `json:"status,omitempty"`
}

// IndividualDetail represents the full individual entrepreneur detail response.
type IndividualDetail struct {
	ID           int `json:"id"`
	BusinessType struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"business_type"`
	Information struct {
		PIN        string `json:"pin"`
		FirstName  string `json:"first_name"`
		LastName   string `json:"last_name"`
		MiddleName string `json:"middle_name"`
		Birthday   string `json:"birthday"`
		Gender     string `json:"gender"`
	} `json:"information"`
	Location struct {
		ActivityRegion struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"activity_region"`
		ActivitySubRegion struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"activity_sub_region"`
		Mahalla *struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"mahalla"`
		ActivityAddress string `json:"activity_address"`
		Phone           string `json:"phone"`
		Email           string `json:"email"`
	} `json:"location"`
	ActivityType struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"activity_type"`
	ActivityTypeOKED *struct {
		OKED struct {
			ID   int    `json:"id"`
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"oked"`
	} `json:"activity_type_oked"`
	BusinessStatus string `json:"business_status"`
	RegisterNumber string `json:"register_number"`
	RegisterDate   string `json:"register_date"`
}

// DetailResponse is the API response wrapper for the detail endpoint.
type DetailResponse struct {
	Status  int     `json:"status"`
	Message *string `json:"message"`
	Data    Detail  `json:"data"`
}

// IndividualDetailResponse is the API response wrapper for the individual detail endpoint.
type IndividualDetailResponse struct {
	Status  int              `json:"status"`
	Message *string          `json:"message"`
	Data    IndividualDetail `json:"data"`
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

// FetchIndividualList fetches a paginated list of individual entrepreneurs.
func (c *Client) FetchIndividualList(ctx context.Context, page, size int) (*IndividualListResponse, error) {
	url := fmt.Sprintf("%s/v1/register/individuals/office?size=%d&page=%d&lang=uz", c.baseURL, size, page)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to create individual list request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: individual list request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to read individual list response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("birdarcha: individual list returned status %d: %s", resp.StatusCode, string(body))
	}

	var result IndividualListResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("birdarcha: failed to parse individual list response: %w", err)
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

// FetchIndividualDetail fetches the full details of a single individual entrepreneur.
func (c *Client) FetchIndividualDetail(ctx context.Context, id int) (*IndividualDetail, error) {
	url := fmt.Sprintf("%s/v1/register/individuals/%d/office?lang=uz", c.baseURL, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to create individual detail request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: individual detail request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("birdarcha: failed to read individual detail response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("birdarcha: individual detail returned status %d: %s", resp.StatusCode, string(body))
	}

	var result IndividualDetailResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("birdarcha: failed to parse individual detail response: %w", err)
	}

	return &result.Data, nil
}
