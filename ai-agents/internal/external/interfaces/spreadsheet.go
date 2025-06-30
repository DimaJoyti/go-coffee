package interfaces

import (
	"context"
	"time"
)

// SpreadsheetProvider defines the interface for spreadsheet platforms
type SpreadsheetProvider interface {
	// Spreadsheet operations
	CreateSpreadsheet(ctx context.Context, req *CreateSpreadsheetRequest) (*Spreadsheet, error)
	GetSpreadsheet(ctx context.Context, spreadsheetID string) (*Spreadsheet, error)
	UpdateSpreadsheet(ctx context.Context, spreadsheetID string, req *UpdateSpreadsheetRequest) (*Spreadsheet, error)
	DeleteSpreadsheet(ctx context.Context, spreadsheetID string) error
	CopySpreadsheet(ctx context.Context, spreadsheetID string, req *CopySpreadsheetRequest) (*Spreadsheet, error)
	
	// Sheet operations
	CreateSheet(ctx context.Context, spreadsheetID string, req *CreateSheetRequest) (*Sheet, error)
	GetSheet(ctx context.Context, spreadsheetID, sheetID string) (*Sheet, error)
	UpdateSheet(ctx context.Context, spreadsheetID, sheetID string, req *UpdateSheetRequest) (*Sheet, error)
	DeleteSheet(ctx context.Context, spreadsheetID, sheetID string) error
	CopySheet(ctx context.Context, spreadsheetID, sheetID string, req *CopySheetRequest) (*Sheet, error)
	DuplicateSheet(ctx context.Context, spreadsheetID, sheetID string, req *DuplicateSheetRequest) (*Sheet, error)
	
	// Cell operations
	GetValues(ctx context.Context, spreadsheetID string, req *GetValuesRequest) (*ValueRange, error)
	UpdateValues(ctx context.Context, spreadsheetID string, req *UpdateValuesRequest) (*UpdateValuesResponse, error)
	BatchGetValues(ctx context.Context, spreadsheetID string, req *BatchGetValuesRequest) (*BatchGetValuesResponse, error)
	BatchUpdateValues(ctx context.Context, spreadsheetID string, req *BatchUpdateValuesRequest) (*BatchUpdateValuesResponse, error)
	AppendValues(ctx context.Context, spreadsheetID string, req *AppendValuesRequest) (*AppendValuesResponse, error)
	ClearValues(ctx context.Context, spreadsheetID string, req *ClearValuesRequest) (*ClearValuesResponse, error)
	
	// Formatting operations
	BatchUpdate(ctx context.Context, spreadsheetID string, req *BatchUpdateRequest) (*BatchUpdateResponse, error)
	FormatCells(ctx context.Context, spreadsheetID string, req *FormatCellsRequest) error
	
	// Chart operations
	CreateChart(ctx context.Context, spreadsheetID string, req *CreateChartRequest) (*Chart, error)
	UpdateChart(ctx context.Context, spreadsheetID string, chartID int, req *UpdateChartRequest) (*Chart, error)
	DeleteChart(ctx context.Context, spreadsheetID string, chartID int) error
	
	// Named range operations
	CreateNamedRange(ctx context.Context, spreadsheetID string, req *CreateNamedRangeRequest) (*NamedRange, error)
	UpdateNamedRange(ctx context.Context, spreadsheetID string, req *UpdateNamedRangeRequest) (*NamedRange, error)
	DeleteNamedRange(ctx context.Context, spreadsheetID, namedRangeID string) error
	
	// Protection operations
	AddProtectedRange(ctx context.Context, spreadsheetID string, req *AddProtectedRangeRequest) (*ProtectedRange, error)
	UpdateProtectedRange(ctx context.Context, spreadsheetID string, req *UpdateProtectedRangeRequest) (*ProtectedRange, error)
	DeleteProtectedRange(ctx context.Context, spreadsheetID string, protectedRangeID int) error
	
	// Sharing and permissions
	ShareSpreadsheet(ctx context.Context, spreadsheetID string, req *SpreadsheetShareRequest) error
	GetPermissions(ctx context.Context, spreadsheetID string) ([]*Permission, error)
	UpdatePermissions(ctx context.Context, spreadsheetID string, req *UpdatePermissionsRequest) error
	
	// Search and filter
	FindReplace(ctx context.Context, spreadsheetID string, req *FindReplaceRequest) (*FindReplaceResponse, error)
	CreateFilter(ctx context.Context, spreadsheetID string, req *CreateFilterRequest) (*Filter, error)
	UpdateFilter(ctx context.Context, spreadsheetID string, req *UpdateFilterRequest) (*Filter, error)
	DeleteFilter(ctx context.Context, spreadsheetID string, filterID int) error
	
	// Data validation
	SetDataValidation(ctx context.Context, spreadsheetID string, req *SetDataValidationRequest) error
	
	// Provider info
	GetProviderInfo() *ProviderInfo
}

