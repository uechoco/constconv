package testdata

// This constant definitions are the part of googleapis' type
// see more: https://github.com/googleapis/go-genproto/blob/master/googleapis/type/dayofweek/dayofweek.pb.go

// Represents a day of the week.
type DayOfWeek int32

const (
	// The day of the week is unspecified.
	DayOfWeek_DAY_OF_WEEK_UNSPECIFIED DayOfWeek = 0
	// Monday
	DayOfWeek_MONDAY DayOfWeek = 1
	// Tuesday
	DayOfWeek_TUESDAY DayOfWeek = 2
	// Wednesday
	DayOfWeek_WEDNESDAY DayOfWeek = 3
	// Thursday
	DayOfWeek_THURSDAY DayOfWeek = 4
	// Friday
	DayOfWeek_FRIDAY DayOfWeek = 5
	// Saturday
	DayOfWeek_SATURDAY DayOfWeek = 6
	// Sunday
	DayOfWeek_SUNDAY DayOfWeek = 7
)

// go:generate constconv -type=DayOfWeek -template=dayofweek.tmpl -data="package=dayofweek;typename=DayOfWeek" testdata/dayofweek.go