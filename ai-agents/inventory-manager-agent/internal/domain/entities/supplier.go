package entities

import (
	"time"

	"github.com/google/uuid"
)

// Supplier represents a supplier entity with comprehensive information
type Supplier struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	Code              string                 `json:"code" redis:"code"`
	Name              string                 `json:"name" redis:"name"`
	LegalName         string                 `json:"legal_name" redis:"legal_name"`
	Type              SupplierType           `json:"type" redis:"type"`
	Category          SupplierCategory       `json:"category" redis:"category"`
	Status            SupplierStatus         `json:"status" redis:"status"`
	Rating            float64                `json:"rating" redis:"rating"`
	ContactInfo       *ContactInformation    `json:"contact_info,omitempty"`
	Address           *Address               `json:"address,omitempty"`
	BillingAddress    *Address               `json:"billing_address,omitempty"`
	ShippingAddresses []*Address             `json:"shipping_addresses,omitempty"`
	PaymentTerms      *PaymentTerms          `json:"payment_terms,omitempty"`
	DeliveryTerms     *DeliveryTerms         `json:"delivery_terms,omitempty"`
	Certifications    []*Certification       `json:"certifications,omitempty"`
	Performance       *SupplierPerformance   `json:"performance,omitempty"`
	Contracts         []*SupplierContract    `json:"contracts,omitempty"`
	Products          []*SupplierProduct     `json:"products,omitempty"`
	PriceList         []*PriceListItem       `json:"price_list,omitempty"`
	TaxInfo           *TaxInformation        `json:"tax_info,omitempty"`
	BankInfo          *BankInformation       `json:"bank_info,omitempty"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	Tags              []string               `json:"tags" redis:"tags"`
	Notes             string                 `json:"notes" redis:"notes"`
	IsActive          bool                   `json:"is_active" redis:"is_active"`
	IsPreferred       bool                   `json:"is_preferred" redis:"is_preferred"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy         string                 `json:"created_by" redis:"created_by"`
	UpdatedBy         string                 `json:"updated_by" redis:"updated_by"`
	Version           int64                  `json:"version" redis:"version"`
}

// SupplierType defines the type of supplier
type SupplierType string

const (
	SupplierTypeManufacturer SupplierType = "manufacturer"
	SupplierTypeDistributor  SupplierType = "distributor"
	SupplierTypeWholesaler   SupplierType = "wholesaler"
	SupplierTypeRetailer     SupplierType = "retailer"
	SupplierTypeService      SupplierType = "service"
	SupplierTypeFreelancer   SupplierType = "freelancer"
)

// SupplierCategory defines the category of supplier
type SupplierCategory string

const (
	SupplierCategoryFood         SupplierCategory = "food"
	SupplierCategoryBeverage     SupplierCategory = "beverage"
	SupplierCategoryPackaging    SupplierCategory = "packaging"
	SupplierCategoryEquipment    SupplierCategory = "equipment"
	SupplierCategoryMaintenance  SupplierCategory = "maintenance"
	SupplierCategoryCleaning     SupplierCategory = "cleaning"
	SupplierCategoryOfficeSupply SupplierCategory = "office_supply"
	SupplierCategoryMarketing    SupplierCategory = "marketing"
	SupplierCategoryLogistics    SupplierCategory = "logistics"
	SupplierCategoryUtilities    SupplierCategory = "utilities"
)

// SupplierStatus defines the status of a supplier
type SupplierStatus string

const (
	SupplierStatusActive     SupplierStatus = "active"
	SupplierStatusInactive   SupplierStatus = "inactive"
	SupplierStatusSuspended  SupplierStatus = "suspended"
	SupplierStatusBlacklisted SupplierStatus = "blacklisted"
	SupplierStatusPending    SupplierStatus = "pending"
	SupplierStatusApproved   SupplierStatus = "approved"
	SupplierStatusRejected   SupplierStatus = "rejected"
)

// ContactInformation contains contact details for a supplier
type ContactInformation struct {
	PrimaryContact   *Contact   `json:"primary_contact,omitempty"`
	SecondaryContact *Contact   `json:"secondary_contact,omitempty"`
	SalesContact     *Contact   `json:"sales_contact,omitempty"`
	SupportContact   *Contact   `json:"support_contact,omitempty"`
	BillingContact   *Contact   `json:"billing_contact,omitempty"`
	Contacts         []*Contact `json:"contacts,omitempty"`
	Website          string     `json:"website" redis:"website"`
	SocialMedia      map[string]string `json:"social_media" redis:"social_media"`
}

