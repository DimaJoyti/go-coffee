package shared

import (
	"time"
)

// AggregateID represents a unique identifier for aggregates
type AggregateID struct {
	value string
}

// NewAggregateID creates a new AggregateID
func NewAggregateID(value string) AggregateID {
	return AggregateID{value: value}
}

// Value returns the string value of AggregateID
func (id AggregateID) Value() string {
	return id.value
}

// String implements the Stringer interface
func (id AggregateID) String() string {
	return id.value
}

// Equals checks if two AggregateIDs are equal
func (id AggregateID) Equals(other AggregateID) bool {
	return id.value == other.value
}

// IsEmpty checks if AggregateID is empty
func (id AggregateID) IsEmpty() bool {
	return id.value == ""
}

// BaseAggregate provides common functionality for all aggregates
type BaseAggregate struct {
	id             AggregateID
	tenantID       TenantID
	version        int
	createdAt      time.Time
	updatedAt      time.Time
	domainEvents   []DomainEvent
}

// NewBaseAggregate creates a new base aggregate
func NewBaseAggregate(id AggregateID, tenantID TenantID) *BaseAggregate {
	now := time.Now()
	return &BaseAggregate{
		id:           id,
		tenantID:     tenantID,
		version:      1,
		createdAt:    now,
		updatedAt:    now,
		domainEvents: make([]DomainEvent, 0),
	}
}

// ID returns the aggregate ID
func (ba *BaseAggregate) ID() AggregateID {
	return ba.id
}

// GetTenantID returns the tenant ID (implements TenantAware)
func (ba *BaseAggregate) GetTenantID() TenantID {
	return ba.tenantID
}

// SetTenantID sets the tenant ID (implements TenantAware)
func (ba *BaseAggregate) SetTenantID(tenantID TenantID) {
	ba.tenantID = tenantID
	ba.markUpdated()
}

// Version returns the aggregate version
func (ba *BaseAggregate) Version() int {
	return ba.version
}

// CreatedAt returns the creation timestamp
func (ba *BaseAggregate) CreatedAt() time.Time {
	return ba.createdAt
}

// UpdatedAt returns the last update timestamp
func (ba *BaseAggregate) UpdatedAt() time.Time {
	return ba.updatedAt
}

// IncrementVersion increments the aggregate version
func (ba *BaseAggregate) IncrementVersion() {
	ba.version++
	ba.markUpdated()
}

// markUpdated updates the updatedAt timestamp
func (ba *BaseAggregate) markUpdated() {
	ba.updatedAt = time.Now()
}

// AddDomainEvent adds a domain event to the aggregate
func (ba *BaseAggregate) AddDomainEvent(event DomainEvent) {
	ba.domainEvents = append(ba.domainEvents, event)
}

// GetDomainEvents returns all domain events
func (ba *BaseAggregate) GetDomainEvents() []DomainEvent {
	return ba.domainEvents
}

// ClearDomainEvents clears all domain events
func (ba *BaseAggregate) ClearDomainEvents() {
	ba.domainEvents = make([]DomainEvent, 0)
}

// HasDomainEvents checks if there are any domain events
func (ba *BaseAggregate) HasDomainEvents() bool {
	return len(ba.domainEvents) > 0
}

// Entity represents a domain entity with identity
type Entity struct {
	id        AggregateID
	tenantID  TenantID
	createdAt time.Time
	updatedAt time.Time
}

// NewEntity creates a new entity
func NewEntity(id AggregateID, tenantID TenantID) *Entity {
	now := time.Now()
	return &Entity{
		id:        id,
		tenantID:  tenantID,
		createdAt: now,
		updatedAt: now,
	}
}

// ID returns the entity ID
func (e *Entity) ID() AggregateID {
	return e.id
}

// GetTenantID returns the tenant ID (implements TenantAware)
func (e *Entity) GetTenantID() TenantID {
	return e.tenantID
}

// SetTenantID sets the tenant ID (implements TenantAware)
func (e *Entity) SetTenantID(tenantID TenantID) {
	e.tenantID = tenantID
	e.markUpdated()
}

// CreatedAt returns the creation timestamp
func (e *Entity) CreatedAt() time.Time {
	return e.createdAt
}

// UpdatedAt returns the last update timestamp
func (e *Entity) UpdatedAt() time.Time {
	return e.updatedAt
}

// markUpdated updates the updatedAt timestamp
func (e *Entity) markUpdated() {
	e.updatedAt = time.Now()
}