// Spreadsheet represents a spreadsheet document
type Spreadsheet struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	URL             string                 `json:"url"`
	
	// Sheets
	Sheets          []*Sheet               `json:"sheets"`
	
	// Properties
	Properties      *SpreadsheetProperties `json:"properties"`
	
	// Named ranges
	NamedRanges     []*NamedRange          `json:"named_ranges,omitempty"`
	
	// Developer metadata
	DeveloperMetadata []*DeveloperMetadata `json:"developer_metadata,omitempty"`
	
	// Timestamps
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	
	// Owner and sharing
	OwnerEmail      string                 `json:"owner_email"`
	Permissions     []*Permission          `json:"permissions,omitempty"`
}

// Sheet represents a sheet within a spreadsheet
type Sheet struct {
	ID              int                    `json:"id"`
	Title           string                 `json:"title"`
	Index           int                    `json:"index"`
	SheetType       SheetType              `json:"sheet_type"`
	
	// Properties
	Properties      *SheetProperties       `json:"properties"`
	
	// Data
	Data            []*GridData            `json:"data,omitempty"`
	
	// Charts
	Charts          []*Chart               `json:"charts,omitempty"`
	
	// Protected ranges
	ProtectedRanges []*ProtectedRange      `json:"protected_ranges,omitempty"`
	
	// Basic filter
	BasicFilter     *BasicFilter           `json:"basic_filter,omitempty"`
	
	// Filter views
	FilterViews     []*FilterView          `json:"filter_views,omitempty"`
	
	// Conditional formatting
	ConditionalFormats []*ConditionalFormatRule `json:"conditional_formats,omitempty"`
	
	// Data validation rules
	DataValidation  []*DataValidationRule  `json:"data_validation,omitempty"`
}

// SpreadsheetProperties contains properties of a spreadsheet
type SpreadsheetProperties struct {
	Title           string                 `json:"title"`
	Locale          string                 `json:"locale"`
	AutoRecalc      RecalculationInterval  `json:"auto_recalc"`
	TimeZone        string                 `json:"time_zone"`
	DefaultFormat   *CellFormat            `json:"default_format,omitempty"`
	IterativeCalculationSettings *IterativeCalculationSettings `json:"iterative_calculation_settings,omitempty"`
}

// SheetProperties contains properties of a sheet
type SheetProperties struct {
	Title           string                 `json:"title"`
	Index           int                    `json:"index"`
	SheetType       SheetType              `json:"sheet_type"`
	GridProperties  *GridProperties        `json:"grid_properties,omitempty"`
	Hidden          bool                   `json:"hidden"`
	TabColor        *Color                 `json:"tab_color,omitempty"`
	RightToLeft     bool                   `json:"right_to_left"`
}

// GridProperties contains properties of a grid
type GridProperties struct {
	RowCount        int                    `json:"row_count"`
	ColumnCount     int                    `json:"column_count"`
	FrozenRowCount  int                    `json:"frozen_row_count"`
	FrozenColumnCount int                  `json:"frozen_column_count"`
	HideGridlines   bool                   `json:"hide_gridlines"`
	RowGroupControlAfter bool              `json:"row_group_control_after"`
	ColumnGroupControlAfter bool           `json:"column_group_control_after"`
}

// ValueRange represents a range of values in a sheet
type ValueRange struct {
	Range           string                 `json:"range"`
	MajorDimension  Dimension              `json:"major_dimension"`
	Values          [][]interface{}        `json:"values"`
}

