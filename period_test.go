package planner

import (
	"testing"
	"time"
)

func TestDaysInYear(t *testing.T) {
	if 366 != DaysInYear(2016, time.UTC) {
		t.Fatal("Unexpected number of days")
	}

	if 365 != DaysInYear(2017, time.UTC) {
		t.Fatal("Unexpected number of days")
	}
}

func TestDaysInMonth(t *testing.T) {
	if 29 != DaysInMonth(2016, 2, time.UTC) {
		t.Fatal("Unexpected number of days")
	}

	if 28 != DaysInMonth(2017, 2, time.UTC) {
		t.Fatal("Unexpected number of days")
	}

	if 31 != DaysInMonth(2017, 8, time.UTC) {
		t.Fatal("Unexpected number of days")
	}

	if 30 != DaysInMonth(2017, 4, time.UTC) {
		t.Fatal("Unexpected number of days")
	}
}

func TestFromString(t *testing.T) {
	result, err := FromString("DT1S")
	if err == nil || result != nil {
		t.Fatal()
	}

	result, err = FromString("P1S")
	if err == nil || result != nil {
		t.Fatal()
	}

	result, err = FromString("P1Q")
	if err == nil || result != nil {
		t.Fatal()
	}

	result, err = FromString("PT1S")
	if err != nil || *result != (Period{Seconds: 1}) {
		t.Fatal()
	}

	result, err = FromString("PT2M")
	if err != nil || *result != (Period{Minutes: 2}) {
		t.Fatal()
	}

	result, err = FromString("PT3H")
	if err != nil || *result != (Period{Hours: 3}) {
		t.Fatal()
	}

	result, err = FromString("P4D")
	if err != nil || *result != (Period{Days: 4}) {
		t.Fatal()
	}

	result, err = FromString("P5M")
	if err != nil || *result != (Period{Months: 5}) {
		t.Fatal()
	}

	result, err = FromString("P6Y")
	if err != nil || *result != (Period{Years: 6}) {
		t.Fatal()
	}

	result, err = FromString("P6Y5M4DT3H2M1S")
	if err != nil || *result != (Period{Years: 6, Months: 5, Days: 4, Hours: 3, Minutes: 2, Seconds: 1}) {
		t.Fatal()
	}

	result, err = FromString("P7W")
	if err != nil || *result != (Period{Weeks: 7}) {
		t.Fatal()
	}
}

