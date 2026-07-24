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

const tashkentCityRegionID = 1726

// Syncer periodically fetches new legal entities from Birdarcha and creates
// them via the entrepreneur service (which also sends to SQB).
type Syncer struct {
	client              *Client
	entrepreneurService *appent.Service
	db                  *sql.DB
	interval            time.Duration
	cutoffDate          string // "dd.MM.yyyy" format — stop syncing before this date
}

// NewSyncer creates a new Birdarcha syncer.
func NewSyncer(client *Client, entrepreneurService *appent.Service, db *sql.DB, interval time.Duration, cutoffDate string) *Syncer {
	return &Syncer{
		client:              client,
		entrepreneurService: entrepreneurService,
		db:                  db,
		interval:            interval,
		cutoffDate:          cutoffDate,
	}
}

// GetToken returns the current birdarcha token from the database.
func (s *Syncer) GetToken(ctx context.Context) (string, error) {
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

// SetToken stores the birdarcha token in the database.
func (s *Syncer) SetToken(ctx context.Context, token string) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO syncer_state (key, value, updated_at)
		VALUES ('birdarcha_token', $1, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = NOW()
	`, token)
	return err
}

// Start runs the sync loop in the foreground. Launch with `go syncer.Start(ctx)`.
func (s *Syncer) Start(ctx context.Context) {
	log.Printf("birdarcha-syncer: starting (interval=%s, cutoff=%s)", s.interval, s.cutoffDate)

	// Run once immediately on startup
	s.runSync(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("birdarcha-syncer: stopped")
			return
		case <-ticker.C:
			s.runSync(ctx)
		}
	}
}

func (s *Syncer) runSync(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("birdarcha-syncer: panic recovered: %v", r)
		}
	}()

	if err := s.sync(ctx); err != nil {
		log.Printf("birdarcha-syncer: sync error: %v", err)
	}
}

func (s *Syncer) setTokenRefreshNeeded(ctx context.Context, needed bool) error {
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

func (s *Syncer) sync(ctx context.Context) error {
	// Load token from DB before each sync
	token, err := s.GetToken(ctx)
	if err != nil {
		return fmt.Errorf("failed to get birdarcha token: %w", err)
	}
	if token == "" {
		log.Println("birdarcha-syncer: no token in DB, skipping sync")
		return nil
	}
	s.client.SetToken(token)

	lastID, err := s.getLastStoredID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get last stored id: %w", err)
	}

	log.Printf("birdarcha-syncer: starting sync (last_stored_id=%d)", lastID)

	// First run: just grab the newest ID to bootstrap the cursor
	if lastID == 0 {
		list, err := s.client.FetchList(ctx, 0, 1)
		if err != nil {
			return fmt.Errorf("failed to bootstrap last_stored_id: %w", err)
		}
		if len(list.Data) > 0 {
			if err := s.setLastStoredID(ctx, list.Data[0].ID); err != nil {
				return fmt.Errorf("failed to set initial last_stored_id: %w", err)
			}
			log.Printf("birdarcha-syncer: bootstrapped last_stored_id=%d, will sync new data from next cycle", list.Data[0].ID)
		}
		return nil
	}

	// Collect all new items first (they come newest-first from the API)
	var newItems []ListItem
	done := false

	for page := 0; !done; page++ {
		list, err := s.client.FetchList(ctx, page, 20)
		if err != nil {
			if strings.Contains(err.Error(), "status 401") {
				log.Println("birdarcha-syncer: token expired, signalling extension")
				_ = s.setTokenRefreshNeeded(ctx, true)
			}
			return fmt.Errorf("failed to fetch list page %d: %w", page, err)
		}

		if len(list.Data) == 0 {
			break
		}

		for _, item := range list.Data {
			// Stop if we've reached the last stored ID
			if lastID > 0 && item.ID <= lastID {
				done = true
				break
			}

			// Stop if registration date is before cutoff
			if s.isBeforeCutoff(item.RegistrationDate) {
				done = true
				break
			}

			newItems = append(newItems, item)
		}
	}

	if len(newItems) == 0 {
		log.Println("birdarcha-syncer: no new entities to sync")
		return nil
	}

	log.Printf("birdarcha-syncer: found %d new entities, processing...", len(newItems))

	// Process oldest first (reverse order)
	// Track the last contiguously successful ID — stop advancing on first failure
	// so failed items are retried next cycle
	created := 0
	skipped := 0
	failed := 0
	lastSuccessID := lastID
	hitFailure := false

	for i := len(newItems) - 1; i >= 0; i-- {
		item := newItems[i]

		if item.ActivityRegion.ID != tashkentCityRegionID {
			log.Printf("birdarcha-syncer: skipped non-Tashkent-city tin=%d name=%s region_id=%d region=%s", item.TIN, item.Name, item.ActivityRegion.ID, item.ActivityRegion.Name)
			skipped++
			if !hitFailure {
				lastSuccessID = item.ID
			}
			continue
		}

		// Fetch full details
		detail, err := s.client.FetchDetail(ctx, item.ID)
		if err != nil {
			log.Printf("birdarcha-syncer: failed to fetch detail for id=%d tin=%d: %v", item.ID, item.TIN, err)
			failed++
			hitFailure = true
			continue
		}

		// Map to CreateInput
		input := s.mapToCreateInput(item, detail)

		// Create via entrepreneur service (also sends to SQB)
		_, err = s.entrepreneurService.Create(ctx, input)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
				log.Printf("birdarcha-syncer: skipped duplicate tin=%d name=%s", item.TIN, item.Name)
				skipped++
			} else {
				log.Printf("birdarcha-syncer: failed to create tin=%d name=%s: %v", item.TIN, item.Name, err)
				failed++
				hitFailure = true
				continue
			}
		} else {
			created++
		}

		// Only advance last_stored_id if we haven't hit any failure yet
		if !hitFailure {
			lastSuccessID = item.ID
		}
	}

	// Update last_stored_id only up to the last contiguous success
	if lastSuccessID > lastID {
		if err := s.setLastStoredID(ctx, lastSuccessID); err != nil {
			log.Printf("birdarcha-syncer: failed to update last_stored_id: %v", err)
		}
	}

	log.Printf("birdarcha-syncer: sync complete — created=%d, skipped=%d, failed=%d", created, skipped, failed)
	return nil
}

func (s *Syncer) mapToCreateInput(item ListItem, detail *Detail) appent.CreateInput {
	inn := strconv.FormatInt(item.TIN, 10)

	// Director name from manager
	directorName := ""
	if detail.Manager != nil {
		parts := []string{}
		if detail.Manager.LastName != "" {
			parts = append(parts, detail.Manager.LastName)
		}
		if detail.Manager.FirstName != "" {
			parts = append(parts, detail.Manager.FirstName)
		}
		if detail.Manager.MiddleName != "" {
			parts = append(parts, detail.Manager.MiddleName)
		}
		directorName = strings.Join(parts, " ")
	}

	// Founders as comma-separated names
	founders := ""
	if detail.Founders != nil {
		var names []string
		for _, f := range detail.Founders.FounderIndividualList {
			name := strings.TrimSpace(f.LastName + " " + f.FirstName + " " + f.MiddleName)
			if name != "" {
				names = append(names, name)
			}
		}
		founders = strings.Join(names, ", ")
	}

	// Charter fund
	var charterFund int32
	if detail.Founders != nil {
		charterFund = int32(detail.Founders.TotalShareAmount)
	}

	// Registration date: convert "dd.MM.yyyy" to "yyyy-MM-dd"
	regDate := convertDate(item.RegistrationDate)

	// Legal form: use business_type ID as string (e.g. "152")
	legalForm := strconv.Itoa(item.BusinessType.ID)

	// OKED/IFUT code
	ifutCode := detail.OKED.ID

	// Phone and email from detail
	phone := detail.Location.Phone
	email := detail.Location.Email

	// Activity status
	activityStatus := item.Status == "ACTIVE"

	// Address
	address := detail.Location.ActivityAddress
	activitySubRegion := item.ActivitySubRegion.Name
	if detail.Location.ActivitySubRegion.Name != "" {
		activitySubRegion = detail.Location.ActivitySubRegion.Name
	}

	// Registration number
	regNumber := detail.RegisterNumber

	return appent.CreateInput{
		InnName:               inn,
		LegalName:             item.Name,
		RegistrationAuthority: "birdarcha",
		RegistrationDate:      regDate,
		RegistrationNumber:    regNumber,
		LegalForm:             legalForm,
		IfutCode:              ifutCode,
		DbibtCode:             0,
		ActivityStatus:        activityStatus,
		CharterFund:           charterFund,
		Founders:              founders,
		Email:                 email,
		Phone:                 phone,
		MhobtCode:             "",
		Address:               address,
		ActivitySubRegion:     activitySubRegion,
		DirectorName:          directorName,
	}
}

// convertDate converts "dd.MM.yyyy" to "yyyy-MM-dd".
func convertDate(d string) string {
	parts := strings.Split(d, ".")
	if len(parts) == 3 {
		return parts[2] + "-" + parts[1] + "-" + parts[0]
	}
	return d
}

// isBeforeCutoff checks if a "dd.MM.yyyy" date is before the cutoff date.
func (s *Syncer) isBeforeCutoff(dateStr string) bool {
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

// getLastStoredID returns the last synced birdarcha entity ID, or 0 if none.
func (s *Syncer) getLastStoredID(ctx context.Context) (int, error) {
	var value string
	err := s.db.QueryRowContext(ctx,
		`SELECT value FROM syncer_state WHERE key = 'birdarcha_last_id'`,
	).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, err
	}
	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid last_id value %q: %w", value, err)
	}
	return id, nil
}

// setLastStoredID upserts the last synced birdarcha entity ID.
func (s *Syncer) setLastStoredID(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO syncer_state (key, value, updated_at)
		VALUES ('birdarcha_last_id', $1, NOW())
		ON CONFLICT (key) DO UPDATE SET value = $1, updated_at = NOW()
	`, strconv.Itoa(id))
	return err
}
