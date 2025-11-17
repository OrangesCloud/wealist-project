package domain

// FieldType represents the type of a custom field
type FieldType string

// FieldType constants
const (
	FieldTypeStage      FieldType = "stage"
	FieldTypeRole       FieldType = "role"
	FieldTypeImportance FieldType = "importance"
)

// FieldOption represents a selectable option for custom fields (stage, role, importance)
type FieldOption struct {
	BaseModel
	FieldType       FieldType `gorm:"type:varchar(50);not null;index" json:"field_type"`
	Value           string    `gorm:"type:varchar(100);not null" json:"value"`
	Label           string    `gorm:"type:varchar(200);not null" json:"label"`
	Color           string    `gorm:"type:varchar(20);not null" json:"color"`
	DisplayOrder    int       `gorm:"type:int;not null;default:0;index" json:"display_order"`
	IsSystemDefault bool      `gorm:"type:boolean;not null;default:false" json:"is_system_default"`
}

// TableName specifies the table name for FieldOption
func (FieldOption) TableName() string {
	return "field_options"
}
