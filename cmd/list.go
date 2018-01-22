package cmd

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/wtks/gocwi/api"
	"os"
	"strconv"
	"strings"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "show subject list",
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
		list, err := api.GetLectureList()
		if err != nil {
			return err
		}

		for _, q := range list.Terms {
			fmt.Println(q.Name)
			table := tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ID", "Name", "Lecturer", "Period", "Room", "Open Tasks"})
			table.SetAutoMergeCells(true)
			for _, l := range q.Subjects {
				for i := range l.Periods {
					table.Append([]string{
						strconv.Itoa(l.Id),
						l.Name,
						strings.Join(l.Lecturers, ", "),
						l.Periods[i],
						l.Rooms[i],
						strconv.Itoa(l.OpenTaskCount),
					})
				}
			}
			table.Render()
			fmt.Println()
		}

		return nil
	},
	PostRunE: func(cmd *cobra.Command, args []string) error {
		return api.LogoutOcwi()
	},
}