// Contact represents a contact person
type Contact struct {
	ID          uuid.UUID `json:"id" redis:"id"`
	Name        string    `json:"name" redis:"name"`
	Title       string    `json:"title" redis:"title"`
	Department  string    `json:"department" redis:"department"`
	Email       string    `json:"email" redis:"email"`
	Phone       string    `json:"phone" redis:"phone"`
	Mobile      string    `json:"mobile" redis:"mobile"`
	Fax         string    `json:"fax" redis:"fax"`
	Language    string    `json:"language" redis:"language"`
	TimeZone    string    `json:"time_zone" redis:"time_zone"`
	IsActive    bool      `json:"is_active" redis:"is_active"`
	IsPrimary   bool      `json:"is_primary" redis:"is_primary"`
	CreatedAt   time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" redis:"updated_at"`
}

// Address represents a physical address
type Address struct {
	ID           uuid.UUID `json:"id" redis:"id"`
	Type         string    `json:"type" redis:"type"` // billing, shipping, main
	Name         string    `json:"name" redis:"name"`
	Street1      string    `json:"street1" redis:"street1"`
	Street2      string    `json:"street2" redis:"street2"`
	City         string    `json:"city" redis:"city"`
	State        string    `json:"state" redis:"state"`
	PostalCode   string    `json:"postal_code" redis:"postal_code"`
	Country      string    `json:"country" redis:"country"`
	CountryCode  string    `json:"country_code" redis:"country_code"`
	Latitude     float64   `json:"latitude" redis:"latitude"`
	Longitude    float64   `json:"longitude" redis:"longitude"`
	Instructions string    `json:"instructions" redis:"instructions"`
	IsDefault    bool      `json:"is_default" redis:"is_default"`
	IsActive     bool      `json:"is_active" redis:"is_active"`
	CreatedAt    time.Time `json:"created_at" redis:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" redis:"updated_at"`
}

// PaymentTerms defines payment terms for a supplier
type PaymentTerms struct {
	TermsCode        string        `json:"terms_code" redis:"terms_code"`
	Description      string        `json:"description" redis:"description"`
	PaymentDays      int           `json:"payment_days" redis:"payment_days"`
	DiscountDays     int           `json:"discount_days" redis:"discount_days"`
	DiscountPercent  float64       `json:"discount_percent" redis:"discount_percent"`
	PaymentMethods   []string      `json:"payment_methods" redis:"payment_methods"`
	Currency         string        `json:"currency" redis:"currency"`
	CreditLimit      Money         `json:"credit_limit" redis:"credit_limit"`
	MinOrderAmount   Money         `json:"min_order_amount" redis:"min_order_amount"`
	MaxOrderAmount   Money         `json:"max_order_amount" redis:"max_order_amount"`
	LateFeePercent   float64       `json:"late_fee_percent" redis:"late_fee_percent"`
	IsActive         bool          `json:"is_active" redis:"is_active"`
	EffectiveDate    time.Time     `json:"effective_date" redis:"effective_date"`
	ExpirationDate   *time.Time    `json:"expiration_date,omitempty" redis:"expiration_date"`
}

// DeliveryTerms defines delivery terms for a supplier
type DeliveryTerms struct {
	TermsCode         string     `json:"terms_code" redis:"terms_code"`
	Description       string     `json:"description" redis:"description"`
	LeadTimeDays      int        `json:"lead_time_days" redis:"lead_time_days"`
	MinLeadTimeDays   int        `json:"min_lead_time_days" redis:"min_lead_time_days"`
	MaxLeadTimeDays   int        `json:"max_lead_time_days" redis:"max_lead_time_days"`
	DeliveryMethods   []string   `json:"delivery_methods" redis:"delivery_methods"`
	ShippingCost      Money      `json:"shipping_cost" redis:"shipping_cost"`
	FreeShippingMin   Money      `json:"free_shipping_min" redis:"free_shipping_min"`
	DeliveryDays      []string   `json:"delivery_days" redis:"delivery_days"`
	DeliveryTimeStart string     `json:"delivery_time_start" redis:"delivery_time_start"`
	DeliveryTimeEnd   string     `json:"delivery_time_end" redis:"delivery_time_end"`
	SpecialInstructions string   `json:"special_instructions" redis:"special_instructions"`
	IsActive          bool       `json:"is_active" redis:"is_active"`
	EffectiveDate     time.Time  `json:"effective_date" redis:"effective_date"`
	ExpirationDate    *time.Time `json:"expiration_date,omitempty" redis:"expiration_date"`
}

