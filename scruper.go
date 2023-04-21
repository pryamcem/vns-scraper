package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

// Login to VNS by login and password from config.
func Login(page *rod.Page, data Configuration) error {
	err := page.WaitLoad()
	if err != nil {
		return fmt.Errorf("Can't load page")
	}
	loginEntry, err := page.Timeout(time.Second).Element("#username")
	if err != nil {
		return fmt.Errorf("Can't find login entry: %v", err)
	}
	loginEntry.MustInput(data.Login)
	passwordEntry, err := page.Timeout(time.Second).Element("#password")
	if err != nil {
		return fmt.Errorf("Can't find login entry: %v", err)
	}
	passwordEntry.MustInput(data.Password)
	loginBtn, err := page.Timeout(time.Second).Element("#loginbtn")
	if err != nil {
		return fmt.Errorf("Can't find login entry: %v", err)
	}
	loginBtn.MustClick()
	return nil
}

func mustFindTestNum(page *rod.Page) int {
	page.MustWaitLoad()
	// find all <h2> elements on the page and loop through them
	h2Elements := page.MustElements("h2")
	for _, h2 := range h2Elements {
		text := h2.MustText()
		if strings.Contains(text, "Тест") {
			strs := strings.Fields(text)
			num, err := strconv.Atoi(strs[1])
			if err != nil {
				panic(err)
			}
			return num
		}
	}
	panic("Can't find test number")
}
