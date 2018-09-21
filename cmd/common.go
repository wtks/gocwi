package cmd

import (
	"fmt"
	"os"
	"runtime"
)

func makeDirsIfNotExist(path string) error {
	if info, err := os.Stat(path); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0777); err != nil {
			return err
		}
	} else if !info.IsDir() {
		return fmt.Errorf("'%s' is not directory", path)
	}

	return nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func convertToValidRune(rune rune) rune {
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
