package ytdl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"regexp"

	u "github.com/sam1677/ytdl/internal/utils"
	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

type FormatList []*Format

func (fl FormatList) Sort() FormatList {
	mapp := fl.ToItagMap()
	sm := u.SortByMapIntKey(mapp, false)

	m := map[int]*Format{}
	for key, val := range sm {
		m[key] = val.(*Format)
	}

	ret := FormatList{}

	for _, v := range u.MapValues(reflect.ValueOf(m)) {
		ret = append(ret, v.(*Format))
	}

	return ret
}

//VideoInfo Descibes Video's Informations (get_video_info?video_id=(videoID))
type VideoInfo struct {
	PlayabilityStatus struct {
		Status          string `json:"status"`
		PlayableInEmbed bool   `json:"playableInEmbed"`
	} `json:"playabilityStatus"`

	StreamingData struct {
		ExpiresInSeconds string     `json:"expiresInSeconds"`
		Formats          FormatList `json:"formats"`
		AdaptiveFormats  FormatList `json:"adaptiveFormats"`
	} `json:"streamingData"`

	VideoDetails struct {
		VideoID           string    `json:"videoId"`
		Title             string    `json:"title"`
		LengthSeconds     string    `json:"lengthSeconds"`
		Keywords          []string  `json:"keywords"`
		ChannelID         string    `json:"channelId"`
		IsOwnerViewing    bool      `json:"isOwnerViewing"`
		IsCrawlable       bool      `json:"isCrawlable"`
		ShortDescription  string    `json:"shortDescription"`
		Thumbnail         Thumbnail `json:"thumbnail"`
		AverageRating     float64   `json:"averageRating"`
		AllowRatings      bool      `json:"allowRatings"`
		ViewCount         string    `json:"viewCount"`
		Author            string    `json:"author"`
		IsPrivate         bool      `json:"isPrivate"`
		IsUnpluggedCorpus bool      `json:"isUnpluggedCorpus"`
		IsLiveContent     bool      `json:"isLiveContent"`
	} `json:"videoDetails"`

	Microformat struct {
		PlayerMicroformatRenderer struct {
			Thumbnail Thumbnail `json:"thumbnail"`
			Embed     struct {
				IframeURL      string `json:"iframeUrl"`
				FlashURL       string `json:"flashUrl"`
				Width          int    `json:"width"`
				Height         int    `json:"height"`
				FlashSecureURL string `json:"flashSecureUrl"`
			} `json:"embed"`
			Title              SimpleText `json:"title"`
			Description        SimpleText `json:"description"`
			LengthSeconds      string     `json:"lengthSeconds"`
			OwnerProfileURL    string     `json:"ownerProfileUrl"`
			ExternalChannelID  string     `json:"externalChannelId"`
			AvailableCountries []string   `json:"availableCountries"`
			IsUnlisted         bool       `json:"isUnlisted"`
			HasYpcMetadata     bool       `json:"hasYpcMetadata"`
			ViewCount          string     `json:"viewCount"`
			Category           string     `json:"category"`
			PublishDate        string     `json:"publishDate"`
			OwnerChannelName   string     `json:"ownerChannelName"`
			UploadDate         string     `json:"uploadDate"`
		} `json:"playerMicroformatRenderer"`
	} `json:"microformat"`
}

//Format describes video type format
type Format struct {
	Itag              int    `json:"itag"`
	URL               string `json:"url"`
	MimeType          string `json:"mimeType"`
	Bitrate           int    `json:"bitrate"`
	Width             int    `json:"width"`
	Height            int    `json:"height"`
	LastModified      string `json:"lastModified"`
	ContentLength     string `json:"contentLength"`
	Quality           string `json:"quality"`
	FPS               int    `json:"fps"`
	QualityLabel      string `json:"qualityLabel"`
	ProjectionType    string `json:"projectionType"`
	AverageBitrate    int    `json:"averageBitrate"`
	ApproxDurationsMs string `json:"approxDurationsMs,omitempty"`
	SignatureCipher   string `json:"signatureCipher,omitempty"`
	FormatAudio

	//User Field (JSON Ignored)
	Filename string     `json:"-"`
	Parent   *VideoInfo `json:"-"`
	Type     string     `json:"-"`
	ItagProp ItagProp   `json:"-"`
}

