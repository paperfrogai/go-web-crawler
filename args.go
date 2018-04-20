
package main

import "flag"
//import "fmt"
import "regexp"
import "strings"
import "os"


type Config struct {
	SeedUrls       []string
	AllowUrls      []string
	SearchString   string
	NumConnections uint
	Stats          bool
}

/**
* Parse our command line arguments.
* @return {config} Our configuration info
 */
func ParseArgs() (retval Config) {

	retval = Config{[]string{}, []string{}, "", 1, false}

	hostnames := flag.String("seed-url",
		"http://www.27270.com/zt/baobao/3",
		"URL to start with.")
	allowUrls := flag.String("allow-urls",
		"", "Url base names to crawl. "+
			"If specified, this basically acts like a whitelist. "+
			"This may be a comma-delimited list. "+
			"Examples: http://cnn.com/, http://www.apple.com/store")
	flag.UintVar(&retval.NumConnections, "num-connections",
		1, "How many concurrent outbound connections?")
	flag.StringVar(&retval.SearchString, "search-string",
		"baby", "String to search for in alt and title tags of graphics")
	flag.BoolVar(&retval.Stats, "stats", false, "To print out stats once per second")

	h := flag.Bool("h", false, "To get this help")
	help := flag.Bool("help", false, "To get this help")
//	debug_level := flag.String("debug-level", "info", "Set the debug level")

	flag.Parse()

//	log.SetLevelString(*debug_level)
//	log.Error("Debug level: " + *debug_level)

	if *h || *help {
		flag.PrintDefaults()
		os.Exit(1)
	}

	retval.SeedUrls = SplitHostnames(*hostnames)
	retval.AllowUrls = SplitHostnames(*allowUrls)

	return (retval)

}

/**
* Take a comma-delimited string of hostnames and turn it into an array of URLs.
*
* @param {string} Input The comma-delimited string
*
* @return {[]string} Array of URLs
 */
func SplitHostnames(Input string) (retval []string) {

	Results := strings.Split(Input, ",")

	for _, value := range Results {

		if value != "" {
			pattern := "^http(s)?://"
			match, _ := regexp.MatchString(pattern, value)
			if !match {
				value = "http://" + value
			}

		}

		retval = append(retval, value)

	}

	return (retval)

} // End of SplitHostnames()