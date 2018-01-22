package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/wtks/gocwi/api"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"syscall"
)

var RootCmd = &cobra.Command{
	Use:   "gocwi [command]",
	Short: "gocwi is a cui tool for OCWi",
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "login test",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		err := api.Login(getAccountId(), getPassword(), getMatrixRunes)
		if err != nil {
			log.Fatal(err)
		} else {
			fmt.Println("OK")
		}
	},
}

var (
	accountId string
	password  string
	destDir   string
	mSeq      string
)

func getAccountId() string {
	for len(accountId) == 0 {
		fmt.Print("your account> ")
		fmt.Scanf("%s", &accountId)
	}

	return accountId
}

func getPassword() string {
	for len(password) == 0 {
		fmt.Print("your password> ")
		b, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			fmt.Scanf("%s", &password)
		} else {
			password = string(b)
			fmt.Println()
		}
	}

	return password
}

func getMatrixRunes(matrix [3][]string) (m1 rune, m2 rune, m3 rune) {
	if len(mSeq) == 70 {
		m1 = rune(mSeq[int(([]rune(matrix[0][0])[0]-'A')*7+[]rune(matrix[0][1])[0]-'1')])
		m2 = rune(mSeq[int(([]rune(matrix[1][0])[0]-'A')*7+[]rune(matrix[1][1])[0]-'1')])
		m3 = rune(mSeq[int(([]rune(matrix[2][0])[0]-'A')*7+[]rune(matrix[2][1])[0]-'1')])
	} else {
		fmt.Printf("Matrix %s%s>", matrix[0][0], matrix[0][1])
		fmt.Scanf("%c\n", &m1)
		fmt.Printf("Matrix %s%s>", matrix[1][0], matrix[1][1])
		fmt.Scanf("%c\n", &m2)
		fmt.Printf("Matrix %s%s>", matrix[2][0], matrix[2][1])
		fmt.Scanf("%c\n", &m3)
	}
	return
}

func init() {
	home, err := homedir.Expand("~/gocwi")
	if err != nil {
		log.Fatal(err)
	}
	cobra.OnInitialize()
	RootCmd.AddCommand(testCmd)
	RootCmd.AddCommand(listCmd)
	RootCmd.AddCommand(subjectCmd)
	RootCmd.AddCommand(syncCmd)
	RootCmd.PersistentFlags().StringVarP(&accountId, "account", "a", "", "your account id")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "p", "", "**NOT RECOMMENDED** your password")
	RootCmd.PersistentFlags().StringVarP(&destDir, "dest", "d", home, "downloaded files destination (default: '~/gocwi')")
	RootCmd.PersistentFlags().StringVarP(&mSeq, "matrix", "m", "", "**NOT RECOMMENDED** your matrix character's sequence (A1-A7B1-B7...)")
}

func Execute() {
	RootCmd.Execute()
}
