package db

import (
	"encoding/base64"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/soyersoyer/k20a/db/shardbolt"
)

type empty struct{}

type set map[uint32]empty

func (s set) add(id uint32) {
	s[id] = empty{}
}

func (s set) contains(id uint32) bool {
	_, ok := s[id]
	return ok
}

// CollectionDataInputT is the input filter struct for the clients
type CollectionDataInputT struct {
	From     time.Time
	To       time.Time
	Bucket   string
	Timezone string
	Filter   map[string]string
}

// CollectionDataT is the collection's data struct for the clients
type CollectionDataT struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	OwnerEmail   string        `json:"owner_email"`
	SessionSums  []*bucketSumT `json:"session_sums"`
	PageviewSums []*bucketSumT `json:"pageview_sums"`
}

type bucketSumT struct {
	Bucket int64 `json:"bucket"`
	Count  int   `json:"count"`
}

func getDiff(bucketType string) func(time.Time) time.Time {
	switch bucketType {
	default: /*day*/
		return func(t time.Time) time.Time { return t.AddDate(0, 0, 1) }
	case "hour":
		return func(t time.Time) time.Time { return t.Add(time.Hour) }
	case "week":
		return func(t time.Time) time.Time { return t.AddDate(0, 0, 7) }
	case "month":
		return func(t time.Time) time.Time { return t.AddDate(0, 1, 0) }
	}
}

func getTimeMap(bucketType string) func(time.Time, *time.Location) int64 {
	switch bucketType {
	default: /*day*/
		return func(t time.Time, loc *time.Location) int64 {
			t = t.In(loc)
			return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Unix()
		}
	case "hour":
		return func(t time.Time, loc *time.Location) int64 {
			t = t.In(loc)
			return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, loc).Unix()
		}
	case "week":
		return func(t time.Time, loc *time.Location) int64 {
			t = t.In(loc)
			weekFirstDay := t.Day() - (int(t.Weekday())+6)%7
			return time.Date(t.Year(), t.Month(), weekFirstDay, 0, 0, 0, 0, loc).Unix()
		}
	case "month":
		return func(t time.Time, loc *time.Location) int64 {
			t = t.In(loc)
			return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, loc).Unix()
		}
	}
}

type bucketGen struct {
	loc     *time.Location
	timeMap func(time.Time, *time.Location) int64
	Buckets map[int64]int
}

func createBucketGen(bucketType string, begin time.Time, end time.Time, timezone string) *bucketGen {
	diff := getDiff(bucketType)
	timeMap := getTimeMap(bucketType)
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Println("invalid timezone", err, timezone)
	}
	if err == nil {
		begin = begin.In(loc)
		end = end.In(loc)
	}
	loc = begin.Location()

	buckets := map[int64]int{}

	actual := begin
	for {
		if end.Before(actual) || end.Equal(actual) {
			break
		}
		buckets[actual.Unix()] = 0
		actual = diff(actual)
	}

	return &bucketGen{
		loc,
		timeMap,
		buckets,
	}
}

func (bg *bucketGen) Add(t time.Time) {
	bg.Buckets[bg.timeMap(t, bg.loc)]++
}

func (bg *bucketGen) Close() []*bucketSumT {
	bucketSums := make([]*bucketSumT, 0, len(bg.Buckets))
	for k, v := range bg.Buckets {
		bucketSums = append(bucketSums, &bucketSumT{k, v})
	}
	sort.Slice(bucketSums, func(i, j int) bool { return bucketSums[i].Bucket < bucketSums[j].Bucket })
	return bucketSums
}

