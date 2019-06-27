package sched

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	scheduleSeparator    = ";"
	nonYamlTypeSeparator = "="
	DailyType            = "daily"
	MonthlyType          = "monthly"
	WeeklyType           = "weekly"
	PeriodicType         = "periodic"
	retainSeparator      = ","
	MonthlyRetain        = 12
	WeeklyRetain         = 5
	DailyRetain          = 7
)

var (
	// speedUp advances the clock faster for tests.
	speedUp = false
)

// SpeedUp advances teh clock faster for tests.
func SpeedUp() {
	speedUp = true
}

func inSpeedUp() bool {
	return speedUp
}

type IntervalSpec struct {
	Freq    string
	Period  uint64 `yaml:"period,omitempty"`
	Month   int    `yaml:"month,omitempty"`
	Weekday int    `yaml:"weekday,omitempty"`
	Day     int    `yaml:"day,omitempty"`
	Hour    int    `yaml:"hour,omitempty"`
	Minute  int    `yaml:"minute,omitempty"`
}

type Interval interface {
	nextAfter(t time.Time) time.Time
	String() string
	IntervalType() string
	Spec() IntervalSpec
}

type periodic struct {
	delta time.Duration
}

func (p periodic) nextAfter(t time.Time) time.Time {
	return t.Add(p.delta)
}

func (p periodic) String() string {
	return fmt.Sprintf("%v %v", PeriodicType, p.delta)
}

func (p periodic) IntervalType() string {
	return PeriodicType
}

func (p periodic) Spec() IntervalSpec {
	return IntervalSpec{Freq: PeriodicType, Period: uint64(p.delta)}
}

func Periodic(period time.Duration) Interval {
	return periodic{period}
}

type daily struct {
	hour   int
	minute int
}

func (d daily) nextAfter(t time.Time) time.Time {
	h, m := t.Hour(), t.Minute()
	if h < d.hour {
		t = t.Add(time.Duration(d.hour-h) * time.Hour)
	} else if h > d.hour || m >= d.minute {
		t = t.Add(time.Duration(24-h+d.hour) * time.Hour)
	}
	return t.Add(time.Duration(d.minute-m) * time.Minute)
}

func (d daily) String() string {
	return fmt.Sprintf("%s @%02d:%02d", DailyType, d.hour, d.minute)
}

func (p daily) IntervalType() string {
	return DailyType
}

func (d daily) Spec() IntervalSpec {
	return IntervalSpec{Freq: DailyType, Hour: d.hour, Minute: d.minute}
}

func Daily(hour int, minute int) Interval {
	return daily{hour, minute}
}

type weekly struct {
	day time.Weekday
	tod daily
}

func (w weekly) nextAfter(t time.Time) time.Time {
	t = w.tod.nextAfter(t)
	d := t.Weekday()
	delta := time.Duration(w.day - d)
	if w.day < d {
		delta += 7
	}
	return t.Add(delta * 24 * time.Hour)
}

func (w weekly) String() string {
	return fmt.Sprintf("%s %s@%02d:%02d", WeeklyType, w.day.String(), w.tod.hour,
		w.tod.minute)
}

func (p weekly) IntervalType() string {
	return WeeklyType
}

func (w weekly) Spec() IntervalSpec {
	return IntervalSpec{Freq: WeeklyType,
		Weekday: int(w.day), Hour: w.tod.hour, Minute: w.tod.minute}
}

func Weekly(dow time.Weekday, hour int, minute int) Interval {
	return weekly{dow, daily{hour, minute}}
}

type monthly struct {
	day int
	tod daily
}

