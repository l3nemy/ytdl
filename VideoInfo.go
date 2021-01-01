package ytdl

import (
	"encoding/json"
	"fmt"
	"net/url"
	"reflect"
	"regexp"
	"sort"

	u "github.com/sam1677/ytdl/internal/utils"
	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

//FormatList describes FormatPtr slice
type FormatList []*Format

// Audios gets audio list from given FormatList
func (fl FormatList) Audios() FormatList {
	return fl.contentType(Audio)
}

// Videos gets video list from given FormatList
func (fl FormatList) Videos() FormatList {
	return fl.contentType(Video | VandA)
}

func (fl FormatList) contentType(ct ContentType) FormatList {
	ret := FormatList{}
	for _, f := range fl {
		switch {
		case ct&Audio != 0:
			if f.ItagProp.ContentType == Audio {
				ret = append(ret, f)
			}
		case ct&Video != 0:
			if f.ItagProp.ContentType == Video {
				ret = append(ret, f)
			}
		case ct&VandA != 0:
			if f.ItagProp.ContentType == VandA {
				ret = append(ret, f)
			}
		}
	}
	return ret
}

// First gets first element from given FormatList
func (fl FormatList) First() *Format {
	if len(fl) == 0 {
		return nil
	}
	return fl[0]
}

// Last gets last element from given FormatList
func (fl FormatList) Last() *Format {
	if len(fl) == 0 {
		return nil
	}
	return fl[len(fl)-1]
}

// Worst gets worst quality element from given FormatList
func (fl FormatList) Worst() *Format {
	fl.SortByFieldName("Bitrate")
	return fl.SortByQuality().First()
}

// Best gets best quality element from given FormatList
func (fl FormatList) Best() *Format {
	return fl.SortByQuality().Last()
}

// Reverse gets reversed FormatList of given FormatList
func (fl FormatList) Reverse() FormatList {
	temp := FormatList{}
	size := len(temp) - 1
	for i, f := range fl {
		temp[size-i] = f
	}
	return temp
}

// Sort sorts FormatList By Itag
func (fl FormatList) Sort() FormatList {
	return fl.SortByFieldName("Itag")
}

// SortByQuality sorts FormatList By Bitrate
func (fl FormatList) SortByQuality() FormatList {
	return fl.SortByFieldName("Bitrate")
}

// SortByFieldName sorts FormatList by given field's value
func (fl FormatList) SortByFieldName(fieldName string) FormatList {
	sort.Slice(fl, func(i, j int) bool {
		ir := checkNGetElem(reflect.ValueOf(fl[i])).FieldByName(fieldName)
		jr := checkNGetElem(reflect.ValueOf(fl[j])).FieldByName(fieldName)

		switch ir.Interface().(type) {
		case int:
			return ir.Int() < jr.Int()
		case uint:
			return ir.Uint() < jr.Uint()
		case string:
			return ir.String() < jr.String()
		default:
			panic(e.DbgErr(e.ErrFieldType))
		}
	})
	return fl
}

func checkNGetElem(v reflect.Value) reflect.Value {
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
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
		return nil, e.DbgErr(err)
	}

	URL := fmt.Sprintf(getVideoInfoURL, VID)

	_, data, err := u.DownloadFile(URL, tmpVideoInfoDir, VID, !VideoInfoCaching)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	val, err := url.ParseQuery(string(data))
	if err != nil {
		return nil, e.DbgErr(err)
	}

	if val.Get("status") == "fail" {
		return nil, e.DbgErr(e.ErrVIDIsInvalid)
	}

	rawJSON := []byte(val.Get("player_response"))

	if JSONCaching {
		_, err = u.WriteFile(rawJSON, tmpJSONDir, VID+".json")
		if err != nil {
			return nil, e.DbgErr(err)
		}
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

// CombinedFormatList returns FormatList combined Formats and AdaptiveFormats
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
