package data

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type User struct {
	ID        uint64 `gorm:"unique;primaryKey;autoIncrement"`
	FirstName string `binding:"required"`
	LastName  string `binding:"required"`
	Email     string `gorm:"NOT NULL UNIQUE" binding:"required,email"`
	Password  string `gorm:"NOT NULL" binding:"required,min=8,max=100"`
}

type Trigger struct {
	ID          uint64 `gorm:"unique;primaryKey;autoIncrement"`
	Title       string `binding:"required"`
	Description string `binding:"required"`
	Created     time.Time
	Duration    Duration `gorm:"type:bigint"`
	UserID      uint64   `gorm:"required"`
	Api         string
	Payload     string
	Type        string
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = duration
	return nil
}

func (d Duration) Value() (driver.Value, error) {
	return int64(d.Duration), nil
}

func (d *Duration) Scan(value interface{}) error {
	if val, ok := value.(int64); ok {
		d.Duration = time.Duration(val)
		return nil
	}
	return fmt.Errorf("failed to scan Duration field")
}

type Event struct {
	Type      string
	TriggerId uint64
	UserId    uint64
	Timestamp time.Time
	Payload   string
	Execution string
}