func (f *Format) String() string {
	fLength := "ContentLength: %s (ApproxDurations: %s ms)\n"
	if f.ApproxDurationsMs == "" {
		fLength = "ContentLength: %s%s\n"
	}

	switch f.ItagProp.ContentType {
	case Video, VandA:
		return fmt.Sprintf(
			"\nItag : %d\n"+
				"Mime: %s\n"+
				"Quality: %s (Label: %s, FPS: %d)\n"+
				"Bitrate: %d (Average: %d)\n"+
				"Size: (%d x %d)\n"+
				fLength+
				"ExpectedDefaultFilename: %s\n",
			f.Itag,
			f.MimeType,
			f.Quality, f.QualityLabel, f.FPS,
			f.Bitrate, f.AverageBitrate,
			f.Width, f.Height,
			f.ContentLength, f.ApproxDurationsMs,
			f.Filename,
		)
	case Audio:
		return fmt.Sprintf(
			"\nItag : %d\n"+
				"Mime: %s\n"+
				"Quality: %s\n"+
				"Bitrate: %d (Average: %d)\n"+
				fLength+
				"ExpectedDefaultFilename: %s\n",
			f.Itag,
			f.MimeType,
			f.Quality,
			f.Bitrate, f.AverageBitrate,
			f.ContentLength, f.ApproxDurationsMs,
			f.Filename,
		)
	}

	return ""
}

//FormatAudio is optional field in Format
type FormatAudio struct {
	AudioQuality    string `json:"audioQuality,omitempty"`
	AudioSampleRate string `json:"audioSampleRate,omitempty"`
	AudioChannels   int    `json:"audioChannels,omitempty"`
}

//SimpleText for Simplize structs
type SimpleText struct {
	SimpleText string `json:"simpleText"`
}

//Thumbnail describes thumbnail JSON type
type Thumbnail struct {
	Thumbnails []struct {
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"thumbnails"`
}

//GetVideoInfo Gets get_video_info file From Youtube
func GetVideoInfo(VIDorURL string) (*VideoInfo, error) {

	VID, err := getVideoIDFromURL(VIDorURL)
	if err != nil {
		return nil, err
	}

	URL := fmt.Sprintf(getVideoInfoURL, VID)

	res, err := http.Get(URL)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, e.DbgErr(err)
	}
	defer res.Body.Close()

	temp, err := url.QueryUnescape(string(data))
	if err != nil {
		return nil, e.DbgErr(err)
	}

	r := regexp.MustCompile(`player_response=({.*})`)
	matches := r.FindStringSubmatch(temp)

	rawJSON := []byte(matches[1])

	err = u.CreatePath(tmpJSONDir)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	err = ioutil.WriteFile(fmt.Sprintf("%s.json", u.MergePathAndFilename(tmpJSONDir, VID)), rawJSON, 0755)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	vi := new(VideoInfo)

	err = json.Unmarshal(rawJSON, vi)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	err = vi.decipherAll()
	if err != nil {
		return nil, e.DbgErr(err)
	}

	return vi, nil
}

func (vi *VideoInfo) CombinedFormatList() FormatList {
	fl := *new(FormatList)
	fl = append(fl, vi.StreamingData.Formats...)
	fl = append(fl, vi.StreamingData.AdaptiveFormats...)
	return fl
}

func getVideoIDFromURL(URL string) (string, error) {
	rs := []string{
		`https?:\/\/www\.youtube\.com\/watch\?v=(\w+)`,
		`https?:\/\/youtu\.be\/(\w+)`,
		`(\w+)`,
	}
	for _, r := range rs {
		ret, err := regexpSearch(r, URL, 1)
		if err != nil {
			if e.IsRegexpErr(err) {
				continue
			}
			return "", e.DbgErr(err)
		}
		return ret, nil
	}
	return URL, e.DbgErr(e.ErrRegexpNotMatched)
}

func regexpSearch(exp, str string, ind int) (string, error) {
	ret, err := regexpSearchAll(exp, str)
	if err != nil {
		return "", e.DbgErr(err)
	}

	return ret[ind], nil
}

func regexpSearchAll(exp, str string) ([]string, error) {
	r, err := regexp.Compile(exp)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	results := r.FindStringSubmatch(str)
	if results == nil || len(results) == 0 {
		//fmt.Println("\"", exp, "\"", "\n", str)
		return nil, e.DbgErr(e.ErrRegexpNotMatched)
	}

	return results, nil
}
