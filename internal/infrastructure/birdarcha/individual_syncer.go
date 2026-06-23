package birdarcha

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	appent "github.com/prodonik/bank_app/internal/application/entrepreneur"
)

// IndividualSyncer periodically fetches new individual entrepreneurs from Birdarcha.
type IndividualSyncer struct {
	client              *Client
	entrepreneurService *appent.Service
	db                  *sql.DB
	interval            time.Duration
	cutoffDate          string
}

func NewIndividualSyncer(client *Client, entrepreneurService *appent.Service, db *sql.DB, interval time.Duration, cutoffDate string) *IndividualSyncer {
	return &IndividualSyncer{
		client:              client,
		entrepreneurService: entrepreneurService,
		db:                  db,
		interval:            interval,
		cutoffDate:          cutoffDate,
	}
}

func (s *IndividualSyncer) Start(ctx context.Context) {
	log.Printf("birdarcha-individual-syncer: starting (interval=%s, cutoff=%s)", s.interval, s.cutoffDate)

	s.runSync(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("birdarcha-individual-syncer: stopped")
			return
		case <-ticker.C:
			s.runSync(ctx)
		}
	}
}

func (s *IndividualSyncer) runSync(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("birdarcha-individual-syncer: panic recovered: %v", r)
		}
	}()

	if err := s.sync(ctx); err != nil {
		log.Printf("birdarcha-individual-syncer: sync error: %v", err)
	}
}

func (s *IndividualSyncer) sync(ctx context.Context) error {
	token, err := s.getToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get birdarcha token: %w", err)
	}
	if token == "" {
		log.Println("birdarcha-individual-syncer: no token in DB, skipping sync")
		return nil
	}
	s.client.SetToken(token)

	lastID, err := s.getLastStoredID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last stored id: %w", err)
	}

	log.Printf("birdarcha-individual-syncer: starting sync (last_stored_id=%d)", lastID)

	if lastID == 0 {
		list, err := s.client.FetchIndividualList(ctx, 0, 1)
		if err != nil {
			return fmt.Errorf("failed to bootstrap last_stored_id: %w", err)
		}
		if len(list.Data) > 0 {
			if err := s.setLastStoredID(ctx, list.Data[0].ID); err != nil {
				return fmt.Errorf("failed to set initial last_stored_id: %w", err)
			}
			log.Printf("birdarcha-individual-syncer: bootstrapped last_stored_id=%d, will sync new data from next cycle", list.Data[0].ID)
		}
		return nil
	}

	var newItems []IndividualListItem
	done := false

	for page := 0; !done; page++ {
		list, err := s.client.FetchIndividualList(ctx, page, 20)
		if err != nil {
			if strings.Contains(err.Error(), "status 401") {
				log.Println("birdarcha-individual-syncer: token expired, signalling extension")
				_ = s.setTokenRefreshNeeded(ctx, true)
			}
			return fmt.Errorf("failed to fetch list page %d: %w", page, err)
		}

		if len(list.Data) == 0 {
			break
		}

		for _, item := range list.Data {
			if lastID > 0 && item.ID <= lastID {
				done = true
				break
			}
			if s.isBeforeCutoff(item.RegistrationDate) {
				done = true
				break
			}
			newItems = append(newItems, item)
		}
	}

	if len(newItems) == 0 {
		log.Println("birdarcha-individual-syncer: no new entities to sync")
		return nil
	}

	log.Printf("birdarcha-individual-syncer: found %d new entities, processing...", len(newItems))

	created := 0
	skipped := 0
	failed := 0
	lastSuccessID := lastID
	hitFailure := false

	for i := len(newItems) - 1; i >= 0; i-- {
		item := newItems[i]

		detail, err := s.client.FetchIndividualDetail(ctx, item.ID)
		if err != nil {
			log.Printf("birdarcha-individual-syncer: failed to fetch detail for id=%d pin=%s: %v", item.ID, item.PIN, err)
			failed++
			hitFailure = true
			continue
		}

		input := s.mapToCreateInput(item, detail)

		_, err = s.entrepreneurService.Create(ctx, input)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
				log.Printf("birdarcha-individual-syncer: skipped duplicate pin=%s name=%s", item.PIN, item.FullName)
				skipped++
			} else {
				log.Printf("birdarcha-individual-syncer: failed to create pin=%s name=%s: %v", item.PIN, item.FullName, err)
				failed++
				hitFailure = true
				continue
			}
		} else {
			created++
		}

		if !hitFailure {
			lastSuccessID = item.ID
		}
	}

	if lastSuccessID > lastID {
		if err := s.setLastStoredID(ctx, lastSuccessID); err != nil {
			log.Printf("birdarcha-individual-syncer: failed to update last_stored_id: %v", err)
		}
	}

	log.Printf("birdarcha-individual-syncer: sync complete — created=%d, skipped=%d, failed=%d", created, skipped, failed)
	return nil
}