func readSessions(sdb *shardbolt.DB, from, to time.Time,
	filter map[string]string,
	sessionFunc func(session *ExtSession),
	pvFunc func(pv *ExtPageview)) {

	possibleSessionStart := from.Add(-time.Hour * 24)
	fromKey := marshalTime(possibleSessionStart)
	toKey := marshalTime(to)

	session := &ExtSession{}
	pageview := &ExtPageview{}
	var err error

	sessionFilter := createSessionFilter(filter)
	pvFilter := createPageviewFilter(filter)

	sdb.Iterate(BSession, fromKey, toKey, func(k []byte, v []byte) {
		session.PageviewCount = 0
		session.Key = EncodeSessionKey(k)
		session.Begin, err = unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if err := protoDecode(v, &session.Session); err != nil {
			log.Println(err, v)
			return
		}
		if !sessionFilter.match(session) {
			return
		}
		matchSession := false
		sdb.IteratePrefix(BPageview, k, func(pvk []byte, pvv []byte) {
			pageview.Time = GetTimeFromPVKey(pvk)
			session.PageviewCount++
			if session.End == session.Begin.UnixNano() {
				session.End = pageview.Time.UnixNano()
			}
			/* TODO - ability to skip the pageview decoding */
			if err := protoDecode(pvv, &pageview.Pageview); err != nil {
				log.Println(err, v)
				return
			}
			if !pvFilter.match(&pageview.Pageview) {
				return
			}
			matchSession = true

			if pageview.Time.After(from) && pageview.Time.Before(to) && pvFunc != nil {
				pvFunc(pageview)
			}
		})
		if pvFilter != nil && !matchSession {
			return
		}
		if !sessionFilter.matchPVC(session) {
			return
		}
		if session.Begin.After(from) {
			sessionFunc(session)
		}
	})
}

// GetBucketSums returns the bucket by hour or day or week or month
func GetBucketSums(collection *Collection, input *CollectionDataInputT) (*CollectionDataT, error) {
	sdb, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}
	output := &CollectionDataT{
		Name:         collection.Name,
		ID:           collection.ID,
		OwnerEmail:   collection.OwnerEmail,
		SessionSums:  nil,
		PageviewSums: nil,
	}

	sbg := createBucketGen(input.Bucket, input.From, input.To, input.Timezone)
	pvbg := createBucketGen(input.Bucket, input.From, input.To, input.Timezone)

	readSessions(sdb, input.From, input.To, input.Filter,
		func(session *ExtSession) {
			sbg.Add(session.Begin)
		},
		func(pv *ExtPageview) {
			pvbg.Add(pv.Time)
		},
	)

	output.SessionSums = sbg.Close()
	output.PageviewSums = pvbg.Close()

	return output, nil
}

// CollectionStatDataT is the collection's statistic data struct for the clients
type CollectionStatDataT struct {
	SessionTotal         totalT   `json:"session_total"`
	PageviewTotal        totalT   `json:"pageview_total"`
	AvgSessionLength     totalT   `json:"avg_session_length"`
	BounceRate           percentT `json:"bounce_rate"`
	PageSums             []sumT   `json:"page_sums"`
	QueryStringSums      []sumT   `json:"query_string_sums"`
	HostnameSums         []sumT   `json:"hostname_sums"`
	DeviceTypeSums       []sumT   `json:"device_type_sums"`
	DeviceOSSums         []sumT   `json:"device_os_sums"`
	BrowserNameSums      []sumT   `json:"browser_name_sums"`
	BrowserVersionSums   []sumT   `json:"browser_version_sums"`
	BrowserLanguageSums  []sumT   `json:"browser_language_sums"`
	PageviewCountSums    []sumT   `json:"pageview_count_sums"`
	ScreenResolutionSums []sumT   `json:"screen_resolution_sums"`
	WindowResolutionSums []sumT   `json:"window_resolution_sums"`
	CountryCodeSums      []sumT   `json:"country_code_sums"`
	CitySums             []sumT   `json:"city_sums"`
	ASNameSums           []sumT   `json:"as_name_sums"`
	ReferrerSums         []sumT   `json:"referrer_sums"`
}

type totalT struct {
	Count       int     `json:"count"`
	DiffPercent float64 `json:"diff_percent"`
}

type percentT struct {
	Percent     float64 `json:"percent"`
	DiffPercent float64 `json:"diff_percent"`
}

type sumT struct {
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

func getTotal(m *map[string]int) int {
	total := 0
	for _, v := range *m {
		total += v
	}
	return total
}

func getSums(m *map[string]int) []sumT {
	output := []sumT{}
	total := getTotal(m)
	for k, v := range *m {
		output = append(output, sumT{Name: k, Count: v, Percent: float64(v) / float64(total)})
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Count > output[j].Count })
	return output
}