// GridData represents data in a grid
type GridData struct {
	StartRow        int                    `json:"start_row"`
	StartColumn     int                    `json:"start_column"`
	RowData         []*RowData             `json:"row_data"`
	RowMetadata     []*DimensionProperties `json:"row_metadata,omitempty"`
	ColumnMetadata  []*DimensionProperties `json:"column_metadata,omitempty"`
}

// RowData represents data in a row
type RowData struct {
	Values          []*CellData            `json:"values"`
}

// CellData represents data in a cell
type CellData struct {
	UserEnteredValue *ExtendedValue        `json:"user_entered_value,omitempty"`
	EffectiveValue   *ExtendedValue        `json:"effective_value,omitempty"`
	FormattedValue   string                `json:"formatted_value,omitempty"`
	UserEnteredFormat *CellFormat          `json:"user_entered_format,omitempty"`
	EffectiveFormat  *CellFormat           `json:"effective_format,omitempty"`
	Hyperlink        string                `json:"hyperlink,omitempty"`
	Note             string                `json:"note,omitempty"`
	TextFormatRuns   []*TextFormatRun      `json:"text_format_runs,omitempty"`
	DataValidation   *DataValidationRule   `json:"data_validation,omitempty"`
	PivotTable       *PivotTable           `json:"pivot_table,omitempty"`
}

// ExtendedValue represents a value in a cell
type ExtendedValue struct {
	NumberValue     *float64               `json:"number_value,omitempty"`
	StringValue     *string                `json:"string_value,omitempty"`
	BoolValue       *bool                  `json:"bool_value,omitempty"`
	FormulaValue    *string                `json:"formula_value,omitempty"`
	ErrorValue      *ErrorValue            `json:"error_value,omitempty"`
}

// CellFormat represents formatting for a cell
type CellFormat struct {
	NumberFormat    *NumberFormat          `json:"number_format,omitempty"`
	BackgroundColor *Color                 `json:"background_color,omitempty"`
	Borders         *Borders               `json:"borders,omitempty"`
	Padding         *Padding               `json:"padding,omitempty"`
	HorizontalAlignment HorizontalAlign    `json:"horizontal_alignment,omitempty"`
	VerticalAlignment VerticalAlign        `json:"vertical_alignment,omitempty"`
	WrapStrategy    WrapStrategy           `json:"wrap_strategy,omitempty"`
	TextDirection   TextDirection          `json:"text_direction,omitempty"`
	TextFormat      *TextFormat            `json:"text_format,omitempty"`
	HyperlinkDisplayType HyperlinkDisplayType `json:"hyperlink_display_type,omitempty"`
	TextRotation    *TextRotation          `json:"text_rotation,omitempty"`
}

// Chart represents a chart in a sheet
type Chart struct {
	ChartID         int                    `json:"chart_id"`
	Spec            *ChartSpec             `json:"spec"`
	Position        *EmbeddedObjectPosition `json:"position"`
}

// NamedRange represents a named range in a spreadsheet
type NamedRange struct {
	NamedRangeID    string                 `json:"named_range_id"`
	Name            string                 `json:"name"`
	Range           *GridRange             `json:"range"`
}

// ProtectedRange represents a protected range in a sheet
type ProtectedRange struct {
	ProtectedRangeID int                   `json:"protected_range_id"`
	Range           *GridRange             `json:"range"`
	NamedRangeID    string                 `json:"named_range_id,omitempty"`
	Description     string                 `json:"description,omitempty"`
	WarningOnly     bool                   `json:"warning_only"`
	RequestingUserCanEdit bool             `json:"requesting_user_can_edit"`
	UnprotectedRanges []*GridRange         `json:"unprotected_ranges,omitempty"`
	Editors         *Editors               `json:"editors,omitempty"`
}

// Permission represents sharing permissions
type Permission struct {
	ID              string                 `json:"id"`
	Type            PermissionType         `json:"type"`
	Role            PermissionRole         `json:"role"`
	EmailAddress    string                 `json:"email_address,omitempty"`
	Domain          string                 `json:"domain,omitempty"`
	AllowFileDiscovery bool                `json:"allow_file_discovery"`
	DisplayName     string                 `json:"display_name,omitempty"`
	PhotoLink       string                 `json:"photo_link,omitempty"`
	ExpirationTime  *time.Time             `json:"expiration_time,omitempty"`
	Deleted         bool                   `json:"deleted"`
}

