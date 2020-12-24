package ytdl

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	e "github.com/sam1677/ytdl/internal/ytdlerrors"
)

//DownloadFile Downloads file from given URL to path
//
//if onlyData is true
//file won't saved and just return data of it
func DownloadFile(URL string, path string, filename string, onlyData bool) (file *os.File, data []byte, err error) {
	if !onlyData {
		fmt.Println(filename, "Start downloading from :", URL)
		defer fmt.Println(filename, "End downloading from :", URL)
	}

	res, err := http.Get(URL)
	if err != nil {
		return nil, nil, err
	}

	data, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, nil, e.DbgErr(err)
	}

	defer res.Body.Close()

	if onlyData {
		return nil, data, nil
	}

	err = CreatePath(path)
	if err != nil {
		return nil, nil, err
	}

	filename = MergePathAndFilename(path, filename)

	file, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0755)
	if err != nil {
		return nil, nil, err
	}

	_, err = file.Write(data)
	if err != nil {
		return nil, nil, err
	}

	return file, data, nil
}

//CreatePath Check if folder exists and if does not it creates folder
func CreatePath(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(path, 0755)
		fmt.Println("Folder", path, "Created")
		if err != nil {
			return e.DbgErr(err)
		}
	}
	return nil
}

//MergePathAndFilename Merges path and filename
//
//	MergePathAndFilename("./path1", "file")  => "./path1/file"
//	MergePathAndFilename("./path2/", "file") => "./path2/file"
//	MergePathAndFilename("./path3", "/file")  => "./path3/file"
//	MergePathAndFilename("./path4/", "/file") => "./path4/file"
func MergePathAndFilename(path, filename string) string {
	pathFormat := "%s/%s"
	pathEndsWithPS := os.IsPathSeparator(path[len(path)-1])
	fnameStartsWithPS := os.IsPathSeparator(filename[0])
	if pathEndsWithPS != fnameStartsWithPS {
		pathFormat = "%s%s"
	}

	filename = fmt.Sprintf(pathFormat, path, filename)
	if pathEndsWithPS && fnameStartsWithPS {
		filename = path + filename[1:]
	}
	return filename
}
