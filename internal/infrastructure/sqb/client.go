package sqb

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"time"

	domain "github.com/prodonik/bank_app/internal/domain/entrepreneur"
)

type LeadRequest struct {
	ID        string `json:"id"`
	INN       string `json:"inn"`
	RegDate   string `json:"regDate"`
	OrgName   string `json:"orgName"`
	OrgForm   string `json:"orgForm"`
	EcoOrg    string `json:"ecoOrg"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	AdmCode   string `json:"admCode"`
	ChiefName string `json:"chiefName"`
}

type LeadResponse struct {
	Status           string `json:"status"`
	ErrorCode        string `json:"errorCode"`
	ErrorDescription string `json:"errorDescription"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	localAddr  string
}

func NewClient(baseURL, localAddr string) *Client {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if localAddr != "" {
		dialer.LocalAddr = &net.TCPAddr{
			IP: net.ParseIP(localAddr),
		}
	}

	transport := &http.Transport{
		DialContext:       dialer.DialContext,
		TLSClientConfig:  &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: false,
	}

	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   15 * time.Second,
		},
		localAddr: localAddr,
	}
}

func toISO8601(dateStr string) string {
	formats := []string{
		"2006-01-02",
		"02.01.2006",
		"01/02/2006",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		time.RFC3339,
		time.RFC3339Nano,
	}
	for _, f := range formats {
		if t, err := time.Parse(f, dateStr); err == nil {
			return t.Format(time.RFC3339Nano)
		}
	}
	return dateStr
}

var nonDigitRe = regexp.MustCompile(`[^\d]`)

// sanitizePhone strips all non-digit characters except a leading '+'.
func sanitizePhone(phone string) string {
	if len(phone) > 0 && phone[0] == '+' {
		return "+" + nonDigitRe.ReplaceAllString(phone[1:], "")
	}
	return nonDigitRe.ReplaceAllString(phone, "")
}

func (c *Client) SendLead(ctx context.Context, e *domain.Entrepreneur) (*LeadResponse, error) {
	req := LeadRequest{
		ID:        e.ID.String(),
		INN:       e.InnName,
		RegDate:   toISO8601(e.RegistrationDate),
		OrgName:   e.LegalName,
		OrgForm:   e.LegalForm,
		EcoOrg:    e.IfutCodeName,
		Email:     e.Email,
		Phone:     sanitizePhone(e.Phone),
		AdmCode:   e.MhobtCode,
		ChiefName: e.DirectorName,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("sqb: failed to marshal lead: %w", err)
	}

	log.Printf("sqb: sending lead request: %s", string(body))

	url := c.baseURL + "/api/v1/leads/gov"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("sqb: failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("sqb: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("sqb: failed to read response body (status %d): %w", resp.StatusCode, err)
	}

	log.Printf("sqb: response status=%d body=%s", resp.StatusCode, string(respBody))

	var leadResp LeadResponse
	if err := json.Unmarshal(respBody, &leadResp); err != nil {
		return nil, fmt.Errorf("sqb: failed to decode response (status %d, body: %s): %w", resp.StatusCode, string(respBody), err)
	}

	if resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusOK {
		return &leadResp, nil
	}

	return &leadResp, fmt.Errorf("sqb: unexpected status %d: %s - %s", resp.StatusCode, leadResp.ErrorCode, leadResp.ErrorDescription)
}
