package ytdl

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"

	u "github.com/sam1677/ytdl/internal/utils"
	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

//VideoInfo Descibes Video's Informations (get_video_info?video_id=(videoID))
type VideoInfo struct {
	PlayabilityStatus struct {
		Status          string `json:"status"`
		PlayableInEmbed bool   `json:"playableInEmbed"`
	} `json:"playabilityStatus"`

	StreamingData struct {
		ExpiresInSeconds string    `json:"expiresInSeconds"`
		Formats          []*Format `json:"formats"`
		AdaptiveFormats  []*Format `json:"adaptiveFormats"`
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
func GetVideoInfo(VID string) (*VideoInfo, error) {
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

	err = ioutil.WriteFile(fmt.Sprintf("%s%s.json", tmpJSONDir, VID), rawJSON, 0755)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	vi := new(VideoInfo)

	err = json.Unmarshal(rawJSON, vi)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	err = vi.DecipherAll()
	if err != nil {
		return nil, e.DbgErr(err)
	}

	return vi, nil
}
