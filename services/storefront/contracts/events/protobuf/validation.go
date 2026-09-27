package protobuf

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/contracts/events/model"
	a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
	d "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/domain"
)

var offerCode = regexp.MustCompile(`^[A-Z][A-Z0-9]{0,7}$`)
var collectionCode = regexp.MustCompile(`^[A-Z0-9]{6}$`)

// Transport validation protects the application before any handler is invoked.
// It validates published evidence; it does not reproduce mutable owner policy.
func validatePayload(m a.Message) error {
	var ids []string
	sourceID := ""
	switch p := m.Payload.(type) {
	case model.DrinkPublished:
		ids = []string{p.DrinkID}
		sourceID = p.DrinkID
		if strings.TrimSpace(p.Name) == "" || !utf8.ValidString(p.Name) || utf8.RuneCountInString(p.Name) > 80 || p.Revision == 0 {
			return fmt.Errorf("invalid published drink")
		}
	case model.MenuPublished:
		ids = []string{p.EditionID}
		sourceID = p.EditionID
		if len(p.Offers) == 0 || len(p.Offers) > 20 {
			return fmt.Errorf("invalid published menu size")
		}
		seen := map[string]bool{}
		for _, offer := range p.Offers {
			ids = append(ids, offer.DrinkID)
			if seen[offer.Code] || !offerCode.MatchString(offer.Code) ||
				offer.DrinkRevision == 0 || offer.Name == "" || offer.Currency != p.Currency {
				return fmt.Errorf("invalid published menu offer")
			}
			if _, err := d.NewMoney(offer.Minor, offer.Currency); err != nil {
				return err
			}
			seen[offer.Code] = true
		}
	case model.OrderPlaced:
		ids = []string{p.OrderID, p.CustomerID, p.EditionID}
		sourceID = p.OrderID
		total := 0
		seen := map[string]bool{}
		for _, line := range p.Lines {
			ids = append(ids, line.ID)
			if seen[line.ID] || !offerCode.MatchString(line.OfferCode) || line.Name == "" {
				return fmt.Errorf("invalid accepted order line")
			}
			if _, err := d.NewQuantity(line.Quantity); err != nil {
				return err
			}
			if _, err := d.NewMoney(line.Minor, p.Currency); err != nil {
				return err
			}
			total += line.Quantity
			seen[line.ID] = true
		}
		if total < 1 || total > 5 {
			return fmt.Errorf("invalid accepted order quantity")
		}
	case model.DrinksReady:
		ids = []string{p.OrderID, p.CustomerID}
	case model.PickupOpened:
		ids = []string{p.PickupID, p.OrderID, p.CustomerID}
		sourceID = p.PickupID
		if !collectionCode.MatchString(p.CollectionCode) {
			return fmt.Errorf("invalid collection code")
		}
	case model.OrderCollected:
		ids = []string{p.OrderID, p.CustomerID}
	case model.RewardEarned:
		ids = []string{p.AccountID, p.GrantID}
		sourceID = p.AccountID
		if p.Benefit == "" || p.ValidDays < 1 || p.ValidDays > 30 {
			return fmt.Errorf("invalid earned grant terms")
		}
	case model.RewardIssued:
		ids = []string{p.RewardID, p.CustomerID}
		sourceID = p.RewardID
		if _, err := time.Parse(time.RFC3339Nano, p.ExpiresAt); err != nil || p.Benefit == "" {
			return fmt.Errorf("invalid issued reward")
		}
	case model.NotificationRequested:
		ids = []string{p.NotificationID}
		sourceID = p.NotificationID
	default:
		return fmt.Errorf("unknown event payload")
	}
	for _, id := range ids {
		if err := d.ValidateID(id); err != nil {
			return err
		}
	}
	if sourceID != "" && sourceID != m.AggregateID {
		return fmt.Errorf("payload does not identify its source aggregate")
	}
	return nil
}
