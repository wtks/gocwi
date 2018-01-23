package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wtks/gocwi/api"
	"gopkg.in/cheggaaa/pb.v1"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "download all attachments",
	PreRunE: func(cmd *cobra.Command, args []string) error {
		err := api.Login(getAccountId(), getPassword(), getMatrixRunes)
		if err != nil {
			return err
		}

		err = api.LoginOcwi()
		if err != nil {
			return err
		}

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := makeDirsIfNotExist(destDir); err != nil {
			return err
		}

		fmt.Print("fetching subject lists...")
		list, err := api.GetLectureList()
		if err != nil {
			return err
		}
		fmt.Println("done")

		for _, q := range list.Terms {
			for _, s := range q.Subjects {
				if !s.IsNotesAvailable {
					continue
				}

				fmt.Printf("fetching the notes of %s...", s.Name)
				notes, err := api.GetLectureNote(s.Id)
				if err != nil {
					return err
				}
				fmt.Println("done")

				for _, c := range notes.Classes {
					if len(c.Attachments) > 0 {
						fmt.Printf("#%s\n", c.Title)
						dir := path.Join(destDir, s.Name)
						if err := makeDirsIfNotExist(dir); err != nil {
							return err
						}

						dups := map[string]int{}
						for _, a := range c.Attachments {
							fmt.Printf(" + %s(%s) - %4d/%2d/%2d\n", a.Title, a.Type, a.Year, a.Month, a.Day)
							bar := pb.New(0)
							bar.Units = pb.U_BYTES_DEC
							name := ""
							if i, ok := dups[a.Title+"."+a.Ext]; ok {
								name = a.Title + "(" + strconv.Itoa(i) + ")." + a.Ext
								dups[a.Title+"."+a.Ext]++
							} else {
								dups[a.Title+"."+a.Ext] = 2
								name = a.Title + "." + a.Ext
							}
							dest := path.Join(dir, strings.Map(convertInValidRune, c.Title+" - "+name))
							if exists(dest) {
								fmt.Println("the file already exists. skip.")
								continue
							}

							if err := api.DownloadFile(a.Url, dest, bar); err != nil {
								return err
							}
						}
						fmt.Println()
					}
				}
			}
		}

		fmt.Println()
		fmt.Println("complete!")
		return nil
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		return api.LogoutOcwi()
	},
}

func makeDirsIfNotExist(path string) error {
	if info, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0777); err != nil {
			return err
		}
	} else if !info.IsDir() {
		return errors.New("the destination is not directory")
	}

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func convertInValidRune(rune rune) rune {
	if runtime.GOOS == "windows" {
		switch rune {
		case '\\':
			return '￥'
		case '/':
			return '／'
		case ':':
			return '：'
		case '*':
			return '＊'
		case '"':
			return '”'
		case '?':
			return '？'
		case '<':
			return '＜'
		case '>':
			return '＞'
		case '|':
			return '｜'
		default:
			return rune
		}
	} else {
		if rune == '/' {
			return '／'
		} else {
			return rune
		}
	}
}
