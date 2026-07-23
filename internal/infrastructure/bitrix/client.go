package bitrix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	domain "github.com/prodonik/bank_app/internal/domain/entrepreneur"
)

type Client struct {
	webhookURL string
	httpClient *http.Client
	location   *time.Location
}

type response struct {
	Result           int    `json:"result"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

type userListResponse struct {
	Result           []user `json:"result"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
	Next             *int   `json:"next"`
}

type user struct {
	ID     string `json:"ID"`
	Active bool   `json:"ACTIVE"`
}

const bitrixResponsibleUserID = 31

func NewClient(webhookURL string) *Client {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL != "" && !strings.HasSuffix(webhookURL, "/") {
		webhookURL += "/"
	}

	location, err := time.LoadLocation("Asia/Tashkent")
	if err != nil {
		location = time.FixedZone("Asia/Tashkent", 5*60*60)
	}

	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
		location:   location,
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.webhookURL != ""
}

func (c *Client) CreateTask(ctx context.Context, e *domain.Entrepreneur) (int, error) {
	if !c.Enabled() {
		return 0, nil
	}

	auditorIDs, err := c.fetchAuditorUserIDs(ctx)
	if err != nil {
		return 0, err
	}

	payload := map[string]any{
		"TASKDATA": map[string]any{
			"TITLE":          e.LegalName,
			"DESCRIPTION":    description(e),
			"DEADLINE":       c.todayDeadline(),
			"RESPONSIBLE_ID": bitrixResponsibleUserID,
			"AUDITORS":       auditorIDs,
		},
	}

	id, err := c.call(ctx, "task.item.add.json", payload)
	if err != nil {
		return 0, fmt.Errorf("bitrix: create task failed: %w", err)
	}
	return id, nil
}

func (c *Client) fetchAuditorUserIDs(ctx context.Context) ([]int, error) {
	users, err := c.fetchActiveUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("bitrix: fetch users failed: %w", err)
	}

	auditorIDs := make([]int, 0, len(users))
	for _, u := range users {
		id, err := strconv.Atoi(u.ID)
		if err != nil {
			return nil, fmt.Errorf("invalid user id %q: %w", u.ID, err)
		}
		if id != bitrixResponsibleUserID {
			auditorIDs = append(auditorIDs, id)
		}
	}

	return auditorIDs, nil
}

func (c *Client) fetchActiveUsers(ctx context.Context) ([]user, error) {
	var activeUsers []user
	start := 0

	for {
		result, err := c.fetchUsersPage(ctx, start)
		if err != nil {
			return nil, err
		}

		for _, u := range result.Result {
			if u.Active {
				activeUsers = append(activeUsers, u)
			}
		}

		if result.Next == nil {
			break
		}
		start = *result.Next
	}

	return activeUsers, nil
}

func (c *Client) fetchUsersPage(ctx context.Context, start int) (*userListResponse, error) {
	url := fmt.Sprintf("%suser.get.json?start=%d", c.webhookURL, start)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var result userListResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response status=%d body=%s: %w", resp.StatusCode, string(respBody), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status %d: %s - %s", resp.StatusCode, result.Error, result.ErrorDescription)
	}
	if result.Error != "" {
		return nil, fmt.Errorf("%s - %s", result.Error, result.ErrorDescription)
	}

	return &result, nil
}

func (c *Client) call(ctx context.Context, method string, payload any) (int, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return 0, fmt.Errorf("marshal request: %w", err)
	}

	url := c.webhookURL + method
	log.Printf("bitrix: calling %s: %s", method, string(body))

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read response: %w", err)
	}
	log.Printf("bitrix: response status=%d body=%s", resp.StatusCode, string(respBody))

	var result response
	if err := json.Unmarshal(respBody, &result); err != nil {
		return 0, fmt.Errorf("decode response status=%d body=%s: %w", resp.StatusCode, string(respBody), err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("unexpected status %d: %s - %s", resp.StatusCode, result.Error, result.ErrorDescription)
	}
	if result.Error != "" {
		return 0, fmt.Errorf("%s - %s", result.Error, result.ErrorDescription)
	}
	return result.Result, nil
}

func (c *Client) todayDeadline() string {
	now := time.Now().In(c.location)
	deadline := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 0, 0, c.location)
	return deadline.Format(time.RFC3339)
}

func description(e *domain.Entrepreneur) string {
	lines := []string{
		"INN/PIN: " + e.InnName,
		"Name: " + e.LegalName,
		"Registration date: " + e.RegistrationDate,
		"Registration number: " + e.RegistrationNumber,
		"Legal form: " + e.LegalForm,
		"OKED / IFUT: " + e.IfutCodeName,
		"Activity status: " + activityStatus(e.ActivityStatus),
		"Charter fund: " + fmt.Sprintf("%d", e.CharterFund),
		"Founders: " + e.Founders,
		"Email: " + e.Email,
		"Phone: " + e.Phone,
		"MHOBT code: " + e.MhobtCode,
		"Address: " + e.Address,
		"Director / chief: " + e.DirectorName,
	}
	return strings.Join(lines, "\n")
}

func activityStatus(active bool) string {
	if active {
		return "active"
	}
	return "inactive"
}