// Request types
type CreateSpreadsheetRequest struct {
	Title           string                 `json:"title"`
	Sheets          []*CreateSheetRequest  `json:"sheets,omitempty"`
	Properties      *SpreadsheetProperties `json:"properties,omitempty"`
}

type UpdateSpreadsheetRequest struct {
	Title           *string                `json:"title,omitempty"`
	Properties      *SpreadsheetProperties `json:"properties,omitempty"`
}

type CopySpreadsheetRequest struct {
	Title           string                 `json:"title"`
	Parents         []string               `json:"parents,omitempty"`
}

type CreateSheetRequest struct {
	Title           string                 `json:"title"`
	SheetType       SheetType              `json:"sheet_type,omitempty"`
	GridProperties  *GridProperties        `json:"grid_properties,omitempty"`
	Hidden          bool                   `json:"hidden,omitempty"`
	TabColor        *Color                 `json:"tab_color,omitempty"`
	RightToLeft     bool                   `json:"right_to_left,omitempty"`
}

type UpdateSheetRequest struct {
	Title           *string                `json:"title,omitempty"`
	Hidden          *bool                  `json:"hidden,omitempty"`
	TabColor        *Color                 `json:"tab_color,omitempty"`
	GridProperties  *GridProperties        `json:"grid_properties,omitempty"`
}

type CopySheetRequest struct {
	DestinationSpreadsheetID string        `json:"destination_spreadsheet_id"`
	InsertSheetIndex int                   `json:"insert_sheet_index,omitempty"`
	NewSheetName    string                 `json:"new_sheet_name,omitempty"`
}

type DuplicateSheetRequest struct {
	InsertSheetIndex int                   `json:"insert_sheet_index,omitempty"`
	NewSheetName    string                 `json:"new_sheet_name,omitempty"`
}

type GetValuesRequest struct {
	Range           string                 `json:"range"`
	MajorDimension  Dimension              `json:"major_dimension,omitempty"`
	ValueRenderOption ValueRenderOption    `json:"value_render_option,omitempty"`
	DateTimeRenderOption DateTimeRenderOption `json:"date_time_render_option,omitempty"`
}

type UpdateValuesRequest struct {
	Range           string                 `json:"range"`
	Values          [][]interface{}        `json:"values"`
	MajorDimension  Dimension              `json:"major_dimension,omitempty"`
	ValueInputOption ValueInputOption      `json:"value_input_option,omitempty"`
	IncludeValuesInResponse bool           `json:"include_values_in_response,omitempty"`
	ResponseValueRenderOption ValueRenderOption `json:"response_value_render_option,omitempty"`
	ResponseDateTimeRenderOption DateTimeRenderOption `json:"response_date_time_render_option,omitempty"`
}

type BatchGetValuesRequest struct {
	Ranges          []string               `json:"ranges"`
	MajorDimension  Dimension              `json:"major_dimension,omitempty"`
	ValueRenderOption ValueRenderOption    `json:"value_render_option,omitempty"`
	DateTimeRenderOption DateTimeRenderOption `json:"date_time_render_option,omitempty"`
}

type BatchUpdateValuesRequest struct {
	ValueInputOption ValueInputOption      `json:"value_input_option"`
	Data            []*ValueRange          `json:"data"`
	IncludeValuesInResponse bool           `json:"include_values_in_response,omitempty"`
	ResponseValueRenderOption ValueRenderOption `json:"response_value_render_option,omitempty"`
	ResponseDateTimeRenderOption DateTimeRenderOption `json:"response_date_time_render_option,omitempty"`
}

type AppendValuesRequest struct {
	Range           string                 `json:"range"`
	Values          [][]interface{}        `json:"values"`
	MajorDimension  Dimension              `json:"major_dimension,omitempty"`
	ValueInputOption ValueInputOption      `json:"value_input_option,omitempty"`
	InsertDataOption InsertDataOption      `json:"insert_data_option,omitempty"`
	IncludeValuesInResponse bool           `json:"include_values_in_response,omitempty"`
	ResponseValueRenderOption ValueRenderOption `json:"response_value_render_option,omitempty"`
	ResponseDateTimeRenderOption DateTimeRenderOption `json:"response_date_time_render_option,omitempty"`
}

