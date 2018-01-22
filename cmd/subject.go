package cmd

import (
	"errors"
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/ttacon/chalk"
	"github.com/wtks/gocwi/api"
	"os"
	"strconv"
)

var subjectCmd = &cobra.Command{
	Use:     "subject [subject ID]",
	Aliases: []string{"s"},
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("subject ID is required")
		}
		if _, err := strconv.Atoi(args[0]); err != nil {
			return errors.New("subject ID must be numbers")
		}
		return nil
	},
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
		id, _ := strconv.Atoi(args[0])

		note, err := api.GetLectureNote(id)
		if err != nil {
			return err
		}

		fmt.Printf("%s - %s\n", note.SubjectName, note.SubjectNameEn)
		if len(note.Classes) == 0 {
			fmt.Println("This subject doesn't disclose the class details.")
			return nil
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Content", "DateTime", "Room", "Type"})
		for _, c := range note.Classes {
			if c.IsCanceled {
				table.Append([]string{
					c.Title,
					c.Date,
					c.Room,
					chalk.Red.Color("休講"),
				})
			} else {
				var room string
				if c.IsRoomChanged {
					room = chalk.Red.Color(c.Room)
				} else {
					room = c.Room
				}
				table.Append([]string{
					c.Title,
					c.Date,
					room,
					c.Type,
				})
			}
		}
		table.Render()

		return nil
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		return api.LogoutOcwi()
	},
}
