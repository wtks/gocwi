package api

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var (
	client *http.Client
	jar    *cookiejar.Jar
)

func init() {
	jar, _ = cookiejar.New(nil)
	client = &http.Client{Jar: jar}
}

func getHiddenFormValues(html string) (url.Values, error) {
	result := url.Values{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return nil, err
	}

	doc.Find("form input[type='hidden']").Each(func(i int, s *goquery.Selection) {
		name, _ := s.Attr("name")
		value, _ := s.Attr("value")
		result.Add(name, value)
	})

	return result, nil
}

func getMatrixPositions(html string) ([3][]string, error) {
	result := [3][]string{}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return result, err
	}

	doc.Find(`tr > th[align="left"]`).Each(func(i int, s *goquery.Selection) {
		if i >= 2 && i <= 4 {
			result[i-2] = strings.SplitN(strings.Trim(s.Text(), "[]"), ",", 2)
		}
	})

	return result, nil
}

func DownloadFile(url string, dest string, bar *pb.ProgressBar) error {
	res, err := client.Get(url)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("Server return non-200 status: %v\n", res.Status)
	}

	file, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer file.Close()

	if bar != nil {
		bar.Total, _ = strconv.ParseInt(res.Header.Get("Content-Length"), 10, 64)
		bar.Start()

		proxy := bar.NewProxyReader(res.Body)

		_, err = io.Copy(file, proxy)

		bar.Finish()
	} else {
		_, err = io.Copy(file, res.Body)
	}

	return err
}
