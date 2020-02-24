package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/tebeka/selenium"
)

func titleIsProfileName(webDriver selenium.WebDriver) string {
	title, err := webDriver.Title()
	if err != nil {
		panic(err)
	}

	title = title[0:strings.LastIndex(title, " |")]
	profile := StoredPath + "\\" + title
	if _, err := os.Stat(profile); os.IsNotExist(err) {
		err = os.MkdirAll(profile, 0755)
		if err != nil {
			panic(err)
		}
	}
	return profile
}

func extractUsedParameter(href string) string {
	u, err := url.Parse(href)
	if err != nil {
		panic(err)
	}

	scratch := strings.Split(u.RawQuery, "&")
	var imageStr string
	for _, link := range scratch {
		if strings.HasPrefix(link, "img=") {
			imageStr = link[strings.LastIndex(link, "img=")+len("img=") : len(link)]
			return imageStr
		}
	}
	return ""
}

func main() {
	opts := []selenium.ServiceOption{
		// Enable fake XWindow session.
		// selenium.StartFrameBuffer(),
		// selenium.Output(os.Stderr), // Output debug information to STDERR
	}

	service, err := selenium.NewChromeDriverService(SeleniumPath, Port, opts...)
	if err != nil {
		panic(err)
	}
	defer service.Stop()

	caps := selenium.Capabilities{"browserName": "chrome"}
	webDriver, err := selenium.NewRemote(caps, fmt.Sprintf("http://127.0.0.1:%d/wd/hub", Port))
	if err != nil {
		panic(err)
	}
	defer webDriver.Quit()

	webDriver.Get("http://nlegs.com/girls/2020/02/19/13664.html")

	profile := titleIsProfileName(webDriver)
	elements, err := webDriver.FindElements(selenium.ByCSSSelector, ".col-md-12.col-lg-12.panel.panel-default .panel-body a")
	if err != nil {
		panic(err)
	}

	for i, elt := range elements {
		href, err := elt.GetAttribute("href")
		if err != nil {
			err.Error()
		}
		imgID := extractUsedParameter(href)
		imgURL := "http://nlegs.com/images/" + imgID + ".jpg"

		fmt.Println(imgURL)
		client := &http.Client{}
		request, err := http.NewRequest("GET", imgURL, nil)
		if err != nil {
			log.Fatalln(err)
		}
		request.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/66.0.3359.181 Safari/537.36")
		request.Header.Set("Referer", imgURL)
		response, err := client.Do(request)
		if err != nil {
			log.Fatalln(err)
		}
		defer response.Body.Close()

		shell := strconv.Itoa(i+1) + ".jpg"
		s := []string{profile, shell}
		imgPath := strings.Join(s, "\\")
		file, err := os.Create(imgPath)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		_, err = io.Copy(file, response.Body)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Success!")
}
