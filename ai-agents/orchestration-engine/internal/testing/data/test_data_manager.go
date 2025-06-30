package data

import (
	"sync"
	"time"
)

// TestDataManager provides comprehensive test data management
type TestDataManager struct {
	generators map[string]*DataGenerator
	fixtures   map[string]*TestFixture
	seeders    map[string]*DatabaseSeeder
	cleaners   map[string]*DataCleaner
	snapshots  map[string]*DataSnapshot
	config     *TestDataConfig
	logger     TestLogger
	mutex      sync.RWMutex
}

// DataGenerator generates test data based on schemas
type DataGenerator struct {
	Name         string                 `json:"name"`
	Type         GeneratorType          `json:"type"`
	Schema       *DataSchema            `json:"schema"`
	Rules        []*GenerationRule      `json:"rules"`
	Templates    map[string]interface{} `json:"templates"`
	Constraints  *DataConstraints       `json:"constraints"`
	OutputFormat string                 `json:"output_format"`
	Count        int                    `json:"count"`
	Seed         int64                  `json:"seed"`
	Enabled      bool                   `json:"enabled"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedAt    time.Time              `json:"created_at"`
}

// TestFixture represents test data fixtures
type TestFixture struct {
	Name         string                   `json:"name"`
	Description  string                   `json:"description"`
	Type         FixtureType              `json:"type"`
	Data         []map[string]interface{} `json:"data"`
	Schema       *DataSchema              `json:"schema"`
	Dependencies []string                 `json:"dependencies"`
	LoadOrder    int                      `json:"load_order"`
	Scope        FixtureScope             `json:"scope"`
	Cleanup      bool                     `json:"cleanup"`
	Enabled      bool                     `json:"enabled"`
	Metadata     map[string]interface{}   `json:"metadata"`
	CreatedAt    time.Time                `json:"created_at"`
}

// DatabaseSeeder seeds databases with test data
type DatabaseSeeder struct {
	Name             string                 `json:"name"`
	DatabaseType     string                 `json:"database_type"`
	ConnectionString string                 `json:"connection_string"`
	Tables           []*TableSeeder         `json:"tables"`
	Scripts          []string               `json:"scripts"`
	Fixtures         []string               `json:"fixtures"`
	CleanupOrder     []string               `json:"cleanup_order"`
	Enabled          bool                   `json:"enabled"`
	Metadata         map[string]interface{} `json:"metadata"`
	CreatedAt        time.Time              `json:"created_at"`
}

// DataCleaner cleans up test data
type DataCleaner struct {
	Name      string                 `json:"name"`
	Type      CleanerType            `json:"type"`
	Target    string                 `json:"target"`
	Rules     []*CleanupRule         `json:"rules"`
	Schedule  *CleanupSchedule       `json:"schedule"`
	Enabled   bool                   `json:"enabled"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
}

