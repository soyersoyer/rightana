package db

import (
	"bytes"
	"encoding/base64"
	"log"
	"sort"
	"strconv"
	"time"
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

type CollectionDataInputT struct {
	From     time.Time
	To       time.Time
	Bucket   string
	Timezone string
	Filter   map[string]string
}

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

type BucketGen struct {
	begin   time.Time
	end     time.Time
	next    time.Time
	diff    func(time.Time) time.Time
	Buckets []*bucketSumT
}

func CreateBucketGen(bucketType string, begin time.Time, end time.Time, timezone string) *BucketGen {
	diff := getDiff(bucketType)
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		log.Println("invalid timezone", err, timezone)
	}
	if err == nil {
		begin = begin.In(loc)
		end = end.In(loc)
	}

	return &BucketGen{
		begin,
		end,
		diff(begin),
		diff,
		[]*bucketSumT{{begin.Unix(), 0}},
	}
}

func (bg *BucketGen) Add(t time.Time) {
	for {
		if t.Before(bg.next) {
			bg.Buckets[len(bg.Buckets)-1].Count++
			break
		}
		bg.addNewBucket()
	}
}

func (bg *BucketGen) Close() []*bucketSumT {
	for {
		if bg.end.Before(bg.next) || bg.end.Equal(bg.next) {
			break
		}
		bg.addNewBucket()
	}
	return bg.Buckets
}

func (bg *BucketGen) addNewBucket() {
	bg.Buckets = append(bg.Buckets, &bucketSumT{bg.next.Unix(), 0})
	bg.next = bg.diff(bg.next)
}

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

	fromKey := marshalTime(input.From)
	toKey := marshalTime(input.To)

	session := &Session{}
	pageview := &Pageview{}

	validSessions := set{}
	validPVSessions := set{}

	sessionFilter := createSessionFilter(input.Filter)
	pvFilter := createPageviewFilter(input.Filter)

	if pvFilter != nil {
		sdb.Iterate(BPageview, fromKey, toKey, func(k []byte, v []byte) {
			if err := protoDecode(v, pageview); err != nil {
				log.Println(err, v)
				return
			}
			if !pvFilter.match(pageview) {
				return
			}
			validPVSessions.add(GetIDFromKey(k))
		})
	}

	bg := CreateBucketGen(input.Bucket, input.From, input.To, input.Timezone)
	sdb.Iterate(BSession, fromKey, toKey, func(k []byte, v []byte) {
		t, err := unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if pvFilter != nil && !validPVSessions.contains(GetIDFromKey(k)) {
			return
		}
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		if sessionFilter != nil {
			if !sessionFilter.match(session) {
				return
			}
			validSessions.add(GetIDFromKey(k))
		}
		bg.Add(t)
	})
	output.SessionSums = bg.Close()

	bg = CreateBucketGen(input.Bucket, input.From, input.To, input.Timezone)
	sdb.Iterate(BPageview, fromKey, toKey, func(k []byte, v []byte) {
		t, err := unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if sessionFilter != nil && !validSessions.contains(GetIDFromKey(k)) {
			return
		}
		if pvFilter != nil {
			if err := protoDecode(v, pageview); err != nil {
				log.Println(err, v)
				return
			}
			if !pvFilter.match(pageview) {
				return
			}
		}

		bg.Add(t)
	})
	output.PageviewSums = bg.Close()

	return output, nil
}

type CollectionStatDataT struct {
	SessionTotal         totalT   `json:"session_total"`
	PageviewTotal        totalT   `json:"pageview_total"`
	AvgSessionLength     totalT   `json:"avg_session_length"`
	BounceRate           percentT `json:"bounce_rate"`
	PageSums             []sumT   `json:"page_sums"`
	QueryStringSums      []sumT   `json:"query_string_sums"`
	ReferrerSums         []sumT   `json:"referrer_sums"`
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
}

type totalT struct {
	Count       int     `json:"count"`
	DiffPercent float32 `json:"diff_percent"`
}

type percentT struct {
	Percent     float32 `json:"percent"`
	DiffPercent float32 `json:"diff_percent"`
}

type sumT struct {
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	Percent float32 `json:"percent"`
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
		output = append(output, sumT{Name: k, Count: v, Percent: float32(v) / float32(total)})
	}
	sort.Slice(output, func(i, j int) bool { return output[i].Count > output[j].Count })
	return output
}

