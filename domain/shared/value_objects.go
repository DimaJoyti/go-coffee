package shared

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

// Email represents an email address value object
type Email struct {
	value string
}

// NewEmail creates a new Email value object
func NewEmail(value string) (Email, error) {
	if value == "" {
		return Email{}, errors.New("email cannot be empty")
	}
	
	// Simple email validation regex
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(value) {
		return Email{}, errors.New("invalid email format")
	}
	
	return Email{value: strings.ToLower(value)}, nil
}

// Value returns the email value
func (e Email) Value() string {
	return e.value
}

// String implements the Stringer interface
func (e Email) String() string {
	return e.value
}

// Equals checks if two emails are equal
func (e Email) Equals(other Email) bool {
	return e.value == other.value
}

// Domain returns the domain part of the email
func (e Email) Domain() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}

// LocalPart returns the local part of the email
func (e Email) LocalPart() string {
	parts := strings.Split(e.value, "@")
	if len(parts) == 2 {
		return parts[0]
	}
	return ""
}

// PhoneNumber represents a phone number value object
type PhoneNumber struct {
	value string
}

// NewPhoneNumber creates a new PhoneNumber value object
func NewPhoneNumber(value string) (PhoneNumber, error) {
	if value == "" {
		return PhoneNumber{}, errors.New("phone number cannot be empty")
	}
	
	// Remove all non-digit characters except +
	cleaned := regexp.MustCompile(`[^\d+]`).ReplaceAllString(value, "")
	
	// Basic phone number validation
	if len(cleaned) < 10 || len(cleaned) > 15 {
		return PhoneNumber{}, errors.New("phone number must be between 10 and 15 digits")
	}
	
	return PhoneNumber{value: cleaned}, nil
}

// Value returns the phone number value
func (p PhoneNumber) Value() string {
	return p.value
}

// String implements the Stringer interface
func (p PhoneNumber) String() string {
	return p.value
}

// Equals checks if two phone numbers are equal
func (p PhoneNumber) Equals(other PhoneNumber) bool {
	return p.value == other.value
}

// Money represents a monetary value with currency
type Money struct {
	amount   int64  // Amount in smallest currency unit (e.g., cents)
	currency string // Currency code (e.g., USD, EUR)
}

// NewMoney creates a new Money value object
func NewMoney(amount int64, currency string) (Money, error) {
	if currency == "" {
		return Money{}, errors.New("currency cannot be empty")
	}
	
	// Validate currency code (should be 3 characters)
	if len(currency) != 3 {
		return Money{}, errors.New("currency code must be 3 characters")
	}
	
	return Money{
		amount:   amount,
		currency: strings.ToUpper(currency),
	}, nil
}

// NewMoneyFromFloat creates Money from a float value
func NewMoneyFromFloat(amount float64, currency string) (Money, error) {
	// Convert to smallest unit (multiply by 100 for most currencies)
	amountInSmallestUnit := int64(amount * 100)
	return NewMoney(amountInSmallestUnit, currency)
}

// Amount returns the amount in smallest currency unit
func (m Money) Amount() int64 {
	return m.amount
}

// Currency returns the currency code
func (m Money) Currency() string {
	return m.currency
}

// ToFloat returns the amount as a float
func (m Money) ToFloat() float64 {
	return float64(m.amount) / 100.0
}

// String implements the Stringer interface
func (m Money) String() string {
	return fmt.Sprintf("%.2f %s", m.ToFloat(), m.currency)
}

// Equals checks if two Money values are equal
func (m Money) Equals(other Money) bool {
	return m.amount == other.amount && m.currency == other.currency
}

// Add adds two Money values (must have same currency)
func (m Money) Add(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot add money with different currencies")
	}
	return Money{amount: m.amount + other.amount, currency: m.currency}, nil
}

// Subtract subtracts two Money values (must have same currency)
func (m Money) Subtract(other Money) (Money, error) {
	if m.currency != other.currency {
		return Money{}, errors.New("cannot subtract money with different currencies")
	}
	return Money{amount: m.amount - other.amount, currency: m.currency}, nil
}

// Multiply multiplies Money by a factor
func (m Money) Multiply(factor float64) Money {
	return Money{amount: int64(float64(m.amount) * factor), currency: m.currency}
}

// IsPositive checks if the money amount is positive
func (m Money) IsPositive() bool {
	return m.amount > 0
}

// IsNegative checks if the money amount is negative
func (m Money) IsNegative() bool {
	return m.amount < 0
}

// IsZero checks if the money amount is zero
func (m Money) IsZero() bool {
	return m.amount == 0
}

// Address represents a physical address
type Address struct {
	street     string
	city       string
	state      string
	postalCode string
	country    string
}

// NewAddress creates a new Address value object
func NewAddress(street, city, state, postalCode, country string) (Address, error) {
	if street == "" {
		return Address{}, errors.New("street cannot be empty")
	}
	if city == "" {
		return Address{}, errors.New("city cannot be empty")
	}
	if country == "" {
		return Address{}, errors.New("country cannot be empty")
	}
	
	return Address{
		street:     strings.TrimSpace(street),
		city:       strings.TrimSpace(city),
		state:      strings.TrimSpace(state),
		postalCode: strings.TrimSpace(postalCode),
		country:    strings.TrimSpace(country),
	}, nil
}

// Street returns the street address
func (a Address) Street() string {
	return a.street
}

