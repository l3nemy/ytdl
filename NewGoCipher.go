package ytdl

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"

	u "github.com/sam1677/ytdl/internal/utils"
	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

//DecipherFormat deciphers given Format
func (vi *VideoInfo) DecipherFormat(f *Format) error {
	_, data, err := u.DownloadFile(vi.Microformat.PlayerMicroformatRenderer.Embed.IframeURL, "", "", true)
	if err != nil {
		return e.DbgErr(err)
	}

	err = u.CreatePath(tmpScriptDir)
	if err != nil {
		return e.DbgErr(err)
	}

	err = ioutil.WriteFile(tmpScriptDir+"outerHtml.html", data, 0755)
	if err != nil {
		return e.DbgErr(err)
	}

	//r := regexp.MustCompile(`<\s*script[^<>]*name\s*=\s*"player_ias/base"[^<>]*>[^<>]*<\s*/script\s*>`)
	//r1 := regexp.MustCompile(`src="([^" ]*\.js)"`)

	//matched := r.FindStringSubmatch(string(data))
	//if len(matched) == 0 {
	//	return errors.New("An Error Occured on Loading Script")
	//}

	//scriptStr := matched[0]
	//scriptSrc := r1.FindStringSubmatch(scriptStr)[1]
	//scriptURL := "https://www.youtube.com" + scriptSrc

	r := regexp.MustCompile(`"(\/[^"<>]*\/base.js)"`)

	//TODO: Change it to RegexpSearch function
	scriptSrc := string(r.FindSubmatch(data)[1])
	scriptURL := "https://www.youtube.com" + scriptSrc

	_, data, err = u.DownloadFile(scriptURL, "", "", true)
	if err != nil {
		return e.DbgErr(err)
	}

	err = ioutil.WriteFile(tmpScriptDir+"script.js", data, 0755)
	if err != nil {
		return e.DbgErr(err)
	}

	query, err := url.ParseQuery(f.SignatureCipher)
	if err != nil {
		return e.DbgErr(err)
	}

	signature := query.Get("s")
	sigKey := query.Get("sp")
	cipheredURL := query.Get("url")

	URL, err := url.Parse(cipheredURL)
	if err != nil {
		return e.DbgErr(err)
	}

	val := URL.Query()

	cip, err := newDecipherer(string(data))
	if err != nil {
		return e.DbgErr(err)
	}

	sig, err := cip.GetSignature(signature)
	if err != nil {
		return e.DbgErr(err)
	}

	val.Add(sigKey, sig)
	URL.RawQuery = val.Encode()

	f.URL = URL.String()
	f.SignatureCipher = ""

	return nil
}

//DecipherAll executes DecipherFormat for all Format in the VideoInfo
func (vi *VideoInfo) DecipherAll() error {
	var err error
	for _, sd := range [][]*Format{vi.StreamingData.Formats, vi.StreamingData.AdaptiveFormats} {
		for _, f := range sd {
			if f.URL == "" || f.SignatureCipher != "" {
				err = vi.DecipherFormat(f)
				if err != nil {
					return e.DbgErr(err)
				}
			}

			//Additional Commands
			typ, _, err := mime.ParseMediaType(f.MimeType)
			if err != nil {
				return e.DbgErr(err)
			}
			typLst := strings.Split(typ, "/")
			f.Type = typLst[0]
			fileType := typLst[len(typLst)-1]
			if f.Type == "audio" {
				fileType = "mp3"
				if strings.Contains(f.MimeType, "opus") {
					fileType = "opus"
				}
			}
			f.Parent = vi
			format := "%s-%s-%s.%s"
			if f.QualityLabel == "" {
				format = "%s%s-%s.%s"
			}
			//fmt.Println(f.Type)
			//fmt.Println(f.Filename)
			f.Filename = fmt.Sprintf(format, vi.VideoDetails.VideoID, f.QualityLabel, f.Quality, fileType)
		}
	}
	return nil
}

type decipherer struct {
	TransformPlan   []string
	TransformObject []string
	TransformMap    map[string]interface{}
	js              string
	noIndentJS      string
	logger          *log.Logger
}

var latestCipher *decipherer