func getPercentByKey(m *map[string]int, key string) float32 {
	total := getTotal(m)
	if total == 0 || (*m)[key] == 0 {
		return 1.0
	}
	return float32((*m)[key]) / float32(total)
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
	if *sf == *empty {
		return nil
	}
	return sf
}

func (sf *sessionFilter) match(session *Session) bool {
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
	if sf.PageviewCount != nil && *sf.PageviewCount != session.PageviewCount {
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
	return true
}

type pageviewFilter struct {
	Page        *string
	QueryString *string
	Referrer    *string
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
	if v, ok := filter["referrer"]; ok {
		pvf.Referrer = &v
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
	if pvf.Referrer != nil && *pvf.Referrer != pv.ReferrerURL {
		return false
	}
	return true
}

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
	referrerSums := make(map[string]int)

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

	prevTime := input.From.Add(input.From.Sub(input.To))

	prevKey := marshalTime(prevTime)
	fromKey := marshalTime(input.From)
	toKey := marshalTime(input.To)

	pvFilter := createPageviewFilter(input.Filter)
	sessionFilter := createSessionFilter(input.Filter)

	prevValidPVSessions := set{}
	prevValidSessions := set{}

	validPVSessions := set{}
	validSessions := set{}

	session := &Session{}
	pageview := &Pageview{}

	if pvFilter != nil {
		sdb.Iterate(BPageview, prevKey, fromKey, func(k []byte, v []byte) {
			if err := protoDecode(v, pageview); err != nil {
				log.Println(err, v)
				return
			}
			if !pvFilter.match(pageview) {
				return
			}
			prevValidPVSessions.add(GetIDFromKey(k))
		})
	}

	sdb.Iterate(BSession, prevKey, fromKey, func(k []byte, v []byte) {
		t, err := unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if pvFilter != nil && !prevValidPVSessions.contains(GetIDFromKey(k)) {
			return
		}
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		if sessionFilter != nil {
			if !sessionFilter.match(session) {
				return
			}
			prevValidSessions.add(GetIDFromKey(k))
		}
		prevSessionTotal++
		sumOfPrevSessionLength += int(session.End/1000000000 - t.UnixNano()/1000000000)

		prevPageviewCountSums[strconv.Itoa(int(session.PageviewCount))]++
	})

	sdb.Iterate(BPageview, prevKey, fromKey, func(k []byte, v []byte) {
		sid := GetIDFromKey(k)
		if sessionFilter != nil && !prevValidSessions.contains(sid) {
			return
		}
		if pvFilter != nil && !prevValidPVSessions.contains(sid) {
			return
		}
		if err := protoDecode(v, pageview); err != nil {
			log.Println(err, v)
			return
		}
		if pvFilter != nil && !pvFilter.match(pageview) {
			return
		}
		prevPageviewTotal++
	})

	if pvFilter != nil {
		sdb.Iterate(BPageview, fromKey, toKey, func(k []byte, v []byte) {
			if err := protoDecode(v, pageview); err != nil {
				log.Println(err, v)
				return
			}
			if !pvFilter.match(pageview) {
				return
			}
			validPVSessions.add(GetIDFromKey(k))
		})
	}

	sdb.Iterate(BSession, fromKey, toKey, func(k []byte, v []byte) {
		t, err := unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if pvFilter != nil && !validPVSessions.contains(GetIDFromKey(k)) {
			return
		}
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		if sessionFilter != nil {
			if !sessionFilter.match(session) {
				return
			}
			validSessions.add(GetIDFromKey(k))
		}

		sessionTotal++
		sumOfSessionLength += int(session.End/1000000000 - t.UnixNano()/1000000000)

		hostnameSums[session.Hostname]++
		deviceTypeSums[session.DeviceType]++
		deviceOSSums[session.DeviceOS]++
		browserNameSums[session.BrowserName]++
		browserVersionSums[session.BrowserVersion]++
		browserLanguageSums[session.BrowserLanguage]++
		pageviewCountSums[strconv.Itoa(int(session.PageviewCount))]++
		screenResolutionSums[session.ScreenResolution]++
		windowResolutionSums[session.WindowResolution]++
		countryCodeSums[session.CountryCode]++
		citySums[session.City]++
		asNameSums[session.ASName]++
	})

	sdb.Iterate(BPageview, fromKey, toKey, func(k []byte, v []byte) {
		sid := GetIDFromKey(k)
		if sessionFilter != nil && !validSessions.contains(sid) {
			return
		}
		if pvFilter != nil && !validPVSessions.contains(sid) {
			return
		}
		if err := protoDecode(v, pageview); err != nil {
			log.Println(err, v)
			return
		}
		if pvFilter != nil && !pvFilter.match(pageview) {
			return
		}
		pageviewTotal++

		pageSums[pageview.Path]++
		queryStringSums[pageview.QueryString]++
		referrerSums[pageview.ReferrerURL]++
	})

	avgSessionLength := safeDiv(sumOfSessionLength, sessionTotal)
	prevAvgSessionLength := safeDiv(sumOfPrevSessionLength, prevSessionTotal)

	bounceRate := getPercentByKey(&pageviewCountSums, "1")
	prevBounceRate := getPercentByKey(&prevPageviewCountSums, "1")

	return &CollectionStatDataT{
		SessionTotal:         totalT{sessionTotal, getGrowthPercent(sessionTotal, prevSessionTotal)},
		PageviewTotal:        totalT{pageviewTotal, getGrowthPercent(pageviewTotal, prevPageviewTotal)},
		AvgSessionLength:     totalT{avgSessionLength, getGrowthPercent(avgSessionLength, prevAvgSessionLength)},
		BounceRate:           percentT{bounceRate, bounceRate/prevBounceRate - 1.0},
		PageSums:             getSums(&pageSums),
		QueryStringSums:      getSums(&queryStringSums),
		ReferrerSums:         getSums(&referrerSums),
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
	}, nil
}

func safeDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	return a / b
}

