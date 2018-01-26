package api

import (
	"errors"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	urlOcwiRoot         = "https://secure.ocw.titech.ac.jp/ocwi/index.php"
	urlOcwiPageNotFound = urlOcwiRoot + "?module=Default&action=PageNotFound"
)

var (
	lecturerRegExp      = regexp.MustCompile(`\(.+?\)`)
	lectureIdRegExp     = regexp.MustCompile(`&JWC=(\d+?)$`)
	attachmentRegExp    = regexp.MustCompile(`(.+?)（(\d+?KB)）(.+?) (\d{4})\.(\d{2})\.(\d{2})`)
	attachmentExtRegExp = regexp.MustCompile(`.+?&file=.+?\.(.+?)&JWC=.+?`)
	kadaiIdRegExp       = regexp.MustCompile(`&kadaiid=(\d+?)`)
)

type SubjectListResult struct {
	Terms []Term
}

type Term struct {
	Name     string
	Subjects []Subject
}

type Subject struct {
	Id               int
	Name             string
	Periods          []string
	Lecturers        []string
	Rooms            []string
	LastUpdated      string
	OpenTaskCount    int
	ExamSchedule     []string
	IsNotesAvailable bool
}

type LectureNoteResult struct {
	SubjectName   string
	SubjectNameEn string
	Id            int
	Classes       []Class
}

type Class struct {
	Title             string
	Date              string
	Room              string
	Type              string
	IsRoomChanged     bool
	IsCanceled        bool
	Attachments       []Attachment
	AttachmentComment string
}

type Attachment struct {
	Url   string
	Title string
	Type  string
	Ext   string
	Size  string
	Year  int
	Month int
	Day   int
}

type Task struct {
	Status       string
	Deadline     time.Time
	SubmitTime   *time.Time
	SubjectName  string
	SubjectId    int
	Title        string
	Id           int
	Lecturers    []string
	OpenDateTime time.Time
}

