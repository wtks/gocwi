package api

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	urlTop        = "https://portal.titech.ac.jp"
	urlLoginTop   = "https://portal.nap.gsic.titech.ac.jp/GetAccess/Login?Template=userpass_key&AUTHMETHOD=UserPassword"
	urlLoginPost  = "https://portal.nap.gsic.titech.ac.jp/GetAccess/Login"
	urlPortalMenu = "https://portal.nap.gsic.titech.ac.jp/GetAccess/ResourceList"
)

func Login(id, pass string, matrixSolver func([3][]string) (rune, rune, rune)) error {
	res, err := client.Get(urlLoginTop)
	if err != nil {
		return err
	}
	b, _ := ioutil.ReadAll(res.Body)
	res.Body.Close()
	values, err := GetHiddenFormValues(string(b))
	if err != nil {
		return err
	}
	values.Add("usr_name", id)
	values.Add("usr_password", pass)
	values.Add("OK", "    OK    ")

	req, err := http.NewRequest("POST", urlLoginPost, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", urlLoginTop)

	res, err = client.Do(req)
	if err != nil {
		return err
	}
	loc := res.Request.URL.String()
	b, _ = ioutil.ReadAll(res.Body)
	res.Body.Close()
	if loc == urlTop { //account or Password is wrong
		return errors.New("account or password is wrong")
	} else if loc == urlLoginPost {
		return errors.New("something is wrong")
	}

	values, err = GetHiddenFormValues(string(b))
	if err != nil {
		return err
	}
	matrix, err := GetMatrixPositions(string(b))
	if err != nil {
		return err
	}
	m1, m2, m3 := matrixSolver(matrix)
	values.Add("message3", string(m1))
	values.Add("message4", string(m2))
	values.Add("message5", string(m3))
	values.Add("OK", "    OK    ")

	req, err = http.NewRequest("POST", urlLoginPost, strings.NewReader(values.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Referer", loc)

	res, err = client.Do(req)
	if err != nil {
		return err
	}
	loc = res.Request.URL.String()
	res.Body.Close()
	if loc == urlLoginPost { //matrix is wrong
		return errors.New("matrix is wrong")
	} else if loc != urlPortalMenu {
		return errors.New("something is wrong")
	} else {
		return nil //Success
	}
}
