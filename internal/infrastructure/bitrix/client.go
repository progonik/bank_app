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

const bitrixResponsibleUserID = 31

var bitrixObserverUserIDs = []int{1153}

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

func (c *Client) CreateLead(ctx context.Context, e *domain.Entrepreneur) (int, error) {
	if !c.Enabled() {
		return 0, nil
	}

	id, err := c.call(ctx, "crm.lead.add.json", map[string]any{"fields": leadFields(e)})
	if err != nil {
		return 0, fmt.Errorf("bitrix: create lead failed: %w", err)
	}
	if err := c.addLeadObservers(ctx, id); err != nil {
		return id, fmt.Errorf("bitrix: add lead observers failed for lead=%d: %w", id, err)
	}
	return id, nil
}

func (c *Client) addLeadObservers(ctx context.Context, leadID int) error {
	if len(bitrixObserverUserIDs) == 0 {
		return nil
	}

	_, err := c.call(ctx, "crm.item.update.json", map[string]any{
		"entityTypeId": 1,
		"id":           leadID,
		"fields": map[string]any{
			"observers": bitrixObserverUserIDs,
		},
	})
	return err
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

func leadFields(e *domain.Entrepreneur) map[string]any {
	fields := map[string]any{
		"TITLE":                taskTitle(e),
		"COMPANY_TITLE":        taskTitle(e),
		"NAME":                 e.DirectorName,
		"STATUS_ID":            "NEW",
		"ASSIGNED_BY_ID":       bitrixResponsibleUserID,
		"ADDRESS":              e.Address,
		"ADDRESS_2":            e.ActivitySubRegion,
		"COMMENTS":             description(e),
		"UF_CRM_1638948461838": e.InnName,
		"UF_CRM_UZB_INN_LEAD":  e.InnName,
		"UF_CRM_1638948478586": e.IfutCodeName,
	}
	if e.Phone != "" {
		fields["PHONE"] = []map[string]string{{"VALUE": e.Phone, "VALUE_TYPE": "WORK"}}
	}
	if e.Email != "" {
		fields["EMAIL"] = []map[string]string{{"VALUE": e.Email, "VALUE_TYPE": "WORK"}}
	}
	return fields
}

func description(e *domain.Entrepreneur) string {
	lines := []string{
		"INN/PIN: " + e.InnName,
		"Name: " + taskTitle(e),
		"Registration date: " + e.RegistrationDate,
		"Registration number: " + e.RegistrationNumber,
		"Legal form: " + legalFormDisplay(e.LegalForm),
		"OKED / IFUT: " + e.IfutCodeName,
		"Activity status: " + activityStatus(e.ActivityStatus),
		"Charter fund: " + fmt.Sprintf("%d", e.CharterFund),
		"Founders: " + e.Founders,
		"Email: " + e.Email,
		"Phone: " + e.Phone,
		"MHOBT code: " + e.MhobtCode,
		"Activity sub-region: " + e.ActivitySubRegion,
		"Address: " + e.Address,
		"Director / chief: " + e.DirectorName,
	}
	return strings.Join(lines, "\n")
}

func taskTitle(e *domain.Entrepreneur) string {
	form := legalFormName(e.LegalForm)
	name := strings.TrimSpace(e.LegalName)
	if form == "" || name == "" || hasLegalFormPrefix(name) {
		return name
	}
	return form + " " + name
}

func legalFormDisplay(id string) string {
	name := legalFormName(id)
	if name == "" {
		return id
	}
	return name + " (" + id + ")"
}

func legalFormName(id string) string {
	switch strings.TrimSpace(id) {
	case "110":
		return "XK"
	case "120":
		return "OK"
	case "130":
		return "FX"
	case "1521":
		return "MCHJ"
	default:
		return ""
	}
}

func hasLegalFormPrefix(name string) bool {
	upper := strings.ToUpper(strings.TrimSpace(name))
	for _, prefix := range []string{"MCHJ ", "XK ", "OK ", "FX "} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return false
}

func activityStatus(active bool) string {
	if active {
		return "active"
	}
	return "inactive"
}