type ClearValuesRequest struct {
	Range           string                 `json:"range"`
}

// Response types
type UpdateValuesResponse struct {
	SpreadsheetID   string                 `json:"spreadsheet_id"`
	UpdatedRange    string                 `json:"updated_range"`
	UpdatedRows     int                    `json:"updated_rows"`
	UpdatedColumns  int                    `json:"updated_columns"`
	UpdatedCells    int                    `json:"updated_cells"`
	UpdatedData     *ValueRange            `json:"updated_data,omitempty"`
}

type BatchGetValuesResponse struct {
	SpreadsheetID   string                 `json:"spreadsheet_id"`
	ValueRanges     []*ValueRange          `json:"value_ranges"`
}

type BatchUpdateValuesResponse struct {
	SpreadsheetID   string                 `json:"spreadsheet_id"`
	TotalUpdatedRows int                   `json:"total_updated_rows"`
	TotalUpdatedColumns int                `json:"total_updated_columns"`
	TotalUpdatedCells int                  `json:"total_updated_cells"`
	TotalUpdatedSheets int                 `json:"total_updated_sheets"`
	Responses       []*UpdateValuesResponse `json:"responses"`
}

type AppendValuesResponse struct {
	SpreadsheetID   string                 `json:"spreadsheet_id"`
	TableRange      string                 `json:"table_range"`
	Updates         *UpdateValuesResponse  `json:"updates"`
}

type ClearValuesResponse struct {
	SpreadsheetID   string                 `json:"spreadsheet_id"`
	ClearedRange    string                 `json:"cleared_range"`
}

// Enums
type SheetType string

const (
	SheetTypeGrid   SheetType = "GRID"
	SheetTypeObject SheetType = "OBJECT"
)

type Dimension string

const (
	DimensionRows    Dimension = "ROWS"
	DimensionColumns Dimension = "COLUMNS"
)

type ValueRenderOption string

const (
	ValueRenderFormatted      ValueRenderOption = "FORMATTED_VALUE"
	ValueRenderUnformatted    ValueRenderOption = "UNFORMATTED_VALUE"
	ValueRenderFormula        ValueRenderOption = "FORMULA"
)

type DateTimeRenderOption string

const (
	DateTimeRenderSerial      DateTimeRenderOption = "SERIAL_NUMBER"
	DateTimeRenderFormatted   DateTimeRenderOption = "FORMATTED_STRING"
)

type ValueInputOption string

const (
	ValueInputRaw             ValueInputOption = "RAW"
	ValueInputUserEntered     ValueInputOption = "USER_ENTERED"
)

type InsertDataOption string

const (
	InsertDataOverwrite       InsertDataOption = "OVERWRITE"
	InsertDataInsertRows      InsertDataOption = "INSERT_ROWS"
)

type RecalculationInterval string

const (
	RecalcOnChange            RecalculationInterval = "ON_CHANGE"
	RecalcMinute              RecalculationInterval = "MINUTE"
	RecalcHour                RecalculationInterval = "HOUR"
)

type HorizontalAlign string

const (
	HorizontalAlignLeft       HorizontalAlign = "LEFT"
	HorizontalAlignCenter     HorizontalAlign = "CENTER"
	HorizontalAlignRight      HorizontalAlign = "RIGHT"
)

type VerticalAlign string

const (
	VerticalAlignTop          VerticalAlign = "TOP"
	VerticalAlignMiddle       VerticalAlign = "MIDDLE"
	VerticalAlignBottom       VerticalAlign = "BOTTOM"
)

type WrapStrategy string

const (
	WrapOverflow              WrapStrategy = "OVERFLOW_CELL"
	WrapLegacy                WrapStrategy = "LEGACY_WRAP"
	WrapClip                  WrapStrategy = "CLIP"
	WrapWrap                  WrapStrategy = "WRAP"
)

type TextDirection string

const (
	TextDirectionLTR          TextDirection = "LEFT_TO_RIGHT"
	TextDirectionRTL          TextDirection = "RIGHT_TO_LEFT"
)

