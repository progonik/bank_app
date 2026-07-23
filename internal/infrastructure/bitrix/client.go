package bitrix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	domain "github.com/prodonik/bank_app/internal/domain/entrepreneur"
)

type Client struct {
	webhookURL string
	httpClient *http.Client
}

type response struct {
	Result           int    `json:"result"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func NewClient(webhookURL string) *Client {
	webhookURL = strings.TrimSpace(webhookURL)
	if webhookURL != "" && !strings.HasSuffix(webhookURL, "/") {
		webhookURL += "/"
	}
	return &Client{
		webhookURL: webhookURL,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c != nil && c.webhookURL != ""
}

func (c *Client) CreateContactAndDeal(ctx context.Context, e *domain.Entrepreneur) (contactID, dealID int, err error) {
	if !c.Enabled() {
		return 0, 0, nil
	}

	contactID, err = c.CreateContact(ctx, e)
	if err != nil {
		return 0, 0, err
	}

	dealID, err = c.CreateDeal(ctx, e, contactID)
	if err != nil {
		return contactID, 0, err
	}

	return contactID, dealID, nil
}

func (c *Client) CreateContact(ctx context.Context, e *domain.Entrepreneur) (int, error) {
	fields := map[string]any{
		"NAME":     e.LegalName,
		"COMMENTS": comments(e),
	}
	if e.Phone != "" {
		fields["PHONE"] = []map[string]string{{"VALUE": e.Phone, "VALUE_TYPE": "WORK"}}
	}
	if e.Email != "" {
		fields["EMAIL"] = []map[string]string{{"VALUE": e.Email, "VALUE_TYPE": "WORK"}}
	}

	id, err := c.call(ctx, "crm.contact.add.json", map[string]any{"fields": fields})
	if err != nil {
		return 0, fmt.Errorf("bitrix: create contact failed: %w", err)
	}
	return id, nil
}

func (c *Client) CreateDeal(ctx context.Context, e *domain.Entrepreneur, contactID int) (int, error) {
	fields := map[string]any{
		"TITLE":    e.LegalName,
		"COMMENTS": comments(e),
	}
	if contactID > 0 {
		fields["CONTACT_ID"] = contactID
	}

	id, err := c.call(ctx, "crm.deal.add.json", map[string]any{"fields": fields})
	if err != nil {
		return 0, fmt.Errorf("bitrix: create deal failed: %w", err)
	}
	return id, nil
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

func comments(e *domain.Entrepreneur) string {
	lines := []string{
		"Source: Birdarcha",
		"INN/PIN: " + e.InnName,
		"Registration date: " + e.RegistrationDate,
		"Registration number: " + e.RegistrationNumber,
		"Org form: " + e.LegalForm,
		"OKED: " + e.IfutCodeName,
		"Phone: " + e.Phone,
		"Email: " + e.Email,
		"Address: " + e.Address,
		"Director/chief: " + e.DirectorName,
	}
	return strings.Join(lines, "\n")
}