func (m monthly) nextAfter(t time.Time) time.Time {
	t = m.tod.nextAfter(t)
	y, month, d := t.Date()
	if d > m.day {
		daysIn := time.Date(y, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
		t = t.Add(time.Duration(daysIn*24) * time.Hour)
	}
	return t
}

func (m monthly) String() string {
	return fmt.Sprintf("%s %d@%02d:%02d", MonthlyType, m.day, m.tod.hour,
		m.tod.minute)
}

func (p monthly) IntervalType() string {
	return MonthlyType
}

func (m monthly) Spec() IntervalSpec {
	return IntervalSpec{Freq: MonthlyType,
		Day: m.day, Hour: m.tod.hour, Minute: m.tod.minute}
}

func Monthly(day int, hour int, minute int) Interval {
	return monthly{day, daily{hour, minute}}
}

func parseSpec(spec *IntervalSpec) (Interval, error) {
	switch spec.Freq {
	case PeriodicType:
		return Periodic(time.Duration(spec.Period)), nil
	case DailyType:
		return Daily(spec.Hour, spec.Minute), nil
	case WeeklyType:
		dow := time.Weekday(spec.Weekday)
		return Weekly(dow, spec.Hour, spec.Minute), nil
	case MonthlyType:
		if spec.Day == 0 {
			spec.Day = 1
		}
		return Monthly(spec.Day, spec.Hour, spec.Minute), nil
	}
	return nil, fmt.Errorf("Invalid schedule spec")
}

func parseRetainSpec(spec *RetainIntervalSpec) (RetainInterval, error) {
	s, err := parseSpec(&spec.IntervalSpec)
	if err != nil {
		return nil, err
	}
	return &RetainIntervalImpl{iv: s, retain: spec.Retain}, nil
}

func parseNonYamlSchedule(schedule string) (RetainIntervalSpec, error) {
	parts := strings.Split(schedule, nonYamlTypeSeparator)
	if len(parts) != 2 || !IsIntervalType(parts[0]) {
		return RetainIntervalSpec{},
			fmt.Errorf("Invalid schedule specification: %s", schedule)
	}
	if fn, ok := ParseCLI[parts[0]]; ok {
		return fn(parts[1])
	}
	return ParsePeriodic(parts[1])
}

func ParseSchedule(schedule string) ([]RetainInterval, error) {
	var schedInts []RetainInterval
	if schedule == "" {
		return schedInts, nil
	}
	var sspec []RetainIntervalSpec
	err := yaml.Unmarshal([]byte(schedule), &sspec)
	if err != nil {
		intv, err := parseNonYamlSchedule(schedule)
		if err != nil {
			return nil, err
		}
		sspec = []RetainIntervalSpec{intv}
	}
	for _, s := range sspec {
		interval, err := parseRetainSpec(&s)
		if err != nil {
			return nil, err
		}
		schedInts = append(schedInts, interval)
	}
	return schedInts, nil
}

func ParseScheduleAndPolicies(scheduleString string) (
	[]RetainInterval,
	*PolicyTags,
	error,
) {
	schedules := strings.Split(scheduleString, scheduleSeparator)
	policies := &PolicyTags{}
	var schedInts []RetainInterval
	for _, schedule := range schedules {
		if strings.HasPrefix(schedule, policyTag) {
			if policy, err := ParsePolicyTags(schedule); err == nil {
				policies.Names = append(policies.Names, policy.Names...)
			} else {
				return nil, nil, err
			}
		} else {
			if scheds, err := ParseSchedule(schedule); err == nil {
				if len(scheds) > 0 {
					schedInts = append(schedInts, scheds...)
				}
			} else {
				return nil, nil, err
			}
		}
	}
	if len(policies.Names) == 0 {
		return schedInts, nil, nil
	}
	return schedInts, policies, nil
}

func ScheduleStringRetainInv(
	intvs []RetainInterval,
	p *PolicyTags,
) (string, error) {
	intvSpecs := make([]RetainIntervalSpec, 0)
	for _, intv := range intvs {
		intvSpecs = append(intvSpecs, intv.RetainIntervalSpec())
	}
	return ScheduleString(intvSpecs, p)
}

func ScheduleString(items []RetainIntervalSpec, p *PolicyTags) (string, error) {
	schedString := ""
	if len(items) != 0 {
		s, err := yaml.Marshal(items)
		if err != nil {
			return "", err
		}
		schedString = string(s)
	}
	if p != nil {
		pString := p.String()
		if pString != "" {
			if schedString != "" {
				schedString = schedString + scheduleSeparator
			}
			schedString = schedString + pString
		}
	}
	return schedString, nil
}

func ScheduleIntervalSummary(items []Interval, policyTags *PolicyTags) string {
	if len(items) == 0 {
		return ""
	}
	summary := ""
	for i, iv := range items {
		if i != 0 {
			summary += ", "
		}
		summary += iv.String()
	}
	return summary
}

func ScheduleSummary(items []RetainInterval, policyTags *PolicyTags) string {
	summary := ""
	if policyTags != nil {
		summary = policyTags.String()
	}
	if len(items) == 0 {
		return summary
	}
	if len(summary) > 0 {
		summary = summary + scheduleSeparator
	}
	for i, iv := range items {
		if i != 0 {
			summary += ", "
		}
		summary += iv.String()
	}
	return summary
}

func timeOfDay(hhmm string) (int, int, error) {
	if hhmm == "" {
		return 0, 0, nil
	}
	t := strings.Split(hhmm, ":")
	if len(t) == 1 {
		t = append(t, "0")
	}
	ok := len(t) == 2
	var h, m int
	if ok {
		var e1, e2 error
		h, e1 = strconv.Atoi(t[0])
		m, e2 = strconv.Atoi(t[1])
		ok = e1 == nil && e2 == nil && 0 <= h && h < 24 && 0 <= m && m < 60
	}
	if !ok {
		return 0, 0, fmt.Errorf("invalid start time %v", hhmm)
	}
	return h, m, nil
}

func dayOfWeek(wd string) (time.Weekday, error) {
	day := strings.Title(wd)
	if day == "" {
		day = "Sunday"
	}
	d := time.Sunday
	for d <= time.Saturday && d.String() != day {
		d++
	}
	if d > time.Saturday {
		return time.Sunday, fmt.Errorf("invalid weekday %v", day)
	}
	return d, nil
}

func parseRetainNumber(input string) (RetainIntervalSpec, string, error) {
	r := RetainIntervalSpec{}
	parts := strings.Split(input, retainSeparator)
	if len(parts) > 1 {
		if retain, err := strconv.Atoi(parts[1]); err != nil {
			return r, "", fmt.Errorf("Invalid number: %s", parts[1])
		} else if retain <= 0 {
			return r, "", fmt.Errorf("Keep number should be greater than 0")
		} else {
			r.Retain = uint32(retain)
		}
	}
	return r, parts[0], nil
}

func ParsePeriodic(input string) (RetainIntervalSpec, error) {
	r, intvl, err := parseRetainNumber(input)
	if err != nil {
		return r, err
	}
	if intvl == "" {
		return r, fmt.Errorf("Interval is missing")
	}
	delta, err := strconv.ParseUint(intvl, 10, 64)
	if err != nil {
		return r, fmt.Errorf("Invalid interval %s", intvl)
	}
	dt := time.Duration(delta) * time.Minute
	r.IntervalSpec = Periodic(dt).Spec()
	return r, nil
}

// parseDaily item [@]hh:mm,r
func parseDaily(dailyStr string) (RetainIntervalSpec, error) {
	r, daily, err := parseRetainNumber(dailyStr)
	if err != nil {
		return r, err
	}
	if daily == "" {
		return r, fmt.Errorf("Daily schedule is missing")
	}
	dt := strings.Split(daily, "@")
	h, m, err := timeOfDay(dt[len(dt)-1])
	if err != nil {
		return RetainIntervalSpec{}, err
	}
	r.IntervalSpec = Daily(h, m).Spec()
	return r, nil
}

// parseWeekly item weekday@hh:mm,r
func parseWeekly(weeklyStr string) (RetainIntervalSpec, error) {
	r, weekly, err := parseRetainNumber(weeklyStr)
	if err != nil {
		return r, err
	}
	if weekly == "" {
		return r, fmt.Errorf("Weekly schedule is missing")
	}
	dt := strings.Split(weekly, "@")
	if len(dt) == 1 {
		dt = append(dt, "0:0")
	}
	if len(dt) != 2 {
		return RetainIntervalSpec{},
			fmt.Errorf("Invalid weekly spec %v", weeklyStr)
	}
	d, err := dayOfWeek(dt[0])
	if err != nil {
		return RetainIntervalSpec{}, err
	}
	h, m, err := timeOfDay(dt[1])
	r.IntervalSpec = Weekly(d, h, m).Spec()
	return r, err
}

// parseMonthly item day@hh:mm,r
func parseMonthly(monthlyStr string) (RetainIntervalSpec, error) {
	r, monthly, err := parseRetainNumber(monthlyStr)
	if err != nil {
		return r, err
	}
	if monthly == "" {
		return r, fmt.Errorf("Monthly schedule is missing")
	}
	dt := strings.Split(monthly, "@")
	if len(dt) == 1 {
		dt = append(dt, "0:0")
	}
	if len(dt) != 2 {
		return RetainIntervalSpec{},
			fmt.Errorf("Invalid monthly spec %v", monthlyStr)
	}
	d, err := strconv.Atoi(dt[0])
	if err != nil || d < 0 || d > 31 {
		return RetainIntervalSpec{},
			fmt.Errorf("Invalid day of month %v", dt[0])
	}
	h, m, err := timeOfDay(dt[1])
	if err != nil {
		return RetainIntervalSpec{}, err
	}
	r.IntervalSpec = Monthly(d, h, m).Spec()
	return r, nil
}

var ParseCLI = map[string]func(string) (RetainIntervalSpec, error){
	DailyType:   parseDaily,
	WeeklyType:  parseWeekly,
	MonthlyType: parseMonthly,
}

func IntervalType(interval Interval) string {
	return strings.Split(interval.String(), " ")[0]
}

func IsIntervalType(t string) bool {
	knownTypes := []string{PeriodicType, DailyType, WeeklyType, MonthlyType}
	for _, p := range knownTypes {
		if p == t {
			return true
		}
	}
	return false
}

// RetainIntervalSpec is the serialized form of retain interval
type RetainIntervalSpec struct {
	IntervalSpec `yaml:",inline"`
	Retain       uint32 `yaml:"retain,omitempty"`
}

// RetainInterval is a schedule interval with a number of instances to retain.
type RetainInterval interface {
	Interval
	// RetainNumber is the number of instances to retain
	RetainNumber() uint32
	// RetainIntervalSpec returns RetainIntervalSpec
	RetainIntervalSpec() RetainIntervalSpec
}

// RetainIntervalImpl implements the RetainInterval interface
type RetainIntervalImpl struct {
	iv     Interval
	retain uint32
}

func NewRetainInterval(iv Interval) RetainInterval {
	return RetainIntervalImpl{iv: iv}
}

func (p RetainIntervalImpl) nextAfter(t time.Time) time.Time {
	newTime := p.iv.nextAfter(t)
	if inSpeedUp() {
		return t.Add(time.Minute)
	}
	return newTime
}

func (p RetainIntervalImpl) String() string {
	str := p.iv.String()
	if p.RetainNumber() > 0 {
		return fmt.Sprintf("%s,keep last %d", str, p.RetainNumber())
	}
	return str
}

func (p RetainIntervalImpl) IntervalType() string {
	return p.iv.IntervalType()
}

func (p RetainIntervalImpl) Spec() IntervalSpec {
	return p.iv.Spec()
}

func (p RetainIntervalImpl) RetainNumber() uint32 {
	return p.retain
}

func SetupIntvWithDefaults(intvs []RetainInterval) []RetainInterval {
	retIntvs := make([]RetainInterval, 0)
	for _, intv := range intvs {
		p := &RetainIntervalImpl{iv: intv}
		if intv.RetainNumber() == 0 {
			switch intv.IntervalType() {
			case DailyType:
				p.retain = DailyRetain
			case WeeklyType, PeriodicType:
				p.retain = WeeklyRetain
			case MonthlyType:
				p.retain = MonthlyRetain
			}
		} else {
			p.retain = intv.RetainNumber()
		}
		retIntvs = append(retIntvs, p)
	}
	return retIntvs
}

func (p RetainIntervalImpl) RetainIntervalSpec() RetainIntervalSpec {
	return RetainIntervalSpec{IntervalSpec: p.Spec(), Retain: p.RetainNumber()}
}

const (
	policyTag           = "policy"
	policyStrSeparator  = "="
	policyNameSeparator = ","
)

// PolicyTags groups one or more policies together.
type PolicyTags struct {
	// Names is the list of policy names
	Names []string
}

var policyNameRegex = regexp.MustCompile("[A-Za-z0-9_]+")

func NewPolicyTags(policies string) (*PolicyTags, error) {
	if policies == "" {
		return nil, nil
	}
	p := &PolicyTags{Names: strings.Split(policies, policyNameSeparator)}
	for _, name := range p.Names {
		if !policyNameRegex.MatchString(name) {
			return nil, fmt.Errorf("Invalid policy name '%s'", name)
		}
	}
	return p, nil
}

func (p *PolicyTags) Summary() string {
	if len(p.Names) == 0 {
		return ""
	}
	return fmt.Sprintf("%s=%s", policyTag,
		strings.Join(p.Names, policyNameSeparator))
}

func (p *PolicyTags) String() string {
	return p.Summary()
}

func ParsePolicyTags(policyTagsStr string) (*PolicyTags, error) {
	if policyTagsStr == "" {
		return nil, nil
	}
	names := strings.Split(policyTagsStr, policyStrSeparator)
	if len(names) != 2 || names[0] != policyTag {
		return nil, fmt.Errorf("Invalid policy string %s", policyTagsStr)
	}
	return NewPolicyTags(names[1])
}

func SamePolicyTags(p1, p2 *PolicyTags) bool {
	if p1 == p2 {
		return true
	}
	if p1 == nil && p2 != nil || p2 == nil && p1 != nil ||
		len(p1.Names) != len(p2.Names) {
		return false
	}
next:
	for _, name := range p1.Names {
		for _, othername := range p2.Names {
			if name == othername {
				continue next
			}
		}
		return false
	}
	return true
}