//newDecipherer makes cipher
func newDecipherer(js string) (*decipherer, error) {
	if latestCipher != nil { // if executed this session return latest cipher
		return latestCipher, nil
	}
	if js == "" {
		return nil, e.DbgErr(e.ErrJSstrIsEmpty)
	}

	c := new(decipherer)
	//Create Log File
	logFile, err := os.OpenFile("Cipher.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0755)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	c.logger = log.New(logFile, "", log.Ltime|log.Lmicroseconds)
	c.js = js
	c.noIndentJS = strings.ReplaceAll(js, "\n", " ")

	c.TransformPlan, err = c.getTransformPlan()
	if err != nil {
		return nil, e.DbgErr(err)
	}

	firstFuncName := strings.Split(c.TransformPlan[0], ".")[0]

	//fmt.Println(c.TransformPlan)

	c.TransformObject, err = c.getTransformObject(firstFuncName)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	c.TransformMap, err = c.getTransformMap(firstFuncName)
	if err != nil {
		return nil, e.DbgErr(err)
	}

	latestCipher = c

	return c, nil
}

//GetSignature converts cipheredSignature to decipheredSignature
func (c *decipherer) GetSignature(cipheredSignature string) (string, error) {
	c.logf("Start Getting signature : \n\t\tfrom : %s", cipheredSignature)
	defer c.logf("End Getting signature : \n\t\tfrom : %s", cipheredSignature)

	var signature []string
	signature = strings.Split(cipheredSignature, "")
	for _, jsFunc := range c.TransformPlan {
		name, argument, err := c.parseFunction(jsFunc)
		if err != nil {
			return "", e.DbgErr(err)
		}

		fun, ex := c.TransformMap[name]
		if !ex {
			return "", e.DbgErr(errors.New(name + "'s function is not found"))
		}
		if fun == nil {
			return "", e.DbgErr(errors.New("fun is nil : " + name))
		}

		signature = fun.(func(args ...interface{}) []string)(signature, argument)
		c.logf(
			"applied transform function\n"+
				"\t\toutput: %v\n"+
				"\t\tjsFunction: %s\n"+
				"\t\targument: %d\n"+
				"\t\tfunction: %s\n"+
				"",
			signature,
			name,
			argument,
			c.TransformMap[name],
		)
	}
	return strings.Join(signature, ""), nil
}

func (c *decipherer) parseFunction(jsFunc string) (name string, arg int, err error) {
	c.logf("Start parsing TransformFunc : %s", jsFunc)
	defer c.logf("End parsing TransformFunc : %s", jsFunc)

	matches, err := c.regexpSearchAll(`\w+\.(\w+)\(\w,(\d+)\)`, jsFunc)
	if err != nil {
		return "", 0, e.DbgErr(err)
	}
	if len(matches) < 3 {
		return "", 0, e.DbgErr(e.ErrFuncListIsTooShort)
	}

	arg, err = strconv.Atoi(matches[2])
	if err != nil {
		return "", 0, e.DbgErr(err)
	}

	return matches[1], arg, nil
}

func (c *decipherer) getRootFuncName() (string, error) {
	funcPatterns := []string{
		`\b[cs]\s*&&\s*[adf]\.set\([^,]+\s*,\s*encodeURIComponent\s*\(\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
		`\b[a-zA-Z0-9]+\s*&&\s*[a-zA-Z0-9]+\.set\([^,]+\s*,\s*encodeURIComponent\s*\(\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
		`(?:\b|[^a-zA-Z0-9$])(?P<sig>[a-zA-Z0-9$]{2})\s*=\s*function\(\s*a\s*\)\s*{\s*a\s*=\s*a\.split\(\s*""\s*\)`,
		`(?P<sig>[a-zA-Z0-9$]+)\s*=\s*function\(\s*a\s*\)\s*{\s*a\s*=\s*a\.split\(\s*""\s*\)`,
		`(["\'])signature["\']\s*,\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
		`\.sig\|\|(?P<sig>[a-zA-Z0-9$]+)\(`,
		`yt\.akamaized\.net/\)\s*\|\|\s*.*?\s*[cs]\s*&&\s*[adf]\.set\([^,]+\s*,\s*(?:encodeURIComponent\s*\()?\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
		`\b[cs]\s*&&\s*[adf]\.set\([^,]+\s*,\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
		`\b[a-zA-Z0-9]+\s*&&\s*[a-zA-Z0-9]+\.set\([^,]+\s*,\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
		`\bc\s*&&\s*a\.set\([^,]+\s*,\s*\([^)]*\)\s*\(\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
		`\bc\s*&&\s*[a-zA-Z0-9]+\.set\([^,]+\s*,\s*\([^)]*\)\s*\(\s*(?P<sig>[a-zA-Z0-9$]+)\(`,
	}
	c.log("Start finding first function name")
	defer c.log("End finding first function name")

	for _, pat := range funcPatterns {
		match, err := c.regexpSearch(pat, c.js, 1)
		if err != nil {
			if e.IsRegexpErr(err) {
				continue
			}
			return "", e.DbgErr(err)
		}
		c.log("Root Func Name: " + match)
		return match, nil
	}
	return "", e.DbgErr(e.ErrRegexpNotMatched)
}

func (c *decipherer) getTransformPlan() ([]string, error) {
	c.log("Start getting TransformPlan")

	fname, err := c.getRootFuncName()
	if err != nil {
		return nil, e.DbgErr(err)
	}

	funcName := url.QueryEscape(fname)

	match, err := c.regexpSearch(
		fmt.Sprintf(`%s=function\(\w\){[a-z=\.\(\"\)]*;(.*);(?:return.+)}`, funcName), c.js, 1,
	)
	if err != nil {
		return nil, e.DbgErr(err)
	}
	c.logf("End getting TransformPlan : %s", match)
	return strings.Split(match, ";"), nil
}

func (c *decipherer) getTransformObject(firstFuncName string) ([]string, error) {
	c.log("Start getting TransformObject")
	defer c.log("End getting TramsformObject")

	exp := fmt.Sprintf(`var %s={([\w,;%%.()[\]:={}\s]+)};`, url.QueryEscape(firstFuncName))

	match, err := c.regexpSearch(exp, c.noIndentJS, 1)
	if err != nil {
		return nil, e.DbgErr(err)
	}
	//fmt.Println(match)

	match = strings.ReplaceAll(match, "\n", " ")
	sp := strings.Split(match, "},")
	ret := []string{}
	for i, r := range sp {
		if i != len(sp)-1 {
			r += "}"
		}
		ret = append(ret, r)
	}

	return ret, nil
}

func (c *decipherer) getTransformMap(firstFuncName string) (ret map[string]interface{}, err error) {
	c.log("Start getting TransformMap")
	defer c.log("End getting TransformMap")

	ret = map[string]interface{}{}
	for _, obj := range c.TransformObject {
		c.log(obj)

		obj = strings.TrimSpace(obj)

		//t is js function map string
		t := strings.SplitN(obj, ":", 2)
		c.log(t)
		if len(t) < 2 {
			return nil, errors.New("Function count is less than 2")
		}

		funcName := t[0]
		funcStr := t[1]

		fun, err := c.mapFunctions(funcStr)
		if err != nil {
			return nil, e.DbgErr(err)
		}

		ret[funcName] = fun
	}
	return ret, nil
}

func (c *decipherer) mapFunctions(jsFunc string) (interface{}, error) {
	c.logf("Start mapping functions : %s", jsFunc)
	defer c.logf("End mapping functions : %s", jsFunc)

	mapper := map[string]interface{}{
		// function(a){a.reverse()}
		`\w+\.reverse\(\)`: reverse,

		// function(a,b){a.splice(0,b)}
		`\w+\.splice\(0,\w\)`: splice,

		// function(a,b){var c=a[0];a[0]=a[b%a.length];a[b]=c}
		`var\s\w=\w\[0\];\w\[0\]=\w\[\w\%\w.length\];\w\[\w\]=\w`: swap,

		// function(a,b){var c=a[0];a[0]=a[b%a.length];a[b%a.length]=c}
		`var\s\w=\w\[0\];\w\[0\]=\w\[\w\%\w.length\];\w\[\w\%\w.length\]=\w`: swap,
	}
	for p, fn := range mapper {
		r := regexp.MustCompile(p)
		if r.MatchString(jsFunc) {
			return fn, nil
		}
	}
	return nil, e.DbgErr(e.ErrNoMatchOnFunction)
}

// Utils

func (c *decipherer) log(v ...interface{}) {
	c.logger.Print(v...)
}
func (c *decipherer) logf(format string, v ...interface{}) {
	c.logger.Printf(format, v...)
}

// Returns matched groups[group] group 0=> full 1=> first .....
func (c *decipherer) regexpSearch(exp, str string, group int) (string, error) {
	results, err := c.regexpSearchAll(exp, str)
	if err != nil {
		return "", e.DbgErr(err)
	}

	return results[group], nil
}

func (c *decipherer) regexpSearchAll(exp, str string) ([]string, error) {
	c.logf("Regexp Search Start : \n\t\texp : %s\n", exp)
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

func reverse(args ...interface{}) []string {
	arr := args[0].([]string)
	temparr := []string{}
	for i := 0; i < len(arr); i++ {
		temparr = append(temparr, arr[len(arr)-1-i])
	}
	return temparr
}

func splice(args ...interface{}) []string {
	arr := args[0].([]string)
	b := args[1].(int)
	return append(atoB(arr, 0, b), atoB(arr, b*2, len(arr))...)
}

func swap(args ...interface{}) []string {
	arr := args[0].([]string)
	b := args[1].(int)
	r := b % len(arr)
	//[]string{arr[r]} + AtoB(arr, 1, r) + []string{arr[0]} + AtoB(arr, r+1, len(arr))
	return append(append([]string{arr[r]}, atoB(arr, 1, r)...), append([]string{arr[0]}, atoB(arr, r+1, len(arr))...)...)
}

func atoB(arr []string, a, b int) []string {
	tempArr := []string{}
	for i := a; i < b; i++ {
		tempArr = append(tempArr, arr[i])
	}
	return tempArr
}
