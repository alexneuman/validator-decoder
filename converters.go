package validec

import (
	"reflect"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
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

func PgTypeTextConverter(value string) reflect.Value {
	p := pgtype.Text{String: value}
	return reflect.ValueOf(p)
}

func PgTypeInt2Converter(value string) reflect.Value {
	i, err := strconv.Atoi(value)

	if err != nil {
		i = 0
	}
	i16 := int16(i)
	p := pgtype.Int2{Int16: i16}
	return reflect.ValueOf(p)
}

func PgTypeInt4Converter(value string) reflect.Value {
	i, err := strconv.Atoi(value)

	if err != nil {
		i = 0
	}
	i32 := int32(i)
	p := pgtype.Int4{Int32: i32}
	return reflect.ValueOf(p)
}

func PgTypeDateTimeConverter(value string) reflect.Value {
	parsedTime, err := time.Parse("2006-01-02", value)
	var p pgtype.Date
	if err == nil {
		p.Time = parsedTime
	}

	return reflect.ValueOf(p)
}
