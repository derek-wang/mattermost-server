package model

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	DEALER_NAME_MAX_LENGTH = 64
	DEALER_NAME_MIN_LENGTH = 1
)

type Dealer struct {
	Id          string      `json:"id"`
	CreateAt    int64       `json:"create_at,omitempty"`
	UpdateAt    int64       `json:"update_at,omitempty"`
	DeleteAt    int64       `json:"delete_at"`
	Name        string      `json:"name"`
	PhoneNumber string      `json:"phone_number"`
	Address     string      `json:"address"`
	City        string      `json:"city"`
	Province    string      `json:"province"`
	Country     string      `json:"country"`
	PostalCode  string      `json:"postal_code"`
	Brands      StringArray `json:"brands"`
}

type DealerUpdate struct {
	Old *Dealer
	New *Dealer
}

type DealerPatch struct {
	Name        *string      `json:"name"`
	PhoneNumber *string      `json:"phone_number"`
	Address     *string      `json:"address"`
	City        *string      `json:"city"`
	Province    *string      `json:"province"`
	Country     *string      `json:"country"`
	PostalCode  *string      `json:"postal_code"`
	Brands      *StringArray `json:"brands"`
}

// IsValid validates the dealer and returns an error if it isn't configured
// correctly.
func (d *Dealer) IsValid() *AppError {

	if !IsValidId(d.Id) {
		return InvalidDealerError("id", "")
	}

	if d.CreateAt == 0 {
		return InvalidDealerError("create_at", d.Id)
	}

	if d.UpdateAt == 0 {
		return InvalidDealerError("update_at", d.Id)
	}

	if !IsValidDealerName(d.Name) {
		return InvalidDealerError("name", d.Id)
	}

	if len(d.Address) == 0 {
		return InvalidDealerError("address", d.Id)
	}

	if len(d.City) == 0 {
		return InvalidDealerError("city", d.Id)
	}

	if len(d.Province) == 0 {
		return InvalidDealerError("province", d.Id)
	}

	if len(d.Country) == 0 {
		return InvalidDealerError("country", d.Id)
	}

	if len(d.PostalCode) == 0 {
		return InvalidDealerError("postal_code", d.Id)
	}

	if len(d.PhoneNumber) == 0 {
		return InvalidDealerError("phone_number", d.Id)
	}

	if len(d.Brands) == 0 {
		return InvalidDealerError("brands", d.Id)
	}

	return nil
}

// PreSave will set the Id if missing. It will also fill
// in the CreateAt, UpdateAt times.  It should
// be run before saving the dealer to the db.
func (d *Dealer) PreSave() {
	if d.Id == "" {
		d.Id = NewId()
	}

	d.Name = SanitizeUnicode(d.Name)
	d.City = SanitizeUnicode(d.City)
	d.Address = SanitizeUnicode(d.Address)
	d.PostalCode = SanitizeUnicode(d.PostalCode)

	d.CreateAt = GetMillis()
	d.UpdateAt = d.CreateAt

	if d.Brands == nil {
		d.Brands = []string{}
	}

	d.Brands = RemoveDuplicateStrings(d.Brands)
}

// PreUpdate should be run before updating the dealer in the db.
func (d *Dealer) PreUpdate() {
	d.Name = SanitizeUnicode(d.Name)
	d.City = SanitizeUnicode(d.City)
	d.Address = SanitizeUnicode(d.Address)
	d.PostalCode = SanitizeUnicode(d.PostalCode)

	d.UpdateAt = GetMillis()

	if d.Brands == nil {
		d.Brands = []string{}
	}

	d.Brands = RemoveDuplicateStrings(d.Brands)
}

func (d *Dealer) Patch(patch *DealerPatch) {
	if patch.Name != nil {
		d.Name = *patch.Name
	}

	if patch.City != nil {
		d.City = *patch.City
	}

	if patch.PhoneNumber != nil {
		d.PhoneNumber = *patch.PhoneNumber
	}

	if patch.Address != nil {
		d.Address = *patch.Address
	}

	if patch.Province != nil {
		d.Province = *patch.Province
	}

	if patch.Country != nil {
		d.Country = *patch.Country
	}

	if patch.PostalCode != nil {
		d.PostalCode = *patch.PostalCode
	}

	if patch.Brands != nil {
		d.Brands = *patch.Brands
	}
}

// ToJson convert a Dealer to a json string
func (d *Dealer) ToJson() string {
	b, _ := json.Marshal(d)
	return string(b)
}

func (d *DealerPatch) ToJson() string {
	b, _ := json.Marshal(d)
	return string(b)
}

// DealerFromJson will decode the input and return a Dealer
func DealerFromJson(data io.Reader) *Dealer {
	var dealer *Dealer
	json.NewDecoder(data).Decode(&dealer)
	return dealer
}

func DealerPatchFromJson(data io.Reader) *DealerPatch {
	var dealer *DealerPatch
	json.NewDecoder(data).Decode(&dealer)
	return dealer
}

func InvalidDealerError(fieldName string, dealerId string) *AppError {
	id := fmt.Sprintf("model.dealer.is_valid.%s.app_error", fieldName)
	details := ""
	if dealerId != "" {
		details = "dealer_id=" + dealerId
	}
	return NewAppError("Dealer.IsValid", id, nil, details, http.StatusBadRequest)
}

func IsValidDealerName(s string) bool {
	if len(s) < DEALER_NAME_MIN_LENGTH || len(s) > DEALER_NAME_MAX_LENGTH {
		return false
	}

	return true
}

type DealerSlice []*Dealer

func (d DealerSlice) FilterByName(names []string) DealerSlice {
	var matches []*Dealer
	for _, dealer := range d {
		for _, name := range names {
			if name == dealer.Name {
				matches = append(matches, dealer)
			}
		}
	}
	return DealerSlice(matches)
}
