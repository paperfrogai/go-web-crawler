package main

//import "fmt" // Debugging
import "regexp"
import "strings"
import log "github.com/golang/glog"

//
// Keep track of if we crawled hosts with specific URLs
//
var hostsCrawled map[string]map[string]bool

//
// Our allowed URLs to crawl. If empty, all URLs are crawled.
//
var allowedUrls []string

/**
* Spin up 1 or more goroutines to do crawling.
*
* @param {int} num_instances
* @returm {chan string, chan Response} Our channel to read URLs from,
*	our channel to write responses to.
 */
func NewUrlCrawler(NumInstances uint, AllowedUrls []string) (in chan string, out chan Response) {

	hostsCrawled = make(map[string]map[string]bool)
	allowedUrls = AllowedUrls

	//
	// I haven't yet decided if I want a buffer for this
	//
	//InBufferSize := 1000
	InBufferSize := 0

	//
	// If we don't have a large output buffer, using multiple seed URLs
	// will cause blocking to happen (ooops!)
	//
	OutBufferSize := 1000
	in = make(chan string, InBufferSize)
	out = make(chan Response, OutBufferSize)

	for i := uint(0); i < NumInstances; i++ {
		log.Infof("Spun up crawler instance #%d\n", (i + 1))
		go crawlUrls(in, out)
	}

	return in, out

} // End of NewUrlCrawler()

/**
* This is run as a goroutine which is responsible for doing the crawling and
* returning the results.
*
* @param {chan string} in Our channel to read URLs to crawl from
* @param {chan Response} out Responses will be written on this channel
*
* @return {Response} A response consisting of our code and body
 */
func crawlUrls(in chan string, out chan Response) {

	for {

		log.V(2).Infoln("About to ingest a URL...")
//		stats.IncrStat("go_url_crawler_waiting")
		url := <-in
//		stats.DecrStat("go_url_crawler_waiting")
//		stats.DecrStat("urls_to_be_crawled")

		if !isUrlAllowed(url) {
			log.V(2).Infof("URL '%s' is not allowed!\n", url)
			continue
		}

		url = filterUrl(url)

		if urlBeenHere(url) {
			log.V(2).Infof("We've already been to '%s', skipping!\n", url)
			continue
		}

		if !sanityCheck(url) {
			//
			// In the future, I might make the in channel take a data
			// structure which includes the referrer so I can dig
			// into bad URLs. With a backhoe.
			//
			log.Warningf("URL '%s' fails sanity check, skipping!", url)
			continue
		}

		log.Infof("About to crawl '%s'...", url)
		out <- httpGet(url)
		log.Infof("Done crawling '%s'!", url)

	}

} // End of crawl()

/**
* Filter meaningless things out of URLs. Like hashmarks.
*
* @param {string} url The URL
*
* @return {string} The filtered URL
 */
func filterUrl(url string) string {

	//
	// First, nuke hashmarks (thanks, Apple!)
	//
	regex, _ := regexp.Compile("([^#]+)#")
	results := regex.FindStringSubmatch(url)
	if len(results) >= 2 {
		url = results[1]
	}

	//
	// Replace groups of 2 or more slashes with a single slash (thanks, log4j!)
	//
	regex, _ = regexp.Compile("[^:](/[/]+)")
	for {
		results = regex.FindStringSubmatch(url)
		if len(results) < 2 {
			break
		}

		Dir := results[1]
		//url = regex.ReplaceAllString(url, "/")
		url = strings.Replace(url, Dir, "/", -1)

	}

	//
	// Fix broken methods (thanks, Flickr!)
	//
	regex, _ = regexp.Compile("^(http)(s)?(:/)[^/]")
	results = regex.FindStringSubmatch(url)
	if len(results) > 0 {
		BrokenMethod := results[1] + results[2] + results[3]
		url = strings.Replace(url, BrokenMethod, BrokenMethod+"/", 1)
	}

	//
	// Now, remove references to parent directories, because that's just
	// ASKING for path loops. (thanks, Apple!)
	//
	// Do this by looping as long as we have ".." present.
	//
	regex, _ = regexp.Compile("([^/]+/\\.\\./)")
	for {
		results = regex.FindStringSubmatch(url)
		if len(results) < 2 {
			break
		}

		Dir := results[1]
		url = strings.Replace(url, Dir, "", -1)

	}

	//
	// Replace paths of single dots
	//
	regex, _ = regexp.Compile("/\\./")
	url = regex.ReplaceAllString(url, "/")

	return (url)

} // End of filterUrl()

/**
* Have we already been to this URL?
*
* @param {string} url The URL we want to crawl
*
* @return {bool} True if we've crawled this URL before, false if we have not.
 */
func urlBeenHere(url string) (retval bool) {

	retval = true

	//
	// Grab our URL parts
	//
	results := getUrlParts(url)
	if len(results) < 5 {
		//
		// TODO: Use data structure and print referrer here!
		//
		log.Warningf("urlBeenHere(): Unable to parse URL: '%s'", url)
		return (true)
	}
	Host := results[1]
	Uri := results[4]

	//
	// Create our host entry if we don't already have it.
	//
	if _, ok := hostsCrawled[Host]; !ok {
		hostsCrawled[Host] = make(map[string]bool)
	}

	//
	// If this is our first time here, cool. Otherwise, skip.
	//
	if _, ok := hostsCrawled[Host][Uri]; !ok {
		hostsCrawled[Host][Uri] = true
		retval = false
	}

	return retval

} // End of urlBeenHere()

/**
* Split up our URL into its component parts
 */
func getUrlParts(url string) (retval []string) {

	regex, _ := regexp.Compile("((https?://)([^/]+))(.*)")
	retval = regex.FindStringSubmatch(url)

	if len(retval) < 5 {
		log.Warningf("getUrlParts(): Unable to parse URL: '%s'", url)
	}

	return (retval)

} // End of getUrlParts()

/**
* Check to see if this URL is sane.
*
* @return {bool} True if the URL looks okay, false otherwise.
 */
func sanityCheck(url string) (retval bool) {

	retval = true

	regex, _ := regexp.Compile(" ")
	result := regex.FindString(url)

	if result != "" {
		retval = false
	}

	return (retval)

} // End of sanityCheck()

/**
* Is this URL on our allowed list?
*
* @param {string} The URL to check
*
* @return {bool} If allowed, true. Otherwise, false.
 */
func isUrlAllowed(url string) (retval bool) {

	if len(allowedUrls) == 0 {
		return true
	}

	//
	// Loop through our URLs and return true on the first match
	//
	for _, value := range allowedUrls {
		pattern := "^" + value
		match, _ := regexp.MatchString(pattern, url)
		if match {
			return true
		}
	}

	//
	// If we got here, no match was found. Return false.
	//
	return false

} // End of isUrlAllowed()