// Equals checks if two entities are equal based on ID and tenant
func (e *Entity) Equals(other *Entity) bool {
	if other == nil {
		return false
	}
	return e.id.Equals(other.id) && e.tenantID.Equals(other.tenantID)
}

// Repository interface for aggregate repositories
type Repository interface {
	// Save saves an aggregate
	Save(aggregate interface{}) error
	
	// FindByID finds an aggregate by ID within tenant context
	FindByID(id AggregateID, tenantID TenantID) (interface{}, error)
	
	// Delete removes an aggregate
	Delete(id AggregateID, tenantID TenantID) error
	
	// ExistsByID checks if an aggregate exists
	ExistsByID(id AggregateID, tenantID TenantID) (bool, error)
}

// TenantAwareRepository extends Repository with tenant-specific operations
type TenantAwareRepository interface {
	Repository
	
	// FindAllByTenant finds all aggregates for a tenant
	FindAllByTenant(tenantID TenantID) ([]interface{}, error)
	
	// CountByTenant counts aggregates for a tenant
	CountByTenant(tenantID TenantID) (int64, error)
	
	// DeleteAllByTenant deletes all aggregates for a tenant
	DeleteAllByTenant(tenantID TenantID) error
}

// Specification interface for domain specifications
type Specification interface {
	// IsSatisfiedBy checks if the specification is satisfied
	IsSatisfiedBy(candidate interface{}) bool
	
	// And combines this specification with another using AND logic
	And(other Specification) Specification
	
	// Or combines this specification with another using OR logic
	Or(other Specification) Specification
	
	// Not negates this specification
	Not() Specification
}

// BaseSpecification provides common functionality for specifications
type BaseSpecification struct{}

// IsSatisfiedBy is a default implementation that always returns true
// Concrete specifications should override this method
func (bs *BaseSpecification) IsSatisfiedBy(candidate interface{}) bool {
	return true
}

// And combines specifications with AND logic
func (bs *BaseSpecification) And(other Specification) Specification {
	return &AndSpecification{
		left:  bs,
		right: other,
	}
}

// Or combines specifications with OR logic
func (bs *BaseSpecification) Or(other Specification) Specification {
	return &OrSpecification{
		left:  bs,
		right: other,
	}
}

// Not negates the specification
func (bs *BaseSpecification) Not() Specification {
	return &NotSpecification{
		spec: bs,
	}
}

// AndSpecification combines two specifications with AND logic
type AndSpecification struct {
	left  Specification
	right Specification
}

// IsSatisfiedBy checks if both specifications are satisfied
func (as *AndSpecification) IsSatisfiedBy(candidate interface{}) bool {
	return as.left.IsSatisfiedBy(candidate) && as.right.IsSatisfiedBy(candidate)
}

// And combines with another specification
func (as *AndSpecification) And(other Specification) Specification {
	return &AndSpecification{left: as, right: other}
}

// Or combines with another specification
func (as *AndSpecification) Or(other Specification) Specification {
	return &OrSpecification{left: as, right: other}
}

// Not negates the specification
func (as *AndSpecification) Not() Specification {
	return &NotSpecification{spec: as}
}

// OrSpecification combines two specifications with OR logic
type OrSpecification struct {
	left  Specification
	right Specification
}

// IsSatisfiedBy checks if either specification is satisfied
func (os *OrSpecification) IsSatisfiedBy(candidate interface{}) bool {
	return os.left.IsSatisfiedBy(candidate) || os.right.IsSatisfiedBy(candidate)
}

// And combines with another specification
func (os *OrSpecification) And(other Specification) Specification {
	return &AndSpecification{left: os, right: other}
}

// Or combines with another specification
func (os *OrSpecification) Or(other Specification) Specification {
	return &OrSpecification{left: os, right: other}
}

// Not negates the specification
func (os *OrSpecification) Not() Specification {
	return &NotSpecification{spec: os}
}

// NotSpecification negates a specification
type NotSpecification struct {
	spec Specification
}

// IsSatisfiedBy checks if the specification is not satisfied
func (ns *NotSpecification) IsSatisfiedBy(candidate interface{}) bool {
	return !ns.spec.IsSatisfiedBy(candidate)
}

// And combines with another specification
func (ns *NotSpecification) And(other Specification) Specification {
	return &AndSpecification{left: ns, right: other}
}

// Or combines with another specification
func (ns *NotSpecification) Or(other Specification) Specification {
	return &OrSpecification{left: ns, right: other}
}

// Not negates the specification (double negation)
func (ns *NotSpecification) Not() Specification {
	return ns.spec
}