// City returns the city
func (a Address) City() string {
	return a.city
}

// State returns the state
func (a Address) State() string {
	return a.state
}

// PostalCode returns the postal code
func (a Address) PostalCode() string {
	return a.postalCode
}

// Country returns the country
func (a Address) Country() string {
	return a.country
}

// String implements the Stringer interface
func (a Address) String() string {
	parts := []string{a.street, a.city}
	if a.state != "" {
		parts = append(parts, a.state)
	}
	if a.postalCode != "" {
		parts = append(parts, a.postalCode)
	}
	parts = append(parts, a.country)
	return strings.Join(parts, ", ")
}

// Equals checks if two addresses are equal
func (a Address) Equals(other Address) bool {
	return a.street == other.street &&
		a.city == other.city &&
		a.state == other.state &&
		a.postalCode == other.postalCode &&
		a.country == other.country
}

// DateRange represents a date range value object
type DateRange struct {
	startDate time.Time
	endDate   time.Time
}

// NewDateRange creates a new DateRange value object
func NewDateRange(startDate, endDate time.Time) (DateRange, error) {
	if startDate.After(endDate) {
		return DateRange{}, errors.New("start date cannot be after end date")
	}
	
	return DateRange{
		startDate: startDate,
		endDate:   endDate,
	}, nil
}

// StartDate returns the start date
func (dr DateRange) StartDate() time.Time {
	return dr.startDate
}

// EndDate returns the end date
func (dr DateRange) EndDate() time.Time {
	return dr.endDate
}

// Duration returns the duration of the date range
func (dr DateRange) Duration() time.Duration {
	return dr.endDate.Sub(dr.startDate)
}

// Contains checks if a date is within the range
func (dr DateRange) Contains(date time.Time) bool {
	return (date.Equal(dr.startDate) || date.After(dr.startDate)) &&
		(date.Equal(dr.endDate) || date.Before(dr.endDate))
}

// Overlaps checks if two date ranges overlap
func (dr DateRange) Overlaps(other DateRange) bool {
	return dr.startDate.Before(other.endDate) && dr.endDate.After(other.startDate)
}

// String implements the Stringer interface
func (dr DateRange) String() string {
	return fmt.Sprintf("%s to %s", 
		dr.startDate.Format("2006-01-02"), 
		dr.endDate.Format("2006-01-02"))
}

// Equals checks if two date ranges are equal
func (dr DateRange) Equals(other DateRange) bool {
	return dr.startDate.Equal(other.startDate) && dr.endDate.Equal(other.endDate)
}

// Percentage represents a percentage value
type Percentage struct {
	value float64
}

// NewPercentage creates a new Percentage value object
func NewPercentage(value float64) (Percentage, error) {
	if value < 0 || value > 100 {
		return Percentage{}, errors.New("percentage must be between 0 and 100")
	}
	
	return Percentage{value: value}, nil
}

// Value returns the percentage value
func (p Percentage) Value() float64 {
	return p.value
}

// AsDecimal returns the percentage as a decimal (0.0 to 1.0)
func (p Percentage) AsDecimal() float64 {
	return p.value / 100.0
}

// String implements the Stringer interface
func (p Percentage) String() string {
	return fmt.Sprintf("%.2f%%", p.value)
}

// Equals checks if two percentages are equal
func (p Percentage) Equals(other Percentage) bool {
	return p.value == other.value
}

// Add adds two percentages
func (p Percentage) Add(other Percentage) (Percentage, error) {
	result := p.value + other.value
	if result > 100 {
		return Percentage{}, errors.New("result exceeds 100%")
	}
	return Percentage{value: result}, nil
}

// Subtract subtracts two percentages
func (p Percentage) Subtract(other Percentage) (Percentage, error) {
	result := p.value - other.value
	if result < 0 {
		return Percentage{}, errors.New("result is less than 0%")
	}
	return Percentage{value: result}, nil
}

// Rating represents a rating value (e.g., 1-5 stars)
type Rating struct {
	value float64
	scale int
}

// NewRating creates a new Rating value object
func NewRating(value float64, scale int) (Rating, error) {
	if scale <= 0 {
		return Rating{}, errors.New("scale must be positive")
	}
	if value < 0 || value > float64(scale) {
		return Rating{}, fmt.Errorf("rating must be between 0 and %d", scale)
	}
	
	return Rating{value: value, scale: scale}, nil
}

// Value returns the rating value
func (r Rating) Value() float64 {
	return r.value
}

// Scale returns the rating scale
func (r Rating) Scale() int {
	return r.scale
}

// AsPercentage returns the rating as a percentage
func (r Rating) AsPercentage() Percentage {
	percentage := (r.value / float64(r.scale)) * 100
	p, _ := NewPercentage(percentage) // Safe because we control the input
	return p
}

// String implements the Stringer interface
func (r Rating) String() string {
	return fmt.Sprintf("%.1f/%d", r.value, r.scale)
}

// Equals checks if two ratings are equal
func (r Rating) Equals(other Rating) bool {
	return r.value == other.value && r.scale == other.scale
}

// IsMaxRating checks if this is the maximum rating
func (r Rating) IsMaxRating() bool {
	return r.value == float64(r.scale)
}

// IsMinRating checks if this is the minimum rating
func (r Rating) IsMinRating() bool {
	return r.value == 0
}
