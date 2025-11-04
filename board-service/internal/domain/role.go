package domain

type Role struct {
	BaseModel
	Name        string `gorm:"type:varchar(50);uniqueIndex;not null" json:"name"`
	Level       int    `gorm:"not null" json:"level"`
	Description string `gorm:"type:text" json:"description"`
}

func (Role) TableName() string {
	return "roles"
}
