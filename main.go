package main

import (
	log "github.com/golang/glog"
	"os"
	"os/signal"
	"syscall"
)


func main() {
	config := ParseArgs()
	log.Infof("Config is: %s\n", config)
	log.Infof("SeedURL is: %s", config.SeedUrls)

	if len(config.AllowUrls) > 0 {
		log.Infof("Only allowing URLs  starting with: %s", config.AllowUrls)
	}

	go sigInt()

	/*
    interval := 1.0
	//interval := .1 // Debugging
	if config.Stats {
		go stats.StatDump(interval)
	}
*/

	NumConnections := config.NumConnections

	//
	// Start the crawler and seed it with our very first URL
	//
	UrlCrawlerIn, UrlCrawlerOut := NewUrlCrawler(uint(NumConnections), config.AllowUrls)

	//UrlCrawlerIn <- "http://localhost:8080/" // Debugging
	for _, value := range config.SeedUrls {
//		stats.IncrStat("urls_to_be_crawled")
		UrlCrawlerIn <- value
	}

	//
	// Create our HTML parser
	//
	HtmlBodyIn, ImageCrawlerIn := NewHtml(UrlCrawlerIn)

	//
	// Start up our image crawler
	//
	NewImageCrawler(config, ImageCrawlerIn, NumConnections)

	for {
		//
		// Read a result from our crawler
		//
		Res := <-UrlCrawlerOut

		if Res.Code != 200 {
			log.V(2).Infof("Skipping non-2xx response of %d on URL '%s'",
				Res.Code, Res.Url)
			continue
		}

		//
		// Pass it into the HTML parser.  It will in turn send any URLs
		// it finds into the URL Crawler and any images to the Image Crawler.
		//
		HtmlBodyIn <- []string{Res.Url, Res.Body, Res.ContentType}

	}


}

/**
* Wait for ctrl-c to happen, then exit!
 */
func sigInt() {
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT)
	<-ch
	log.Error("CTRL-C; exiting")
	os.Exit(0)
}