func TestFromDuration(t *testing.T) {
	result, err := FromDuration(0)
	if err != nil || *result != (Period{}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	result, err = FromDuration(time.Second)
	if err != nil || *result != (Period{Seconds: 1}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	result, err = FromDuration(-time.Second)
	if err != nil || *result != (Period{Seconds: -1}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	result, err = FromDuration(27*time.Hour + 74*time.Minute + 63*time.Second)
	if err != nil || *result != (Period{Days: 1, Hours: 4, Minutes: 15, Seconds: 3}) {
		t.Fatalf("Unexpected period: %s", result)
	}
}

func TestBetween(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	start := time.Date(2016, 2, 28, 0, 0, 0, 0, loc)
	end := time.Date(2016, 3, 31, 0, 0, 0, 0, loc)

	// symmetric
	result := Between(start, end, loc)
	if result != (Period{Months: 1, Days: 3}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	result = Between(end, start, loc)
	if result != (Period{Months: -1, Days: -3}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	// DST sensitive
	start = time.Date(2017, 3, 26, 0, 0, 0, 0, loc)
	end = time.Date(2017, 3, 26, 6, 0, 0, 0, loc)

	result = Between(start, end, loc)
	if result != (Period{Hours: 5}) {
		t.Fatalf("Unexpected period: %s", result)
	} else if result.Apply(start) != end {
		t.Fatalf("Unexpected apply: %s", result.Apply(start))
	}

	result = Between(end, start, loc)
	if result != (Period{Hours: -5}) {
		t.Fatalf("Unexpected period: %s", result)
	} else if result.Apply(end) != start {
		t.Fatalf("Unexpected apply: %s", result.Apply(end))
	}

	start = time.Date(2017, 3, 26, 0, 0, 0, 0, loc)
	end = time.Date(2017, 3, 27, 2, 0, 0, 0, loc)

	result = Between(start, end, loc)
	if result != (Period{Days: 1, Hours: 2}) {
		t.Fatalf("Unexpected period: %s", result)
	} else if result.Apply(start) != end {
		t.Fatalf("Unexpected apply: %s", result.Apply(start))
	}

	result = Between(end, start, loc)
	if result != (Period{Days: -1, Hours: -2}) {
		t.Fatalf("Unexpected period: %s", result)
	} else if result.Apply(end) != start {
		t.Fatalf("Unexpected apply: %s", result.Apply(end))
	}

	// part of day/hour/minute handling
	start = time.Date(2017, 3, 28, 7, 0, 0, 0, loc)
	end = time.Date(2017, 3, 29, 5, 0, 0, 0, loc)

	result = Between(end, start, loc)
	if result != (Period{Hours: -22}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	end = time.Date(2017, 3, 30, 5, 0, 0, 0, loc)

	result = Between(end, start, loc)
	if result != (Period{Days: -1, Hours: -22}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	start = time.Date(2017, 3, 28, 7, 10, 0, 0, loc)
	end = time.Date(2017, 3, 28, 8, 5, 0, 0, loc)

	result = Between(start, end, loc)
	if result != (Period{Minutes: 55}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	start = time.Date(2017, 3, 28, 7, 4, 1, 0, loc)
	end = time.Date(2017, 3, 28, 7, 5, 0, 0, loc)

	result = Between(start, end, loc)
	if result != (Period{Seconds: 59}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	// multi year, month, day, hour, minute, second and ignorance of nanos
	start = time.Date(2017, 3, 26, 7, 4, 1, 0, loc)
	end = time.Date(2019, 6, 30, 12, 10, 8, 999999999, loc)

	result = Between(start, end, loc)
	if result != (Period{Years: 2, Months: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7}) {
		t.Fatalf("Unexpected period: %s", result)
	}

	// multi year, month, day, hour, minute, second
	start = time.Date(2017, 3, 26, 7, 4, 1, 0, loc)
	end = time.Date(2016, 2, 25, 6, 3, 0, 0, loc)

	result = Between(start, end, loc)
	if result != (Period{Years: -1, Months: -1, Days: -1, Hours: -1, Minutes: -1, Seconds: -1}) {
		t.Fatalf("Unexpected period: %s", result)
	}
}

func TestPeriodHasTimePart(t *testing.T) {
	period1 := Period{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 0, Minutes: 0, Seconds: 0}
	period2 := Period{Years: 0, Months: 0, Weeks: 0, Days: 0, Hours: 1, Minutes: 2, Seconds: 3}
	period3 := Period{Years: 0, Months: 0, Weeks: 0, Days: 0, Hours: 0, Minutes: 0, Seconds: 1}
	period4 := Period{Years: 0, Months: 0, Weeks: 0, Days: 0, Hours: 0, Minutes: 1, Seconds: 0}
	period5 := Period{Years: 0, Months: 0, Weeks: 0, Days: 0, Hours: 1, Minutes: 0, Seconds: 0}
	period6 := Period{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7}

	if period1.HasTimePart() {
		t.Fatalf("Unexpected time part")
	}

	if !period2.HasTimePart() {
		t.Fatalf("Unexpected time part")
	}

	if !period3.HasTimePart() {
		t.Fatalf("Unexpected time part")
	}

	if !period4.HasTimePart() {
		t.Fatalf("Unexpected time part")
	}

	if !period5.HasTimePart() {
		t.Fatalf("Unexpected time part")
	}

	if !period6.HasTimePart() {
		t.Fatalf("Unexpected time part")
	}
}

func TestPeriodHasDatePart(t *testing.T) {
	period1 := Period{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 0, Minutes: 0, Seconds: 0}
	period2 := Period{Years: 0, Months: 0, Weeks: 0, Days: 0, Hours: 1, Minutes: 2, Seconds: 3}
	period3 := Period{Years: 1, Months: 0, Weeks: 0, Days: 0, Hours: 0, Minutes: 0, Seconds: 0}
	period4 := Period{Years: 0, Months: 1, Weeks: 0, Days: 0, Hours: 0, Minutes: 0, Seconds: 0}
	period5 := Period{Years: 0, Months: 0, Weeks: 1, Days: 0, Hours: 0, Minutes: 0, Seconds: 0}
	period6 := Period{Years: 0, Months: 0, Weeks: 0, Days: 1, Hours: 0, Minutes: 0, Seconds: 0}
	period7 := Period{Years: 1, Months: 2, Weeks: 3, Days: 4, Hours: 5, Minutes: 6, Seconds: 7}

	if !period1.HasDatePart() {
		t.Fatalf("Unexpected date part")
	}

	if period2.HasDatePart() {
		t.Fatalf("Unexpected date part")
	}

	if !period3.HasDatePart() {
		t.Fatalf("Unexpected date part")
	}

	if !period4.HasDatePart() {
		t.Fatalf("Unexpected date part")
	}

	if !period5.HasDatePart() {
		t.Fatalf("Unexpected date part")
	}

	if !period6.HasDatePart() {
		t.Fatalf("Unexpected date part")
	}

	if !period7.HasDatePart() {
		t.Fatalf("Unexpected date part")
	}
}

func TestPeriodNormalize(t *testing.T) {
	period := Period{Years: 1, Months: 15, Weeks: 2, Days: 31, Hours: 27, Minutes: 73, Seconds: 91}
	result := period.Normalize()

	if result.Seconds != 31 || result.Minutes != 14 || result.Hours != 4 || result.Days != 46 || result.Weeks != 0 || result.Months != 3 || result.Years != 2 {
		t.Fatalf("Unexpected normalization: %s", result)
	}
}

func TestPeriodApply(t *testing.T) {
	loc, _ := time.LoadLocation("Europe/Amsterdam")
	ref := time.Date(2017, 3, 7, 12, 30, 45, 999999999, time.UTC)

	period := Period{Years: 1, Months: 1, Days: 1, Hours: 1, Minutes: 1, Seconds: 1}
	result := period.Apply(ref)
	if result != time.Date(2018, 4, 8, 13, 31, 46, 999999999, time.UTC) {
		t.Fatalf("Invalid application: %s", result)
	}

	period = Period{Years: -1, Months: -1, Days: -1, Hours: -1, Minutes: -1, Seconds: -1}
	result = period.Apply(ref)
	if result != time.Date(2016, 2, 6, 11, 29, 44, 999999999, time.UTC) {
		t.Fatalf("Invalid application: %s", result)
	}

	period = Period{Years: 3, Months: 9, Days: 24, Hours: 11, Minutes: 29, Seconds: 15}
	result = period.Apply(ref)
	if result != time.Date(2021, 1, 1, 0, 0, 0, 999999999, time.UTC) {
		t.Fatalf("Invalid application: %s", result)
	}

	// beyond DST
	ref = time.Date(2017, 3, 26, 0, 0, 0, 0, loc)
	period = Period{Days: 1, Hours: 12}
	result = period.Apply(ref)
	if result != time.Date(2017, 3, 27, 12, 0, 0, 0, loc) {
		t.Fatalf("Invalid application: %s", result)
	}

	// through DST
	period = Period{Hours: 36}
	result = period.Apply(ref)
	if result != time.Date(2017, 3, 27, 13, 0, 0, 0, loc) {
		t.Fatalf("Invalid application: %s", result)
	}

	// beyond DST
	ref = time.Date(2017, 10, 29, 0, 0, 0, 0, loc)
	period = Period{Days: 1, Hours: 12}
	result = period.Apply(ref)
	if result != time.Date(2017, 10, 30, 12, 0, 0, 0, loc) {
		t.Fatalf("Invalid application: %s", result)
	}

	// through DST
	period = Period{Hours: 36}
	result = period.Apply(ref)
	if result != time.Date(2017, 10, 30, 11, 0, 0, 0, loc) {
		t.Fatalf("Invalid application: %s", result)
	}
}
