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
		[]*bucketSumT{&bucketSumT{begin.Unix(), 0}},
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
	validSessions := map[uint32]empty{}

	bg := CreateBucketGen(input.Bucket, input.From, input.To, input.Timezone)
	sdb.Iterate(BSession, fromKey, toKey, func(k []byte, v []byte) {
		t, err := unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		if len(input.Filter) != 0 {
			if !matchFilter(input.Filter, session) {
				return
			}
			validSessions[GetIdFromKey(k)] = empty{}
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
		if len(input.Filter) != 0 {
			if _, ok := validSessions[GetIdFromKey(k)]; !ok {
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

func matchFilter(filter map[string]string, session *Session) bool {
	if v, ok := filter["hostname"]; ok && session.Hostname != v {
		return false
	}
	if v, ok := filter["device_type"]; ok && session.DeviceType != v {
		return false
	}
	if v, ok := filter["device_os"]; ok && session.DeviceOS != v {
		return false
	}
	if v, ok := filter["browser_name"]; ok && session.BrowserName != v {
		return false
	}
	if v, ok := filter["browser_version"]; ok && session.BrowserVersion != v {
		return false
	}
	if v, ok := filter["browser_language"]; ok && session.BrowserLanguage != v {
		return false
	}

	if v, ok := filter["pageview_count"]; ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			log.Println("bad pageview_count int filter", v)
		} else if session.PageviewCount != int32(i) {
			return false
		}
	}
	if v, ok := filter["screen_resolution"]; ok && session.ScreenResolution != v {
		return false
	}
	if v, ok := filter["window_resolution"]; ok && session.WindowResolution != v {
		return false
	}
	if v, ok := filter["country_code"]; ok && session.CountryCode != v {
		return false
	}
	if v, ok := filter["city"]; ok && session.City != v {
		return false
	}
	if v, ok := filter["as_name"]; ok && session.ASName != v {
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

	session := &Session{}

	prevKey := marshalTime(prevTime)
	fromKey := marshalTime(input.From)
	toKey := marshalTime(input.To)

	sdb.Iterate(BSession, prevKey, fromKey, func(k []byte, v []byte) {
		t, err := unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		prevSessionTotal++
		sumOfPrevSessionLength += int(session.End/1000000000 - t.UnixNano()/1000000000)

		prevPageviewCountSums[strconv.Itoa(int(session.PageviewCount))]++
	})

	sdb.Iterate(BPageview, prevKey, fromKey, func(k []byte, v []byte) {
		prevPageviewTotal++
	})

	validSessions := map[uint32]empty{}

	sdb.Iterate(BSession, fromKey, toKey, func(k []byte, v []byte) {
		t, err := unmarshalTime(k)
		if err != nil {
			log.Println("bad key: ", k)
			return
		}
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		if len(input.Filter) != 0 {
			if !matchFilter(input.Filter, session) {
				return
			}
			validSessions[GetIdFromKey(k)] = empty{}
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

	pageview := &Pageview{}
	sdb.Iterate(BPageview, fromKey, toKey, func(k []byte, v []byte) {
		if err := protoDecode(v, pageview); err != nil {
			log.Println(err, v)
			return
		}

		if len(input.Filter) != 0 {
			if _, ok := validSessions[GetIdFromKey(k)]; !ok {
				return
			}
		}

		pageviewTotal++

		pageSums[pageview.Path]++
		referrerSums[pageview.ReferrerURL]++
	})

	avgSessionLength := sumOfSessionLength / coalesce(sessionTotal)
	prevAvgSessionLength := sumOfPrevSessionLength / coalesce(prevSessionTotal)

	bounceRate := getPercentByKey(&pageviewCountSums, "1")
	prevBounceRate := getPercentByKey(&prevPageviewCountSums, "1")

	return &CollectionStatDataT{
		SessionTotal:         totalT{sessionTotal, float32(sessionTotal)/float32(coalesce(prevSessionTotal)) - 1.0},
		PageviewTotal:        totalT{pageviewTotal, float32(pageviewTotal)/float32(coalesce(prevPageviewTotal)) - 1.0},
		AvgSessionLength:     totalT{avgSessionLength, float32(avgSessionLength)/float32(coalesce(prevAvgSessionLength)) - 1.0},
		BounceRate:           percentT{bounceRate, bounceRate/prevBounceRate - 1.0},
		PageSums:             getSums(&pageSums),
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

func coalesce(n int) int {
	if n != 0 {
		return n
	}
	return 1
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

	session := &Session{}
	sdb.Iterate(BSession, fromKey, toKey, func(k []byte, v []byte) {
		if err := protoDecode(v, session); err != nil {
			log.Println(err, v)
			return
		}
		if len(input.Filter) != 0 && !matchFilter(input.Filter, session) {
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