func getPercentByKey(m *map[string]int, key string) float64 {
	total := getTotal(m)
	if total == 0 {
		return 1.0
	}
	return float64((*m)[key]) / float64(total)
}

type sessionFilter struct {
	Hostname         *string
	DeviceType       *string
	DeviceOS         *string
	BrowserName      *string
	BrowserVersion   *string
	BrowserLanguage  *string
	PageviewCount    *int32
	ScreenResolution *string
	WindowResolution *string
	CountryCode      *string
	City             *string
	ASName           *string
	Referrer         *string
}

func createSessionFilter(filter map[string]string) *sessionFilter {
	empty := &sessionFilter{}
	sf := &sessionFilter{}
	if v, ok := filter["hostname"]; ok {
		sf.Hostname = &v
	}
	if v, ok := filter["device_type"]; ok {
		sf.DeviceType = &v
	}
	if v, ok := filter["device_os"]; ok {
		sf.DeviceOS = &v
	}
	if v, ok := filter["browser_name"]; ok {
		sf.BrowserName = &v
	}
	if v, ok := filter["browser_version"]; ok {
		sf.BrowserVersion = &v
	}
	if v, ok := filter["browser_language"]; ok {
		sf.BrowserLanguage = &v
	}
	if v, ok := filter["pageview_count"]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Println("bad pageview_count int filter", v)
		} else {
			ic := int32(i)
			sf.PageviewCount = &ic
		}
	}
	if v, ok := filter["screen_resolution"]; ok {
		sf.ScreenResolution = &v
	}
	if v, ok := filter["window_resolution"]; ok {
		sf.WindowResolution = &v
	}
	if v, ok := filter["country_code"]; ok {
		sf.CountryCode = &v
	}
	if v, ok := filter["city"]; ok {
		sf.City = &v
	}
	if v, ok := filter["as_name"]; ok {
		sf.ASName = &v
	}
	if v, ok := filter["referrer"]; ok {
		sf.Referrer = &v
	}
	if *sf == *empty {
		return nil
	}
	return sf
}

func (sf *sessionFilter) match(session *ExtSession) bool {
	if sf == nil {
		return true
	}
	if sf.Hostname != nil && *sf.Hostname != session.Hostname {
		return false
	}
	if sf.DeviceType != nil && *sf.DeviceType != session.DeviceType {
		return false
	}
	if sf.DeviceOS != nil && *sf.DeviceOS != session.DeviceOS {
		return false
	}
	if sf.BrowserName != nil && *sf.BrowserName != session.BrowserName {
		return false
	}
	if sf.BrowserVersion != nil && *sf.BrowserVersion != session.BrowserVersion {
		return false
	}
	if sf.BrowserLanguage != nil && *sf.BrowserLanguage != session.BrowserLanguage {
		return false
	}
	if sf.ScreenResolution != nil && *sf.ScreenResolution != session.ScreenResolution {
		return false
	}
	if sf.WindowResolution != nil && *sf.WindowResolution != session.WindowResolution {
		return false
	}
	if sf.CountryCode != nil && *sf.CountryCode != session.CountryCode {
		return false
	}
	if sf.City != nil && *sf.City != session.City {
		return false
	}
	if sf.ASName != nil && *sf.ASName != session.ASName {
		return false
	}
	if sf.Referrer != nil && *sf.Referrer != session.Referrer {
		return false
	}
	return true
}
func (sf *sessionFilter) matchPVC(session *ExtSession) bool {
	if sf == nil {
		return true
	}
	if sf.PageviewCount != nil && int(*sf.PageviewCount) != session.PageviewCount {
		return false
	}
	return true
}

type pageviewFilter struct {
	Page        *string
	QueryString *string
}

func createPageviewFilter(filter map[string]string) *pageviewFilter {
	empty := &pageviewFilter{}
	pvf := &pageviewFilter{}
	if v, ok := filter["page"]; ok {
		pvf.Page = &v
	}
	if v, ok := filter["query_string"]; ok {
		pvf.QueryString = &v
	}
	if *pvf == *empty {
		return nil
	}
	return pvf
}