type HyperlinkDisplayType string

const (
	HyperlinkDisplayLinked    HyperlinkDisplayType = "LINKED"
	HyperlinkDisplayPlainText HyperlinkDisplayType = "PLAIN_TEXT"
)

type PermissionType string

const (
	PermissionTypeUser        PermissionType = "user"
	PermissionTypeGroup       PermissionType = "group"
	PermissionTypeDomain      PermissionType = "domain"
	PermissionTypeAnyone      PermissionType = "anyone"
)

type PermissionRole string

const (
	PermissionRoleOwner       PermissionRole = "owner"
	PermissionRoleOrganizer   PermissionRole = "organizer"
	PermissionRoleFileOrganizer PermissionRole = "fileOrganizer"
	PermissionRoleWriter      PermissionRole = "writer"
	PermissionRoleCommenter   PermissionRole = "commenter"
	PermissionRoleReader      PermissionRole = "reader"
)

// Additional types (simplified for brevity)
type Color struct {
	Red   float32 `json:"red"`
	Green float32 `json:"green"`
	Blue  float32 `json:"blue"`
	Alpha float32 `json:"alpha"`
}

type NumberFormat struct {
	Type    string `json:"type"`
	Pattern string `json:"pattern"`
}

type Borders struct {
	Top    *Border `json:"top,omitempty"`
	Bottom *Border `json:"bottom,omitempty"`
	Left   *Border `json:"left,omitempty"`
	Right  *Border `json:"right,omitempty"`
}

type Border struct {
	Style BorderStyle `json:"style"`
	Width int         `json:"width"`
	Color *Color      `json:"color"`
}

type BorderStyle string

const (
	BorderStyleNone   BorderStyle = "NONE"
	BorderStyleSolid  BorderStyle = "SOLID"
	BorderStyleDotted BorderStyle = "DOTTED"
	BorderStyleDashed BorderStyle = "DASHED"
)

type Padding struct {
	Top    int `json:"top"`
	Right  int `json:"right"`
	Bottom int `json:"bottom"`
	Left   int `json:"left"`
}

type TextFormat struct {
	ForegroundColor *Color `json:"foreground_color,omitempty"`
	FontFamily      string `json:"font_family,omitempty"`
	FontSize        int    `json:"font_size,omitempty"`
	Bold            bool   `json:"bold,omitempty"`
	Italic          bool   `json:"italic,omitempty"`
	Strikethrough   bool   `json:"strikethrough,omitempty"`
	Underline       bool   `json:"underline,omitempty"`
}

type TextRotation struct {
	Angle    int  `json:"angle,omitempty"`
	Vertical bool `json:"vertical,omitempty"`
}

type GridRange struct {
	SheetID          int `json:"sheet_id"`
	StartRowIndex    int `json:"start_row_index"`
	EndRowIndex      int `json:"end_row_index"`
	StartColumnIndex int `json:"start_column_index"`
	EndColumnIndex   int `json:"end_column_index"`
}

// Placeholder types for complex structures
type ChartSpec struct{}
type EmbeddedObjectPosition struct{}
type DimensionProperties struct{}
type TextFormatRun struct{}
type ErrorValue struct{}
type PivotTable struct{}
type BasicFilter struct{}
type FilterView struct{}
type ConditionalFormatRule struct{}
type DataValidationRule struct{}
type DeveloperMetadata struct{}
type Editors struct{}
type IterativeCalculationSettings struct{}
type BatchUpdateRequest struct{}
type BatchUpdateResponse struct{}
type FormatCellsRequest struct{}
type CreateChartRequest struct{}
type UpdateChartRequest struct{}
type CreateNamedRangeRequest struct{}
type UpdateNamedRangeRequest struct{}
type AddProtectedRangeRequest struct{}
type UpdateProtectedRangeRequest struct{}
type SpreadsheetShareRequest struct{}
type UpdatePermissionsRequest struct{}
type FindReplaceRequest struct{}
type FindReplaceResponse struct{}
type CreateFilterRequest struct{}
type UpdateFilterRequest struct{}
type Filter struct{}
type SetDataValidationRequest struct{}
