package validec

import (
	"reflect"
	"time"
)

type CustomTime struct {
	time.Time
}

func (c *CustomTime) Decode(value string) error {
	parsedTime, err := time.Parse("2006-01-02", value)
	if err != nil {
		return err
	}
	c.Time = parsedTime
	return nil
}

func timeConverter(value string) reflect.Value {
	parsedTime, err := time.Parse("2006-01-02", value)
	if err != nil {
		return reflect.ValueOf(time.Time{})
	}
	return reflect.ValueOf(parsedTime)
}