func (pvf *pageviewFilter) match(pv *Pageview) bool {
	if pvf == nil {
		return true
	}
	if pvf.Page != nil && *pvf.Page != pv.Path {
		return false
	}
	if pvf.QueryString != nil && *pvf.QueryString != pv.QueryString {
		return false
	}
	return true
}

// GetStatistics returns the statistic data for the collection
func GetStatistics(collection *Collection, input *CollectionDataInputT) (*CollectionStatDataT, error) {
	sdb, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}

	sessionTotal := 0
	pageviewTotal := 0
	sumOfSessionLength := 0

	prevSessionTotal := 0
	prevPageviewTotal := 0
	sumOfPrevSessionLength := 0
	prevPageviewCountSums := make(map[string]int)

	pageSums := make(map[string]int)
	queryStringSums := make(map[string]int)

	hostnameSums := make(map[string]int)
	deviceTypeSums := make(map[string]int)
	deviceOSSums := make(map[string]int)
	browserNameSums := make(map[string]int)
	browserVersionSums := make(map[string]int)
	browserLanguageSums := make(map[string]int)
	pageviewCountSums := make(map[string]int)
	screenResolutionSums := make(map[string]int)
	windowResolutionSums := make(map[string]int)
	countryCodeSums := make(map[string]int)
	citySums := make(map[string]int)
	asNameSums := make(map[string]int)
	referrerSums := make(map[string]int)

	prevTime := input.From.Add(input.From.Sub(input.To))

	readSessions(sdb, prevTime, input.From, input.Filter,
		func(session *ExtSession) {
			prevSessionTotal++
			sumOfPrevSessionLength += int(session.End/1000000000 - session.Begin.UnixNano()/1000000000)

			prevPageviewCountSums[strconv.Itoa(session.PageviewCount)]++
		},
		func(pv *ExtPageview) {
			prevPageviewTotal++
		})

	readSessions(sdb, input.From, input.To, input.Filter,
		func(session *ExtSession) {
			sessionTotal++
			sumOfSessionLength += int(session.End/1000000000 - session.Begin.UnixNano()/1000000000)

			hostnameSums[session.Hostname]++
			deviceTypeSums[session.DeviceType]++
			deviceOSSums[session.DeviceOS]++
			browserNameSums[session.BrowserName]++
			browserVersionSums[session.BrowserVersion]++
			browserLanguageSums[session.BrowserLanguage]++
			pageviewCountSums[strconv.Itoa(session.PageviewCount)]++
			screenResolutionSums[session.ScreenResolution]++
			windowResolutionSums[session.WindowResolution]++
			countryCodeSums[session.CountryCode]++
			citySums[session.City]++
			asNameSums[session.ASName]++
			referrerSums[session.Referrer]++
		},
		func(pv *ExtPageview) {
			pageviewTotal++

			pageSums[pv.Path]++
			queryStringSums[pv.QueryString]++
		})

	avgSessionLength := safeDiv(sumOfSessionLength, sessionTotal)
	prevAvgSessionLength := safeDiv(sumOfPrevSessionLength, prevSessionTotal)

	bounceRate := getPercentByKey(&pageviewCountSums, "1")
	prevBounceRate := getPercentByKey(&prevPageviewCountSums, "1")

	return &CollectionStatDataT{
		SessionTotal:         totalT{sessionTotal, getGrowthPercent(sessionTotal, prevSessionTotal)},
		PageviewTotal:        totalT{pageviewTotal, getGrowthPercent(pageviewTotal, prevPageviewTotal)},
		AvgSessionLength:     totalT{avgSessionLength, getGrowthPercent(avgSessionLength, prevAvgSessionLength)},
		BounceRate:           percentT{bounceRate, getGrowthPercentF(bounceRate, prevBounceRate)},
		PageSums:             getSums(&pageSums),
		QueryStringSums:      getSums(&queryStringSums),
		HostnameSums:         getSums(&hostnameSums),
		DeviceTypeSums:       getSums(&deviceTypeSums),
		DeviceOSSums:         getSums(&deviceOSSums),
		BrowserNameSums:      getSums(&browserNameSums),
		BrowserVersionSums:   getSums(&browserVersionSums),
		BrowserLanguageSums:  getSums(&browserLanguageSums),
		PageviewCountSums:    getSums(&pageviewCountSums),
		ScreenResolutionSums: getSums(&screenResolutionSums),
		WindowResolutionSums: getSums(&windowResolutionSums),
		CountryCodeSums:      getSums(&countryCodeSums),
		CitySums:             getSums(&citySums),
		ASNameSums:           getSums(&asNameSums),
		ReferrerSums:         getSums(&referrerSums),
	}, nil
}

func safeDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

func getGrowthPercent(actual, prev int) float64 {
	if prev == 0 {
		return 1.0
	}
	return float64(actual)/float64(prev) - 1.0
}

func getGrowthPercentF(actual, prev float64) float64 {
	if prev == 0.0 {
		return 1.0
	}
	return actual/prev - 1.0
}

// SessionDataT is the session data struct for the clients
type SessionDataT struct {
	Key              string `json:"key"`
	Hostname         string `json:"hostname"`
	DeviceOS         string `json:"device_os"`
	BrowserName      string `json:"browser_name"`
	BrowserVersion   string `json:"browser_version"`
	BrowserLanguage  string `json:"browser_language"`
	ScreenResolution string `json:"screen_resolution"`
	WindowResolution string `json:"window_resolution"`
	DeviceType       string `json:"device_type"`
	CountryCode      string `json:"country_code"`
	City             string `json:"city"`
	ASNumber         int32  `json:"as_number"`
	ASName           string `json:"as_name"`
	UserAgent        string `json:"user_agent"`
	UserIP           string `json:"user_ip"`
	UserHostname     string `json:"user_hostname"`
	Begin            int64  `json:"begin"`
	End              int64  `json:"end"`
	PageviewCount    int    `json:"pageview_count"`
	Referrer         string `json:"referrer"`
}

// EncodeSessionKey encodes a session key with base64
func EncodeSessionKey(k []byte) string {
	return base64.StdEncoding.EncodeToString(k)
}

// DecodeSessionKey decodes a session key from base64
func DecodeSessionKey(key string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(key)
}

// GetSessions returns the collection's sessions
func GetSessions(collection *Collection, input *CollectionDataInputT) ([]*SessionDataT, error) {
	sdb, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}

	ret := []*SessionDataT{}

	readSessions(sdb, input.From, input.To, input.Filter,
		func(session *ExtSession) {
			ret = append(ret, &SessionDataT{
				Key:              session.Key,
				Hostname:         session.Hostname,
				DeviceOS:         session.DeviceOS,
				BrowserName:      session.BrowserName,
				BrowserVersion:   session.BrowserVersion,
				BrowserLanguage:  session.BrowserLanguage,
				ScreenResolution: session.ScreenResolution,
				WindowResolution: session.WindowResolution,
				DeviceType:       session.DeviceType,
				CountryCode:      session.CountryCode,
				City:             session.City,
				ASNumber:         session.ASNumber,
				ASName:           session.ASName,
				UserAgent:        session.UserAgent,
				UserIP:           session.UserIP,
				UserHostname:     session.UserHostname,
				Begin:            session.Begin.UnixNano(),
				End:              session.End,
				PageviewCount:    session.PageviewCount,
				Referrer:         session.Referrer,
			})
		}, nil)

	return ret, nil
}

// PageviewDataT is the pageview data struct for the clients
type PageviewDataT struct {
	Time        int64  `json:"time"`
	Path        string `json:"path"`
	QueryString string `json:"query_string"`
}

// GetPageviews returns the session's pageviews
func GetPageviews(collection *Collection, sessionKey []byte) ([]*PageviewDataT, error) {
	sdb, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}

	pageviews := []*PageviewDataT{}

	pageView := &Pageview{}
	sdb.IteratePrefix(BPageview, sessionKey, func(pvk []byte, pvv []byte) {
		if err := protoDecode(pvv, pageView); err != nil {
			log.Println(err, pvv)
			return
		}

		pvtime := GetTimeFromPVKey(pvk)
		pageviews = append(pageviews, &PageviewDataT{
			Time:        pvtime.UnixNano(),
			Path:        pageView.Path,
			QueryString: pageView.QueryString,
		})
	})
	return pageviews, nil
}
