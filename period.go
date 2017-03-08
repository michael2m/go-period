package planner

import (
	"bytes"
	"fmt"
	"html/template"
	"regexp"
	"strconv"
	"time"
)

var (
	tmplPeriod = template.Must(template.New("duration").Parse(`P{{if .Years}}{{.Years}}Y{{end}}{{if .Months}}{{.Months}}M{{end}}{{if .Weeks}}{{.Weeks}}W{{end}}{{if .Days}}{{.Days}}D{{end}}{{if .HasTimePart}}T{{end }}{{if .Hours}}{{.Hours}}H{{end}}{{if .Minutes}}{{.Minutes}}M{{end}}{{if .Seconds}}{{.Seconds}}S{{end}}`))

	regexpPeriod = regexp.MustCompile(`^P((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<day>\d+)D)?(T((?P<hour>\d+)H)?((?P<minute>\d+)M)?((?P<second>\d+)S)?)?$`)
	regexpWeek   = regexp.MustCompile(`^P((?P<week>\d+)W)$`)

	errPeriodOverflow  = fmt.Errorf("Period overflow")
	errPeriodBadFormat = fmt.Errorf("Bad period format")
)

// Period ISO-8601, independent of calendar (unless at moment it is applied to a timestamp with calendar reference).
type Period struct {
	Years   int
	Months  int
	Weeks   int
	Days    int
	Hours   int
	Minutes int
	Seconds int
}

func (period Period) String() string {
	var buff bytes.Buffer

	err := tmplPeriod.Execute(&buff, &period)
	if err != nil {
		panic(err)
	}

	return buff.String()
}

// DaysInYear returns the number of days in given given year.
func DaysInYear(year int, loc *time.Location) int {
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, loc)
	start := time.Date(year, 1, 1, 0, 0, 0, 0, loc)
	duration := end.Sub(start)

	return int(duration / (time.Hour * 24))
}

// DaysInMonth returns the number of days in given month of given year.
func DaysInMonth(year int, month time.Month, loc *time.Location) int {
	end := time.Date(year, month+1, 1, 0, 0, 0, 0, loc)
	start := time.Date(year, month, 1, 0, 0, 0, 0, loc)
	duration := end.Sub(start)

	return int(duration / (time.Hour * 24))
}

// Between returns the period between two timestamps in a given location (hence DST may be noticeable).
// It first determines the date period between two timestamps (i.e. years, months, days).
// The start date is adjusted by the date period and then the time period between the timestamps are determined (i.e. hours, minutes, seconds).
// Finally the date and time parts are adjusted by flowing components, e.g. negative seconds are adjusted for positive minutes (and vice versa).
//
// Example:
// 2016-02-28 and 2016-03-31 are +(1 month and 3 days) or -(1 month and 3 days) apart, note the inclusion of a leap day 2016
// 2017-03-26 00:00:00 and 2016-03-26 06:00:00 Europe/Amsterdam are +(5 hours) or -(5 hours) apart, note the DST transition
// 2017-03-26 00:00:00 and 2016-03-27 02:00:00 Europe/Amsterdam are +(1 day and 2 hours) or -(1 day and 2 hours) apart, note the DST insensitivity
// 2017-03-31 04:15:00 and 2016-03-31 05:10:00 are +(55 minutes) or -(55 minutes) apart, note the partial hour (similar in case of partial minute)
func Between(start, end time.Time, loc *time.Location) (period Period) {
	if start.Equal(end) {
		return period
	}

	start = start.In(loc)
	end = end.In(loc)

	year1, month1, day1 := start.Date()
	year2, month2, day2 := end.Date()

	years := year2 - year1
	months := month2 - month1
	days := day2 - day1

	hour, minute, second := start.Clock()
	adjusted := time.Date(year1+years, month1+months, day1+days, hour, minute, second, 0, loc)
	diff := end.Sub(adjusted)
	hours := int(diff / time.Hour)
	minutes := int((diff % time.Hour) / time.Minute)
	seconds := int(((diff % time.Hour) % time.Minute) / time.Second)

	if minutes > 0 && seconds < 0 {
		seconds += 60
		minutes--
	} else if minutes < 0 && seconds > 0 {
		seconds -= 24
		minutes++
	}

	if hours > 0 && minutes < 0 {
		minutes += 60
		hours--
	} else if hours < 0 && minutes > 0 {
		minutes -= 60
		hours++
	}

	if days > 0 && hours < 0 {
		hours += 24
		days--
	} else if days < 0 && hours > 0 {
		hours -= 24
		days++
	}

	if years > 0 && months < 0 {
		months += 12
		years--
	} else if years < 0 && months > 0 {
		months -= 12
		years++
	}

	return Period{Years: years, Months: int(months), Weeks: 0, Days: days, Hours: hours, Minutes: minutes, Seconds: seconds}
}

