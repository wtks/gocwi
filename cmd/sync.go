package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/wtks/gocwi/api"
	"gopkg.in/cheggaaa/pb.v1"
	"path"
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
						header := fmt.Sprintf("# %s\n", c.Title)
						headerWritten := false
						dir := path.Join(destDir, s.Name)
						if err := makeDirsIfNotExist(dir); err != nil {
							return err
						}

						dups := map[string]int{}
						for _, a := range c.Attachments {
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
							dest := path.Join(dir, strings.Map(convertToValidRune, c.Title+" - "+name))
							if exists(dest) {
								if verbose {
									if !headerWritten {
										fmt.Print(header)
										headerWritten = true
									}
									fmt.Printf(" + %s(%s) - %4d/%2d/%2d\n", a.Title, a.Type, a.Year, a.Month, a.Day)
									fmt.Println("the file already exists. skip.")
								}
								continue
							} else {
								if !headerWritten {
									fmt.Print(header)
									headerWritten = true
								}
								fmt.Printf(" + %s(%s) - %4d/%2d/%2d\n", a.Title, a.Type, a.Year, a.Month, a.Day)
							}

							if err := api.DownloadFile(a.Url, dest, bar); err != nil {
								return err
							}
						}
						if headerWritten {
							fmt.Println()
						}
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