func (s *IndividualSyncer) getToken(ctx context.Context) (string, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM syncer_state WHERE key = 'birdarcha_token'`,
	).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func (s *IndividualSyncer) setTokenRefreshNeeded(ctx context.Context, needed bool) error {
	val := "false"
	if needed {
		val = "true"
	}
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO syncer_state (key, value, updated_at)
		VALUES ('birdarcha_token_refresh_needed', $1, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = NOW()
	`, val)
	return err
}

func (s *IndividualSyncer) getLastStoredID(ctx context.Context) (int, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM syncer_state WHERE key = 'birdarcha_individual_last_id'`,
	).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid individual last_id value %q: %w", value, err)
	}
	return id, nil
}

func (s *IndividualSyncer) setLastStoredID(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO syncer_state (key, value, updated_at)
		VALUES ('birdarcha_individual_last_id', $1, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = NOW()
	`, strconv.Itoa(id))
	return err
}

func (s *IndividualSyncer) isBeforeCutoff(dateStr string) bool {
	date, err := time.Parse("02.01.2006", dateStr)
	if err != nil {
		return false
	}
	cutoff, err := time.Parse("02.01.2006", s.cutoffDate)
	if err != nil {
		return false
	}
	return date.Before(cutoff)
}

func (s *IndividualSyncer) mapToCreateInput(item IndividualListItem, detail *IndividualDetail) appent.CreateInput {
	fullName := individualFullName(detail)
	if fullName == "" {
		fullName = item.FullName
	}

	pin := detail.Information.PIN
	if pin == "" {
		pin = item.PIN
	}

	regDate := detail.RegisterDate
	if regDate == "" {
		regDate = convertDate(item.RegistrationDate)
	}

	ifutCode := ""
	if detail.ActivityTypeOKED != nil {
		ifutCode = detail.ActivityTypeOKED.OKED.Code
	}

	legalFormID := detail.BusinessType.ID
	if legalFormID == 0 {
		legalFormID = item.BusinessType.ID
	}

	activityStatus := detail.BusinessStatus == "ACTIVE"
	if detail.BusinessStatus == "" {
		activityStatus = item.BusinessStatus == "ACTIVE"
	}

	return appent.CreateInput{
		InnName:               pin,
		LegalName:             fullName,
		RegistrationAuthority: "birdarcha_individual",
		RegistrationDate:      regDate,
		RegistrationNumber:    detail.RegisterNumber,
		LegalForm:             strconv.Itoa(legalFormID),
		IfutCode:              ifutCode,
		DbibtCode:             0,
		ActivityStatus:        activityStatus,
		CharterFund:           0,
		Founders:              "",
		Email:                 detail.Location.Email,
		Phone:                 detail.Location.Phone,
		MhobtCode:             "",
		Address:               detail.Location.ActivityAddress,
		DirectorName:          fullName,
	}
}

func individualFullName(detail *IndividualDetail) string {
	if detail == nil {
		return ""
	}
	parts := []string{}
	if detail.Information.LastName != "" {
		parts = append(parts, detail.Information.LastName)
	}
	if detail.Information.FirstName != "" {
		parts = append(parts, detail.Information.FirstName)
	}
	if detail.Information.MiddleName != "" {
		parts = append(parts, detail.Information.MiddleName)
	}
	return strings.Join(parts, " ")
}