func getGrowthPercent(actual, prev int) float32 {
	if prev == 0 {
		return 1.0
	}
	return float32(actual)/float32(prev) - 1.0
}

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
	Start            int64  `json:"start"`
	End              int64  `json:"end"`
	PageviewCount    int32  `json:"pageview_count"`
}

func EncodeSessionKey(k []byte) string {
	return base64.StdEncoding.EncodeToString(k)
}

func DecodeSessionKey(key string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(key)
}

func GetSessions(collection *Collection, input *CollectionDataInputT) ([]*SessionDataT, error) {
	sdb, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}

	ret := []*SessionDataT{}

	fromKey := marshalTime(input.From)
	toKey := marshalTime(input.To)

	sessionFilter := createSessionFilter(input.Filter)
	pvFilter := createPageviewFilter(input.Filter)

	session := &Session{}
	pageview := &Pageview{}

	validPVSessions := set{}

	if pvFilter != nil {
		sdb.Iterate(BPageview, fromKey, toKey, func(k []byte, v []byte) {
			if err := protoDecode(v, pageview); err != nil {
				log.Println(err, v)
				return
			}
			if !pvFilter.match(pageview) {
				return
			}
			validPVSessions.add(GetIDFromKey(k))
		})
	}

	sdb.Iterate(BSession, fromKey, toKey, func(k []byte, v []byte) {
		if pvFilter != nil && !validPVSessions.contains(GetIDFromKey(k)) {
			return
		}
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		if !sessionFilter.match(session) {
			return
		}
		ret = append(ret, &SessionDataT{
			Key:              EncodeSessionKey(k),
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
			Start:            GetTimeFromKey(k).UnixNano(),
			End:              session.End,
			PageviewCount:    session.PageviewCount,
		})
	})
	return ret, nil
}

type PageviewDataT struct {
	Time        int64  `json:"time"`
	Path        string `json:"path"`
	ReferrerURL string `json:"referrer_url"`
}

func GetPageviews(collection *Collection, sessionKey []byte, session *Session) ([]*PageviewDataT, error) {
	sdb, err := getShardDB(collection.ID)
	if err != nil {
		return nil, err
	}

	ret := []*PageviewDataT{}

	fromKey := sessionKey[:8]
	toKey := GetKeyFromTime(time.Unix(0, session.End+1))
	idSuffix := sessionKey[8:]

	pageview := &Pageview{}
	sdb.Iterate(BPageview, fromKey, toKey, func(k []byte, v []byte) {
		if bytes.HasSuffix(k, idSuffix) {
			if err := protoDecode(v, pageview); err != nil {
				log.Println(err, v)
				return
			}
			ret = append(ret, &PageviewDataT{
				Time:        GetTimeFromKey(k).UnixNano(),
				Path:        pageview.Path,
				ReferrerURL: pageview.ReferrerURL,
			})
		}
	})
	return ret, nil
}
