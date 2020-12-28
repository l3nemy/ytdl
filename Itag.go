package ytdl

//ContentType only Audio, Video or both
type ContentType int

const (
	//Audio only Audio
	Audio ContentType = iota
	//Video only Video
	Video
	//VandA Video includes Audio
	VandA
)

func (t ContentType) String() string {
	switch t {
	case Audio:
		return "audio"
	case Video:
		return "video"
	case VandA:
		return "video/audio"
	}
	return ""
}

//SpecialType is VR or HDR
type SpecialType int

const (
	//Non SpecialType
	Non SpecialType = iota
	//VR VR or 3D SpecialType
	VR
	//HDR HDR SpecialType
	HDR
)

func (t SpecialType) String() string {
	switch t {
	case Non:
		return "None"
	case VR:
		return "VR/3D"
	case HDR:
		return "HDR"
	}
	return ""
}

//ItagProp Itag's property
type ItagProp struct {
	FileType    string
	ContentType ContentType
	SpecialType SpecialType

	// video
	Resolution string

	// audio
	Bitrate string
}

const (
	mp3 = "mp3"
	mp4 = "mp4"
	flv = "flv"
	tgp = "3gp"
	wem = "webm"
	hls = "hls"
	m4a = "m4a"
	ops = "opus"
)

var itagList = map[int]ItagProp{
	5:   {flv, VandA, Non, "240p", ""},
	6:   {flv, VandA, Non, "270p", ""},
	17:  {tgp, VandA, Non, "144p", ""},
	18:  {mp4, VandA, Non, "360p", ""},
	22:  {mp4, VandA, Non, "720p", ""},
	34:  {flv, VandA, Non, "360p", ""},
	35:  {flv, VandA, Non, "480p", ""},
	36:  {tgp, VandA, Non, "180p", ""},
	37:  {mp4, VandA, Non, "1080p", ""},
	38:  {mp4, VandA, Non, "3072p", ""},
	43:  {wem, VandA, Non, "360p", ""},
	44:  {wem, VandA, Non, "480p", ""},
	45:  {wem, VandA, Non, "720p", ""},
	46:  {wem, VandA, Non, "1080p", ""},
	82:  {mp4, VandA, VR, "360p", ""},
	83:  {mp4, VandA, VR, "480p", ""},
	84:  {mp4, VandA, VR, "720p", ""},
	85:  {mp4, VandA, VR, "1080p", ""},
	92:  {hls, VandA, VR, "240p", ""},
	93:  {hls, VandA, VR, "360p", ""},
	94:  {hls, VandA, VR, "480p", ""},
	95:  {hls, VandA, VR, "720p", ""},
	96:  {hls, VandA, Non, "1080p", ""},
	100: {wem, VandA, VR, "360p", ""},
	101: {wem, VandA, VR, "480p", ""},
	102: {wem, VandA, VR, "720p", ""},
	132: {hls, VandA, VR, "240p", ""},
	133: {mp4, Video, Non, "240p", ""},
	134: {mp4, Video, Non, "360p", ""},
	135: {mp4, Video, Non, "480p", ""},
	136: {mp4, Video, Non, "720p", ""},
	137: {mp4, Video, Non, "1080p", ""},
	138: {mp4, Video, Non, "2160p60", ""},
	139: {m4a, Audio, Non, "", "48k"},
	140: {m4a, Audio, Non, "", "128k"},
	141: {m4a, Audio, Non, "", "256k"},
	151: {hls, VandA, Non, "72p", ""},
	160: {mp4, Video, Non, "144p", ""},
	167: {wem, Video, Non, "360p", ""},
	168: {wem, Video, Non, "480p", ""},
	169: {wem, Video, Non, "1080p", ""},
	171: {wem, Audio, Non, "", "128k"},
	218: {wem, Video, Non, "480p", ""},
	219: {wem, Video, Non, "144p", ""},
	242: {wem, Video, Non, "240p", ""},
	243: {wem, Video, Non, "360p", ""},
	244: {wem, Video, Non, "480p", ""},
	245: {wem, Video, Non, "480p", ""},
	246: {wem, Video, Non, "480p", ""},
	247: {wem, Video, Non, "720p", ""},
	248: {wem, Video, Non, "1080p", ""},
	249: {ops, Audio, Non, "", "50k"},
	250: {ops, Audio, Non, "", "70k"},
	251: {ops, Audio, Non, "", "160k"},
	264: {mp4, Video, Non, "1440p", ""},
	266: {mp4, Video, Non, "2160p60", ""},
	271: {wem, Video, Non, "1440p", ""},
	272: {wem, Video, Non, "2880p/4320p", ""},
	278: {wem, Video, Non, "144p", ""},
	298: {mp4, Video, Non, "720p60", ""},
	299: {mp4, Video, Non, "1080p60", ""},
	302: {wem, Video, Non, "720p60", ""},
	303: {wem, Video, Non, "1080p60", ""},
	308: {wem, Video, Non, "1440p60", ""},
	313: {wem, Video, Non, "2160p", ""},
	315: {wem, Video, Non, "2160p60", ""},
	330: {wem, Video, HDR, "144p60", ""},
	331: {wem, Video, HDR, "240p60", ""},
	332: {wem, Video, HDR, "360p60", ""},
	333: {wem, Video, HDR, "480p60", ""},
	334: {wem, Video, HDR, "720p60", ""},
	335: {wem, Video, HDR, "1080p60", ""},
	336: {wem, Video, HDR, "1440p60", ""},
	337: {wem, Video, HDR, "2160p60", ""},
	394: {mp4, Video, Non, "144p", ""},
	395: {mp4, Video, Non, "240p", ""},
	396: {mp4, Video, Non, "360p", ""},
	397: {mp4, Video, Non, "480p", ""},
	398: {mp4, Video, Non, "720p", ""},
	399: {mp4, Video, Non, "1080p", ""},
	400: {mp4, Video, Non, "1440p", ""},
	401: {mp4, Video, Non, "2160p", ""},
	402: {mp4, Video, Non, "2880p", ""},
}

//ToItagMap Converts []*Format into map[Itag]*Format
func (fl FormatList) ToItagMap() map[int]*Format {
	ret := map[int]*Format{}
	for _, f := range fl {
		ret[f.Itag] = f
	}
	return ret
}