// FromDuration returns period, floored to nearest second.
func FromDuration(duration time.Duration) (*Period, error) {
	seconds := duration / time.Second
	maxSeconds := time.Duration(1<<(strconv.IntSize-1) - 1)
	if seconds > maxSeconds {
		return nil, errPeriodOverflow
	}

	period := Period{Seconds: int(seconds)}.Normalize()
	return &period, nil
}

// FromString returns period parsed from string.
func FromString(str string) (*Period, error) {
	var match []string
	var re *regexp.Regexp

	match = regexpWeek.FindStringSubmatch(str)
	re = regexpWeek

	if match == nil {
		match = regexpPeriod.FindStringSubmatch(str)
		re = regexpPeriod

		if match == nil {
			return nil, errPeriodBadFormat
		}
	}

	var period Period

	for i, name := range re.SubexpNames() {
		part := match[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return nil, err
		}

		switch name {
		case "year":
			period.Years = val
		case "month":
			period.Months = val
		case "week":
			period.Weeks = val
		case "day":
			period.Days = val
		case "hour":
			period.Hours = val
		case "minute":
			period.Minutes = val
		case "second":
			period.Seconds = val
		}
	}

	return &period, nil
}

// HasDatePart returns true iff any of years, months, weeks or days is not 0.
func (period Period) HasDatePart() bool {
	return period.Years != 0 || period.Months != 0 || period.Weeks != 0 || period.Days != 0
}

// HasTimePart returns true iff any of hours, minutes or seconds is not 0.
func (period Period) HasTimePart() bool {
	return period.Hours != 0 || period.Minutes != 0 || period.Seconds != 0
}

// Normalize period(seconds to minutes, minutes to hours, hours to days, weeks to days, months to years).
func (period Period) Normalize() Period {
	// years unaffected
	// every 12 months are replaced by 1 year
	// every week is replaced by 7 days
	// cannot flow from days to year or month as it depends on calendar (i.e. leap year or month in year)
	// every 24 hours are replaced by 1 day
	// every 60 minutes are replaced by 1 hour
	// every 60 seconds are replaced by 1 minute

	seconds := period.Seconds % 60
	minutes := period.Seconds/60 + period.Minutes

	minutes, hours := minutes%60, minutes/60+period.Hours

	hours, days := hours%24, hours/24+period.Days+period.Weeks*7

	months, years := period.Months%12, period.Months/12+period.Years

	return Period{years, months, 0, days, hours, minutes, seconds}
}

// Apply period to timestamp and return result.
func (period Period) Apply(t time.Time) time.Time {
	year, month, day := t.Date()
	hour, minute, second, nanos := t.Hour(), t.Minute(), t.Second(), t.Nanosecond()

	duration := time.Hour*time.Duration(period.Hours) + time.Minute*time.Duration(period.Minutes) + time.Second*time.Duration(period.Seconds)

	// handles DST transitions appropriately by adding duration, instead of adding hours/minutes/seconds directly upon date construction
	result := time.Date(year+period.Years, month+time.Month(period.Months), day+period.Days, hour, minute, second, nanos, t.Location()).Add(duration)
	return result
}