// Certification represents a supplier certification
type Certification struct {
	ID             uuid.UUID  `json:"id" redis:"id"`
	Name           string     `json:"name" redis:"name"`
	Type           string     `json:"type" redis:"type"`
	IssuingBody    string     `json:"issuing_body" redis:"issuing_body"`
	CertificateNumber string  `json:"certificate_number" redis:"certificate_number"`
	IssueDate      time.Time  `json:"issue_date" redis:"issue_date"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty" redis:"expiration_date"`
	Status         string     `json:"status" redis:"status"`
	DocumentURL    string     `json:"document_url" redis:"document_url"`
	VerifiedBy     string     `json:"verified_by" redis:"verified_by"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty" redis:"verified_at"`
	Notes          string     `json:"notes" redis:"notes"`
	IsActive       bool       `json:"is_active" redis:"is_active"`
	CreatedAt      time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" redis:"updated_at"`
}

// SupplierPerformance tracks supplier performance metrics
type SupplierPerformance struct {
	OverallRating        float64                `json:"overall_rating" redis:"overall_rating"`
	QualityRating        float64                `json:"quality_rating" redis:"quality_rating"`
	DeliveryRating       float64                `json:"delivery_rating" redis:"delivery_rating"`
	ServiceRating        float64                `json:"service_rating" redis:"service_rating"`
	PriceRating          float64                `json:"price_rating" redis:"price_rating"`
	OnTimeDeliveryRate   float64                `json:"on_time_delivery_rate" redis:"on_time_delivery_rate"`
	QualityRejectRate    float64                `json:"quality_reject_rate" redis:"quality_reject_rate"`
	OrderFulfillmentRate float64                `json:"order_fulfillment_rate" redis:"order_fulfillment_rate"`
	ResponseTime         time.Duration          `json:"response_time" redis:"response_time"`
	TotalOrders          int                    `json:"total_orders" redis:"total_orders"`
	TotalValue           Money                  `json:"total_value" redis:"total_value"`
	AverageOrderValue    Money                  `json:"average_order_value" redis:"average_order_value"`
	LastOrderDate        *time.Time             `json:"last_order_date,omitempty" redis:"last_order_date"`
	Metrics              map[string]interface{} `json:"metrics" redis:"metrics"`
	LastUpdated          time.Time              `json:"last_updated" redis:"last_updated"`
}

// SupplierContract represents a contract with a supplier
type SupplierContract struct {
	ID             uuid.UUID              `json:"id" redis:"id"`
	ContractNumber string                 `json:"contract_number" redis:"contract_number"`
	Type           string                 `json:"type" redis:"type"`
	Status         string                 `json:"status" redis:"status"`
	StartDate      time.Time              `json:"start_date" redis:"start_date"`
	EndDate        *time.Time             `json:"end_date,omitempty" redis:"end_date"`
	Value          Money                  `json:"value" redis:"value"`
	Currency       string                 `json:"currency" redis:"currency"`
	PaymentTerms   *PaymentTerms          `json:"payment_terms,omitempty"`
	DeliveryTerms  *DeliveryTerms         `json:"delivery_terms,omitempty"`
	Terms          string                 `json:"terms" redis:"terms"`
	DocumentURL    string                 `json:"document_url" redis:"document_url"`
	Attributes     map[string]interface{} `json:"attributes" redis:"attributes"`
	IsActive       bool                   `json:"is_active" redis:"is_active"`
	CreatedAt      time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at" redis:"updated_at"`
	CreatedBy      string                 `json:"created_by" redis:"created_by"`
	UpdatedBy      string                 `json:"updated_by" redis:"updated_by"`
}

