package project

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// Labels 是字符串映射,持久化为 JSONB。
type Labels map[string]string

// Value 实现 driver.Valuer。
func (l Labels) Value() (driver.Value, error) {
	if len(l) == 0 {
		return nil, nil
	}
	return json.Marshal(l)
}

// Scan 实现 sql.Scanner。
func (l *Labels) Scan(src any) error {
	if src == nil {
		*l = nil
		return nil
	}
	var raw []byte
	switch v := src.(type) {
	case []byte:
		raw = v
	case string:
		raw = []byte(v)
	default:
		return errors.New("project: unsupported scan source for Labels")
	}
	if len(raw) == 0 {
		*l = nil
		return nil
	}
	m := map[string]string{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return err
	}
	*l = m
	return nil
}
