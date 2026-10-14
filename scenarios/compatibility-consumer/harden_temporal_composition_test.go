// These scenarios connect finite public occurrence calculations and instant
// membership. They do not qualify Runner timers or newer opening-hour expansion.
package compatibilityconsumer_test

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"testing"
	"time"

	calendar "github.com/faustbrian/go-calendar/v2"
	"github.com/faustbrian/go-clock/manual"
	opening "github.com/faustbrian/go-opening-hours/v3"
	openingcalendar "github.com/faustbrian/go-opening-hours/v3/adapters/calendar"
	openingtemporal "github.com/faustbrian/go-opening-hours/v3/adapters/temporal"
	scheduler "github.com/faustbrian/go-scheduler"
	temporal "github.com/faustbrian/go-temporal/v2"
	"github.com/faustbrian/go-temporal/v2/instant"
	"github.com/faustbrian/go-temporal/v2/timeofday"
)

func TestHardenPublishedFiniteDSTComposition(t *testing.T) {
	// Preload the real canonical location before preparing the independent
	// oracle. Existing same-process scenarios may already have done this; no
	// environment change or private timezone-cache reset is needed.
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}
	// This exact 1744-byte fixture is the authenticated Go 1.27.2 New York
	// TZif retained by Opening Hours R04: source-tag commit
	// 022c8636110ebac86a9b88724cda168ed713c21f, lib/time/zoneinfo.zip
	// archive SHA256 b2d18a7c8fa8142097a48c99609fb3c92db5ee98bc740294e57eab8ae9f94779.
	// The digest binds the independent oracle, not runtime-loaded host data.
	const newYorkSHA = "d7f2206b3a45989fc9ad63d558922532fa7352280d5f87176bf1db79cb1d1fa9"
	const newYorkTZif = `VFppZjIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAQAAAAEAAAAAAAAAVFppZjIA
AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAACvAAAABQAAABT/////XgPwkP////+eph5w////
/5+662D/////oIYAcP////+hms1g/////6Jl4nD/////o4Pp4P////+kaq5w/////6U1p2D/////
plPK8P////+nFYlg/////6gzrPD/////qP6l4P////+qE47w/////6reh+D/////q/Nw8P////+s
vmng/////63TUvD/////rp5L4P////+vszTw/////7B+LeD/////sZxRcP////+yZ0pg/////7N8
M3D/////tEcsYP////+1XBVw/////7YnDmD/////tzv3cP////+4BvBg/////7kb2XD/////uebS
YP////+7BPXw/////7vGtGD/////vOTX8P////+9r9Dg/////77EufD/////v4+y4P/////ApJvw
/////8FvlOD/////woR98P/////DT3bg/////8RkX/D/////xS9Y4P/////GTXxw/////8cPOuD/
////yC1ecP/////I+Fdg/////8oNQHD/////ytg5YP/////LiPBw/////9Ij9HD/////0mD74P//
///TdeTw/////9RA3eD/////1VXG8P/////WIL/g/////9c1qPD/////2ACh4P/////ZFYrw////
/9ngg+D/////2v6ncP/////bwGXg/////9zeiXD/////3amCYP/////evmtw/////9+JZGD/////
4J5NcP/////haUZg/////+J+L3D/////40koYP/////kXhFw/////+VXLuD/////5kct8P/////n
NxDg/////+gnD/D/////6Rby4P/////qBvHw/////+r21OD/////6+bT8P/////s1rbg/////+3G
tfD/////7r/TYP/////vr9Jw//////CftWD/////8Y+0cP/////yf5dg//////NvlnD/////9F95
YP/////1T3hw//////Y/W2D/////9y9acP/////4KHfg//////kPPHD/////+ghZ4P/////6+Fjw
//////voO+D//////Ng68P/////9yB3g//////64HPD//////6f/4AAAAAAAl/7wAAAAAAGH4eAA
AAAAAnfg8AAAAAADcP5gAAAAAARg/XAAAAAABVDgYAAAAAAGQN9wAAAAAAcwwmAAAAAAB40ZcAAA
AAAJEKRgAAAAAAmtlPAAAAAACvCGYAAAAAAL4IVwAAAAAAzZouAAAAAADcBncAAAAAAOuYTgAAAA
AA+pg/AAAAAAEJlm4AAAAAARiWXwAAAAABJ5SOAAAAAAE2lH8AAAAAAUWSrgAAAAABVJKfAAAAAA
FjkM4AAAAAAXKQvwAAAAABgiKWAAAAAAGQjt8AAAAAAaAgtgAAAAABryCnAAAAAAG+HtYAAAAAAc
0exwAAAAAB3Bz2AAAAAAHrHOcAAAAAAfobFgAAAAACB2APAAAAAAIYGTYAAAAAAiVeLwAAAAACNq
r+AAAAAAJDXE8AAAAAAlSpHgAAAAACYVpvAAAAAAJypz4AAAAAAn/sNwAAAAACkKVeAAAAAAKd6l
cAAAAAAq6jfgAAAAACu+h3AAAAAALNNUYAAAAAAtnmlwAAAAAC6zNmAAAAAAL35LcAAAAAAwkxhg
AAAAADFnZ/AAAAAAMnL6YAAAAAAzR0nwAAAAADRS3GAAAAAANScr8AAAAAA2Mr5gAAAAADcHDfAA
AAAAOBva4AAAAAA45u/wAAAAADn7vOAAAAAAOsbR8AAAAAA7257gAAAAADyv7nAAAAAAPbuA4AAA
AAA+j9BwAAAAAD+bYuAAAAAAQG+ycAAAAABBhH9gAAAAAEJPlHAAAAAAQ2RhYAAAAABEL3ZwAAAA
AEVEQ2AAAAAARfOo8AIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIB
AgECAQIDBAIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIB
AgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgECAQIBAgEC
AQIBAgECAQIBAgECAQIBAgH//7qeAAD//8fAAQT//7mwAAj//8fAAQz//8fAARBMTVQARURUAEVT
VABFV1QARVBUAApFU1Q1RURULE0zLjIuMCxNMTEuMS4wCg==`
	tzif, err := base64.StdEncoding.DecodeString(newYorkTZif)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(tzif)
	if len(tzif) != 1744 || hex.EncodeToString(digest[:]) != newYorkSHA {
		t.Fatal("New York TZif differs from the attributed finite corpus")
	}
	oracle, err := time.LoadLocationFromTZData("America/New_York", tzif)
	if err != nil {
		t.Fatal(err)
	}
	start, err := timeofday.New(1, 0, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	end, err := timeofday.New(4, 0, 0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	localWindow, err := timeofday.Between(start, end, temporal.ClosedOpen)
	if err != nil {
		t.Fatal(err)
	}
	window, err := openingtemporal.RangeFromInterval(localWindow)
	if err != nil {
		t.Fatal(err)
	}
	rule, err := opening.OpenRanges([]opening.Range{window}, opening.RejectOverlapAndAdjacent)
	if err != nil {
		t.Fatal(err)
	}
	weekly := map[time.Weekday]opening.DayRule{}
	for day := time.Sunday; day <= time.Saturday; day++ {
		weekly[day] = rule
	}
	openingSchedule, err := opening.NewSchedule(opening.Config{Timezone: location.String(), Weekly: weekly})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name, expression string
		after            time.Time
		want             []time.Time
	}{
		{"spring-gap", "30 2 * * *", time.Date(2026, time.March, 8, 6, 31, 0, 0, time.UTC), []time.Time{time.Date(2026, time.March, 9, 6, 30, 0, 0, time.UTC)}},
		{"autumn-fold", "30 1 * * *", time.Date(2026, time.November, 1, 4, 31, 0, 0, time.UTC), []time.Time{time.Date(2026, time.November, 1, 5, 30, 0, 0, time.UTC), time.Date(2026, time.November, 1, 6, 30, 0, 0, time.UTC)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			schedule, err := scheduler.NewSchedule(test.name, "open-orders", scheduler.Cron(test.expression), scheduler.WithTimezone(location.String()))
			if err != nil {
				t.Fatal(err)
			}
			registry, err := scheduler.Compile(schedule)
			if err != nil {
				t.Fatal(err)
			}
			// A caller-owned Clock supplies the reference instant; no synthetic
			// sleep, wall-clock waiting or timer adapter is introduced.
			clock := manual.NewFixed(test.after)
			after := clock.Now()
			for _, expected := range test.want {
				due, err := registry.Next(test.name, after)
				if err != nil || !due.Equal(expected) {
					t.Fatalf("physical occurrence = %v, %v; want %v", due, err, expected)
				}
				date, err := calendar.DateFromTime(due, location)
				if err != nil {
					t.Fatal(err)
				}
				civil, err := openingcalendar.FromDate(date)
				if err != nil || civil.String() != expected.In(oracle).Format("2006-01-02") {
					t.Fatal("scheduled instant lost the owning local civil date")
				}
				open, err := openingSchedule.IsOpen(due)
				if err != nil || !open.Open {
					t.Fatalf("scheduled occurrence fell outside the composed local window: %v", err)
				}
				after = due
			}
			if len(test.want) == 2 {
				first, second := test.want[0], test.want[1]
				if first.In(oracle).Format("15:04") != second.In(oracle).Format("15:04") {
					t.Fatal("fold fixture does not represent repeated civil time")
				}
				period, err := instant.New(first, second, temporal.ClosedOpen)
				if err != nil {
					t.Fatal(err)
				}
				duration, err := period.Duration()
				if err != nil || duration != time.Hour || !period.Includes(first) || period.Includes(second) {
					t.Fatal("Temporal collapsed repeated civil time or lost half-open physical membership")
				}
			}
		})
	}
	// The spring jump crosses two wall-clock hours in one physical hour.
	before := time.Date(2026, time.March, 8, 6, 30, 0, 0, time.UTC)
	after := time.Date(2026, time.March, 8, 7, 30, 0, 0, time.UTC)
	period, err := instant.Range(before, after)
	if err != nil {
		t.Fatal(err)
	}
	duration, err := period.Duration()
	if err != nil || duration != time.Hour || before.In(oracle).Hour() != 1 || after.In(oracle).Hour() != 3 {
		t.Fatal("spring gap was normalized to a civil-duration substitute")
	}
	// Compare canonical sibling-loader outcomes before/at/after both physical
	// transitions and at the half-open close. Agreement proves these finite
	// outcomes, not byte identity of host TZif or every timezone rule.
	spring := time.Date(2026, time.March, 8, 7, 0, 0, 0, time.UTC)
	fold := time.Date(2026, time.November, 1, 6, 0, 0, 0, time.UTC)
	for _, point := range []time.Time{before, spring.Add(-time.Nanosecond), spring, spring.Add(time.Nanosecond), after,
		time.Date(2026, time.March, 8, 8, 0, 0, 0, time.UTC), fold.Add(-30 * time.Minute), fold.Add(-time.Nanosecond), fold, fold.Add(time.Nanosecond), fold.Add(30 * time.Minute)} {
		wantLocal := point.In(oracle)
		if point.In(location).Format(time.RFC3339Nano) != wantLocal.Format(time.RFC3339Nano) {
			t.Fatalf("runtime civil coordinates disagree with the pinned oracle at %v", point)
		}
		date, err := calendar.DateFromTime(point, location)
		if err != nil || date.String() != wantLocal.Format("2006-01-02") {
			t.Fatal("Calendar civil date disagrees with the pinned oracle")
		}
		open, err := openingSchedule.IsOpen(point)
		wantOpen := wantLocal.Hour() >= 1 && wantLocal.Hour() < 4
		if err != nil || open.Open != wantOpen {
			t.Fatalf("Opening Hours point membership disagrees with the pinned oracle at %v", point)
		}
	}
}