// SupplierProduct represents a product offered by a supplier
type SupplierProduct struct {
	ID                uuid.UUID              `json:"id" redis:"id"`
	SupplierSKU       string                 `json:"supplier_sku" redis:"supplier_sku"`
	InternalSKU       string                 `json:"internal_sku" redis:"internal_sku"`
	Name              string                 `json:"name" redis:"name"`
	Description       string                 `json:"description" redis:"description"`
	Category          string                 `json:"category" redis:"category"`
	Unit              MeasurementUnit        `json:"unit" redis:"unit"`
	PackSize          float64                `json:"pack_size" redis:"pack_size"`
	MinOrderQuantity  float64                `json:"min_order_quantity" redis:"min_order_quantity"`
	MaxOrderQuantity  float64                `json:"max_order_quantity" redis:"max_order_quantity"`
	LeadTimeDays      int                    `json:"lead_time_days" redis:"lead_time_days"`
	ShelfLifeDays     int                    `json:"shelf_life_days" redis:"shelf_life_days"`
	StorageConditions *StorageRequirements   `json:"storage_conditions,omitempty"`
	Attributes        map[string]interface{} `json:"attributes" redis:"attributes"`
	IsActive          bool                   `json:"is_active" redis:"is_active"`
	IsPreferred       bool                   `json:"is_preferred" redis:"is_preferred"`
	CreatedAt         time.Time              `json:"created_at" redis:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" redis:"updated_at"`
}

// PriceListItem represents a price list item from a supplier
type PriceListItem struct {
	ID               uuid.UUID  `json:"id" redis:"id"`
	SupplierProductID uuid.UUID `json:"supplier_product_id" redis:"supplier_product_id"`
	SupplierSKU      string     `json:"supplier_sku" redis:"supplier_sku"`
	UnitPrice        Money      `json:"unit_price" redis:"unit_price"`
	MinQuantity      float64    `json:"min_quantity" redis:"min_quantity"`
	MaxQuantity      float64    `json:"max_quantity" redis:"max_quantity"`
	DiscountPercent  float64    `json:"discount_percent" redis:"discount_percent"`
	EffectiveDate    time.Time  `json:"effective_date" redis:"effective_date"`
	ExpirationDate   *time.Time `json:"expiration_date,omitempty" redis:"expiration_date"`
	IsActive         bool       `json:"is_active" redis:"is_active"`
	CreatedAt        time.Time  `json:"created_at" redis:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" redis:"updated_at"`
}

// TaxInformation contains tax-related information for a supplier
type TaxInformation struct {
	TaxID           string `json:"tax_id" redis:"tax_id"`
	VATNumber       string `json:"vat_number" redis:"vat_number"`
	TaxExempt       bool   `json:"tax_exempt" redis:"tax_exempt"`
	TaxExemptNumber string `json:"tax_exempt_number" redis:"tax_exempt_number"`
	TaxRate         float64 `json:"tax_rate" redis:"tax_rate"`
	TaxCategory     string `json:"tax_category" redis:"tax_category"`
	TaxCountry      string `json:"tax_country" redis:"tax_country"`
	TaxState        string `json:"tax_state" redis:"tax_state"`
}

// BankInformation contains banking information for a supplier
type BankInformation struct {
	BankName        string `json:"bank_name" redis:"bank_name"`
	AccountName     string `json:"account_name" redis:"account_name"`
	AccountNumber   string `json:"account_number" redis:"account_number"`
	RoutingNumber   string `json:"routing_number" redis:"routing_number"`
	IBAN            string `json:"iban" redis:"iban"`
	SWIFT           string `json:"swift" redis:"swift"`
	BankAddress     *Address `json:"bank_address,omitempty"`
	AccountType     string `json:"account_type" redis:"account_type"`
	Currency        string `json:"currency" redis:"currency"`
	IsActive        bool   `json:"is_active" redis:"is_active"`
}