// DataSnapshot captures data state for testing
type DataSnapshot struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        SnapshotType           `json:"type"`
	Source      string                 `json:"source"`
	Data        map[string]interface{} `json:"data"`
	Checksum    string                 `json:"checksum"`
	Compressed  bool                   `json:"compressed"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Supporting types
type GeneratorType string

const (
	GeneratorTypeRandom     GeneratorType = "random"
	GeneratorTypeSequential GeneratorType = "sequential"
	GeneratorTypeTemplate   GeneratorType = "template"
	GeneratorTypeFaker      GeneratorType = "faker"
	GeneratorTypeCustom     GeneratorType = "custom"
)

type FixtureType string

const (
	FixtureTypeJSON   FixtureType = "json"
	FixtureTypeYAML   FixtureType = "yaml"
	FixtureTypeCSV    FixtureType = "csv"
	FixtureTypeSQL    FixtureType = "sql"
	FixtureTypeCustom FixtureType = "custom"
)

type FixtureScope string

const (
	FixtureScopeTest    FixtureScope = "test"
	FixtureScopeSuite   FixtureScope = "suite"
	FixtureScopePackage FixtureScope = "package"
	FixtureScopeGlobal  FixtureScope = "global"
)

type CleanerType string

const (
	CleanerTypeDatabase CleanerType = "database"
	CleanerTypeFile     CleanerType = "file"
	CleanerTypeMemory   CleanerType = "memory"
	CleanerTypeAPI      CleanerType = "api"
)

type SnapshotType string

const (
	SnapshotTypeDatabase SnapshotType = "database"
	SnapshotTypeFile     SnapshotType = "file"
	SnapshotTypeMemory   SnapshotType = "memory"
	SnapshotTypeAPI      SnapshotType = "api"
)

// Schema and rule types
type DataSchema struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Fields      []*FieldSchema         `json:"fields"`
	Relations   []*RelationSchema      `json:"relations"`
	Constraints []*SchemaConstraint    `json:"constraints"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type FieldSchema struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Required    bool                   `json:"required"`
	Unique      bool                   `json:"unique"`
	Default     interface{}            `json:"default"`
	Constraints []*FieldConstraint     `json:"constraints"`
	Generator   *FieldGenerator        `json:"generator"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type RelationSchema struct {
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Cardinality string                 `json:"cardinality"`
	Metadata    map[string]interface{} `json:"metadata"`
}

type SchemaConstraint struct {
	Name       string                 `json:"name"`
	Type       string                 `json:"type"`
	Expression string                 `json:"expression"`
	Message    string                 `json:"message"`
	Metadata   map[string]interface{} `json:"metadata"`
}

type FieldConstraint struct {
	Type    string      `json:"type"`
	Value   interface{} `json:"value"`
	Message string      `json:"message"`
}

type FieldGenerator struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
	Template   string                 `json:"template"`
	Function   func() interface{}     `json:"-"`
}

type GenerationRule struct {
	Name       string                 `json:"name"`
	Condition  string                 `json:"condition"`
	Action     string                 `json:"action"`
	Parameters map[string]interface{} `json:"parameters"`
	Priority   int                    `json:"priority"`
	Enabled    bool                   `json:"enabled"`
}

type DataConstraints struct {
	MinRecords   int                    `json:"min_records"`
	MaxRecords   int                    `json:"max_records"`
	UniqueFields []string               `json:"unique_fields"`
	References   []*ReferenceConstraint `json:"references"`
	Custom       []*CustomConstraint    `json:"custom"`
}

type ReferenceConstraint struct {
	Field    string `json:"field"`
	Table    string `json:"table"`
	Column   string `json:"column"`
	OnDelete string `json:"on_delete"`
	OnUpdate string `json:"on_update"`
}

type CustomConstraint struct {
	Name       string                 `json:"name"`
	Expression string                 `json:"expression"`
	Validator  func(interface{}) bool `json:"-"`
	Message    string                 `json:"message"`
}

// Database seeding types
type TableSeeder struct {
	Name      string                   `json:"name"`
	Table     string                   `json:"table"`
	Data      []map[string]interface{} `json:"data"`
	Generator string                   `json:"generator"`
	Count     int                      `json:"count"`
	Truncate  bool                     `json:"truncate"`
	Order     int                      `json:"order"`
	Enabled   bool                     `json:"enabled"`
}

// Cleanup types
type CleanupRule struct {
	Name       string                 `json:"name"`
	Pattern    string                 `json:"pattern"`
	Action     string                 `json:"action"`
	Condition  string                 `json:"condition"`
	Parameters map[string]interface{} `json:"parameters"`
	Enabled    bool                   `json:"enabled"`
}

type CleanupSchedule struct {
	Type      string        `json:"type"`
	Interval  time.Duration `json:"interval"`
	Time      string        `json:"time"`
	Condition string        `json:"condition"`
	Enabled   bool          `json:"enabled"`
}

// Configuration types
type TestDataConfig struct {
	DefaultGenerator    string                     `json:"default_generator"`
	DefaultFixtureScope FixtureScope               `json:"default_fixture_scope"`
	DataPath            string                     `json:"data_path"`
	FixturePath         string                     `json:"fixture_path"`
	SnapshotPath        string                     `json:"snapshot_path"`
	AutoCleanup         bool                       `json:"auto_cleanup"`
	CleanupOnFailure    bool                       `json:"cleanup_on_failure"`
	EnableSnapshots     bool                       `json:"enable_snapshots"`
	EnableSeeding       bool                       `json:"enable_seeding"`
	ParallelGeneration  bool                       `json:"parallel_generation"`
	CacheEnabled        bool                       `json:"cache_enabled"`
	CacheTTL            time.Duration              `json:"cache_ttl"`
	Databases           map[string]*DatabaseConfig `json:"databases"`
	Environment         map[string]string          `json:"environment"`
}

type DatabaseConfig struct {
	Type             string        `json:"type"`
	Host             string        `json:"host"`
	Port             int           `json:"port"`
	Database         string        `json:"database"`
	Username         string        `json:"username"`
	Password         string        `json:"password"`
	ConnectionString string        `json:"connection_string"`
	MaxConnections   int           `json:"max_connections"`
	Timeout          time.Duration `json:"timeout"`
}

// TestLogger interface
type TestLogger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, err error, args ...interface{})
}

// NewTestDataManager creates a new test data manager
func NewTestDataManager(config *TestDataConfig, logger TestLogger) *TestDataManager {
	if config == nil {
		config = DefaultTestDataConfig()
	}

	return &TestDataManager{
		generators: make(map[string]*DataGenerator),
		fixtures:   make(map[string]*TestFixture),
		seeders:    make(map[string]*DatabaseSeeder),
		cleaners:   make(map[string]*DataCleaner),
		snapshots:  make(map[string]*DataSnapshot),
		config:     config,
		logger:     logger,
	}
}

// DefaultTestDataConfig returns default test data configuration
func DefaultTestDataConfig() *TestDataConfig {
	return &TestDataConfig{
		DefaultGenerator:    "random",
		DefaultFixtureScope: FixtureScopeTest,
		DataPath:            "./test-data",
		FixturePath:         "./test-fixtures",
		SnapshotPath:        "./test-snapshots",
		AutoCleanup:         true,
		CleanupOnFailure:    true,
		EnableSnapshots:     true,
		EnableSeeding:       true,
		ParallelGeneration:  true,
		CacheEnabled:        true,
		CacheTTL:            1 * time.Hour,
		Databases:           make(map[string]*DatabaseConfig),
		Environment:         make(map[string]string),
	}
}

// CreateDataGenerator creates a new data generator
func (tdm *TestDataManager) CreateDataGenerator(name string, generatorType GeneratorType) *DataGenerator {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()

	generator := &DataGenerator{
		Name:         name,
		Type:         generatorType,
		Schema:       &DataSchema{},
		Rules:        make([]*GenerationRule, 0),
		Templates:    make(map[string]interface{}),
		Constraints:  &DataConstraints{},
		OutputFormat: "json",
		Count:        100,
		Seed:         time.Now().UnixNano(),
		Enabled:      true,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}

	tdm.generators[name] = generator
	tdm.logger.Info("Data generator created", "name", name, "type", generatorType)

	return generator
}

// CreateTestFixture creates a new test fixture
func (tdm *TestDataManager) CreateTestFixture(name, description string, fixtureType FixtureType) *TestFixture {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()

	fixture := &TestFixture{
		Name:         name,
		Description:  description,
		Type:         fixtureType,
		Data:         make([]map[string]interface{}, 0),
		Schema:       &DataSchema{},
		Dependencies: make([]string, 0),
		LoadOrder:    0,
		Scope:        tdm.config.DefaultFixtureScope,
		Cleanup:      tdm.config.AutoCleanup,
		Enabled:      true,
		Metadata:     make(map[string]interface{}),
		CreatedAt:    time.Now(),
	}

	tdm.fixtures[name] = fixture
	tdm.logger.Info("Test fixture created", "name", name, "type", fixtureType)

	return fixture
}

// CreateDatabaseSeeder creates a new database seeder
func (tdm *TestDataManager) CreateDatabaseSeeder(name, dbType, connectionString string) *DatabaseSeeder {
	tdm.mutex.Lock()
	defer tdm.mutex.Unlock()

	seeder := &DatabaseSeeder{
		Name:             name,
		DatabaseType:     dbType,
		ConnectionString: connectionString,
		Tables:           make([]*TableSeeder, 0),
		Scripts:          make([]string, 0),
		Fixtures:         make([]string, 0),
		CleanupOrder:     make([]string, 0),
		Enabled:          true,
		Metadata:         make(map[string]interface{}),
		CreatedAt:        time.Now(),
	}

	tdm.seeders[name] = seeder
	tdm.logger.Info("Database seeder created", "name", name, "type", dbType)

	return seeder
}