func LoginOcwi() error {
	req, err := http.NewRequest("GET", urlOcwiRoot, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Referer", urlPortalMenu)

	res, err := client.Do(req)
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func LogoutOcwi() error {
	values := url.Values{}
	values.Add("module", "Ocwi")
	values.Add("action", "Logout")

	res, err := client.Get(urlOcwiRoot + "?" + values.Encode())
	if err != nil {
		return err
	}
	res.Body.Close()
	return nil
}

func GetLectureList() (*SubjectListResult, error) {
	values := url.Values{}
	values.Add("module", "Ocwi")
	values.Add("action", "LectureList")

	res, err := client.Get(urlOcwiRoot + "?" + values.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return parseLectureListHtml(res.Body)
}

func parseLectureListHtml(reader io.Reader) (*SubjectListResult, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	result := &SubjectListResult{}
	tables := doc.Find("div#mainarea > div.contents > table")
	if tables.Length() == 0 {
		return nil, errors.New("something is wrong")
	}
	tables.Each(func(_ int, q *goquery.Selection) {
		quarter := Term{}

		quarter.Name = q.Find(`tr > th`).First().Text()
		q.Find("tr + input[type='hidden']").Each(func(_ int, i *goquery.Selection) {
			id, _ := strconv.Atoi(i.AttrOr("value", "ERROR"))
			quarter.Subjects = append(quarter.Subjects, Subject{Id: id})
		})
		q.Find("tr").Has("td").Each(func(i int, l *goquery.Selection) {
			lecture := &quarter.Subjects[i]
			l.Find("td").Each(func(j int, d *goquery.Selection) {
				switch j {
				case 0: //講義名
					if a := d.Find("a"); a.Size() != 0 {
						lecture.Name = a.Text()
						lecture.IsNotesAvailable = true
					} else {
						lecture.Name = strings.TrimSpace(d.Text())
						lecture.IsNotesAvailable = false
					}
				case 1: //時限
					lecture.Periods = strings.Split(strings.TrimSpace(d.Text()), "\n")
				case 2: //教員
					teachers := lecturerRegExp.FindAllString(d.Text(), -1)
					for i := range teachers {
						teachers[i] = strings.Trim(teachers[i], "()")
					}
					lecture.Lecturers = teachers
				case 3: //講義室
					lecture.Rooms = strings.Split(strings.TrimSpace(d.Text()), "\n")
				case 4: //更新日時
					lecture.LastUpdated = d.Text()
				case 5: //受付課題数
					if len(d.Text()) > 0 {
						if n, err := strconv.Atoi(d.Text()); err == nil {
							lecture.OpenTaskCount = n
						}
					}
				case 6: //試験・補講日程
					d.Find("span").Has("span").Each(func(_ int, s *goquery.Selection) {
						lecture.ExamSchedule = append(lecture.ExamSchedule, strings.Join(strings.Fields(s.Text()), " "))
					})
				}
			})
		})
		result.Terms = append(result.Terms, quarter)
	})

	return result, nil
}

func GetLectureNote(id int) (*LectureNoteResult, error) {
	values := url.Values{}
	values.Add("module", "Ocwi")
	values.Add("action", "KougiNote")
	values.Add("JWC", strconv.Itoa(id))

	res, err := client.Get(urlOcwiRoot + "?" + values.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.Request.URL.String() == urlOcwiPageNotFound {
		return nil, errors.New("the subject is not found")
	}

	return parseLectureNoteHtml(res.Body)
}

func parseLectureNoteHtml(reader io.Reader) (*LectureNoteResult, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	result := &LectureNoteResult{}

	h1 := doc.Find("#lectureTtl > h1")
	result.SubjectNameEn = h1.Find("div").Text()
	result.SubjectName = strings.TrimSuffix(h1.Text(), result.SubjectNameEn)
	result.Id, _ = strconv.Atoi(lectureIdRegExp.FindStringSubmatch(doc.Find("#lnaviExercise > a").AttrOr("href", ""))[1])

	notes := doc.Find("#mainarea > div.contents > div.lectureNote")
	notes.Each(func(_ int, note *goquery.Selection) {
		class := Class{}
		class.Title = note.Find("h2 > div").Text()
		class.Type = note.Find("h2 > img").AttrOr("alt", "不明")
		class.IsCanceled = false
		class.IsRoomChanged = false

		dateAndRoom := note.Find("ul.leftLine > li").First()
		if changed := dateAndRoom.Find("em"); changed.Length() > 0 {
			if attr, ok := note.Find("ul.leftLine > li > img").Attr("src"); ok && attr == "images/ico_cancel.gif" {
				class.IsCanceled = true
			} else if ok && attr == "images/ico_change.gif" {
				class.IsRoomChanged = true
			}
			arr := strings.Fields(strings.TrimSpace(changed.Text()))
			class.Date = arr[0]
			class.Room = arr[1]
		} else {
			arr := strings.Fields(strings.TrimSpace(dateAndRoom.Text()))
			class.Date = arr[0]
			class.Room = arr[1]
		}

		note.Find("ul.icon > li").Each(func(_ int, l *goquery.Selection) {
			if l.HasClass("file") {
				a := l.Find("a")
				detail := attachmentRegExp.FindStringSubmatch(strings.TrimSpace(a.Text()))
				at := Attachment{}
				at.Url = urlOcwiRoot + a.AttrOr("href", "")
				at.Ext = attachmentExtRegExp.FindStringSubmatch(at.Url)[1]
				at.Title = detail[1]
				at.Size = detail[2]
				at.Type = detail[3]
				at.Year, _ = strconv.Atoi(detail[4])
				at.Month, _ = strconv.Atoi(detail[5])
				at.Day, _ = strconv.Atoi(detail[6])

				class.Attachments = append(class.Attachments, at)
			} else if l.HasClass("comment") {
				class.AttachmentComment = strings.TrimSpace(l.Text())
			}
		})

		result.Classes = append(result.Classes, class)
	})

	return result, nil
}

func GetTaskList() ([]Task, error) {
	values := url.Values{}
	values.Add("module", "Ocwi")
	values.Add("action", "Subject")

	res, err := client.Get(urlOcwiRoot + "?" + values.Encode())
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	return parseTaskList(res.Body)
}

func parseTaskList(reader io.Reader) ([]Task, error) {
	doc, err := goquery.NewDocumentFromReader(reader)
	if err != nil {
		return nil, err
	}

	var result []Task

	doc.Find("div#container > div.contents > table > tr").Has("td").Each(func(_ int, t *goquery.Selection) {
		task := Task{}
		t.Find("td").Each(func(i int, s *goquery.Selection) {
			switch i {
			case 0: //提出状況
				task.Status = s.Find("a > img").AttrOr("alt", "不明")
			case 1: //締め切り
				task.Deadline, _ = time.Parse("2006/01/02 15:04", strings.Replace(s.Text(), "\n", " ", -1))
			case 2: //提出日
				submit := s.Text()
				if len(submit) > 0 {
					sTime, err := time.Parse("2006/01/02 15:04", strings.Replace(s.Text(), "\n", " ", -1))
					if err == nil {
						task.SubmitTime = &sTime
					} else {
						task.SubmitTime = nil
					}
				} else {
					task.SubmitTime = nil
				}
			case 3: //講義名
				a := s.Find("a")
				task.SubjectId, _ = strconv.Atoi(lectureIdRegExp.FindStringSubmatch(a.AttrOr("href", ""))[1])
				task.SubjectName = a.Text()
			case 4: //課題名
				a := s.Find("a")
				task.Title = a.Text()
				task.Id, _ = strconv.Atoi(kadaiIdRegExp.FindStringSubmatch(a.AttrOr("href", ""))[1])
			case 5: //教員名
				teachers := lecturerRegExp.FindAllString(s.Text(), -1)
				for i := range teachers {
					teachers[i] = strings.Trim(teachers[i], "()")
				}
				task.Lecturers = teachers
			case 6: //公開日
				task.OpenDateTime, _ = time.Parse("2006/01/02 15:04", strings.Replace(s.Text(), "\n", " ", -1))
			}
		})
		result = append(result, task)
	})

	return result, nil
}