// NewSupplier creates a new supplier with default values
func NewSupplier(name, code string, supplierType SupplierType, category SupplierCategory) *Supplier {
	now := time.Now()
	return &Supplier{
		ID:          uuid.New(),
		Code:        code,
		Name:        name,
		Type:        supplierType,
		Category:    category,
		Status:      SupplierStatusPending,
		Rating:      0.0,
		Attributes:  make(map[string]interface{}),
		Tags:        []string{},
		IsActive:    true,
		IsPreferred: false,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
}

// UpdateRating updates the supplier's overall rating
func (s *Supplier) UpdateRating(newRating float64) {
	if newRating < 0 {
		newRating = 0
	}
	if newRating > 5 {
		newRating = 5
	}
	
	s.Rating = newRating
	s.UpdatedAt = time.Now()
	s.Version++
}

// Activate activates the supplier
func (s *Supplier) Activate() {
	s.IsActive = true
	s.Status = SupplierStatusActive
	s.UpdatedAt = time.Now()
	s.Version++
}

// Deactivate deactivates the supplier
func (s *Supplier) Deactivate() {
	s.IsActive = false
	s.Status = SupplierStatusInactive
	s.UpdatedAt = time.Now()
	s.Version++
}

// Suspend suspends the supplier
func (s *Supplier) Suspend() {
	s.Status = SupplierStatusSuspended
	s.UpdatedAt = time.Now()
	s.Version++
}

// Blacklist blacklists the supplier
func (s *Supplier) Blacklist() {
	s.IsActive = false
	s.Status = SupplierStatusBlacklisted
	s.UpdatedAt = time.Now()
	s.Version++
}

// AddProduct adds a product to the supplier's catalog
func (s *Supplier) AddProduct(product *SupplierProduct) {
	s.Products = append(s.Products, product)
	s.UpdatedAt = time.Now()
	s.Version++
}

// GetActiveProducts returns all active products from the supplier
func (s *Supplier) GetActiveProducts() []*SupplierProduct {
	var activeProducts []*SupplierProduct
	for _, product := range s.Products {
		if product.IsActive {
			activeProducts = append(activeProducts, product)
		}
	}
	return activeProducts
}

// GetProductBySKU returns a product by supplier SKU
func (s *Supplier) GetProductBySKU(sku string) *SupplierProduct {
	for _, product := range s.Products {
		if product.SupplierSKU == sku && product.IsActive {
			return product
		}
	}
	return nil
}

// HasValidContract checks if the supplier has a valid active contract
func (s *Supplier) HasValidContract() bool {
	now := time.Now()
	for _, contract := range s.Contracts {
		if contract.IsActive && 
		   contract.StartDate.Before(now) && 
		   (contract.EndDate == nil || contract.EndDate.After(now)) {
			return true
		}
	}
	return false
}

// GetCurrentPrice returns the current price for a product
func (s *Supplier) GetCurrentPrice(supplierSKU string, quantity float64) *Money {
	now := time.Now()
	var bestPrice *Money
	
	for _, priceItem := range s.PriceList {
		if priceItem.SupplierSKU == supplierSKU && 
		   priceItem.IsActive &&
		   priceItem.EffectiveDate.Before(now) &&
		   (priceItem.ExpirationDate == nil || priceItem.ExpirationDate.After(now)) &&
		   quantity >= priceItem.MinQuantity &&
		   (priceItem.MaxQuantity == 0 || quantity <= priceItem.MaxQuantity) {
			
			price := priceItem.UnitPrice
			if priceItem.DiscountPercent > 0 {
				discountAmount := price.Amount * (priceItem.DiscountPercent / 100)
				price.Amount -= discountAmount
			}
			
			if bestPrice == nil || price.Amount < bestPrice.Amount {
				bestPrice = &price
			}
		}
	}
	
	return bestPrice
}

// IsReliable checks if the supplier is considered reliable based on performance
func (s *Supplier) IsReliable() bool {
	if s.Performance == nil {
		return false
	}
	
	return s.Performance.OverallRating >= 4.0 &&
		   s.Performance.OnTimeDeliveryRate >= 0.9 &&
		   s.Performance.QualityRejectRate <= 0.05
}

// CanFulfillOrder checks if the supplier can fulfill an order
func (s *Supplier) CanFulfillOrder(items map[string]float64) bool {
	if !s.IsActive || s.Status != SupplierStatusActive {
		return false
	}
	
	for sku, quantity := range items {
		product := s.GetProductBySKU(sku)
		if product == nil || !product.IsActive {
			return false
		}
		
		if quantity < product.MinOrderQuantity {
			return false
		}
		
		if product.MaxOrderQuantity > 0 && quantity > product.MaxOrderQuantity {
			return false
		}
	}
	
	return true
}
