package natsproxy

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/nats-io/nats"
)

var (
	pathrgxp = regexp.MustCompile(":[A-z,0-9,$,-,_,.,+,!,*,',(,),\\,]{1,}")
)

// URLToNats builds the channel name
// from an URL and Method of http.Request
func URLToNats(method string, urlPath string) string {
	subURL := strings.Replace(urlPath, "/", ".", -1)
	subURL = fmt.Sprintf("%s:%s", method, subURL)
	return subURL
}

// SubscribeURLToNats buils the subscription
// channel name with placeholders (started with ":").
// The placeholders are than used to obtain path variables
func SubscribeURLToNats(method string, urlPath string) string {
	subURL := pathrgxp.ReplaceAllString(urlPath, "*")
	// subURL = lastpathrgxp.ReplaceAllString(subURL, ".*")
	subURL = strings.Replace(subURL, "/", ".", -1)
	subURL = fmt.Sprintf("%s:%s", method, subURL)
	return subURL
}

// copy the values into protocol buffer
// struct
func copyMap(values map[string][]string) map[string]*Values {
	headerMap := make(map[string]*Values, 0)
	for k, v := range values {
		headerMap[k] = &Values{
			v,
		}
	}
	return headerMap
}

func testConnection(conn *nats.Conn) error {
	if conn == nil {
		return fmt.Errorf("natsproxy: Connection cannot be nil")
	}
	if conn.Status() != nats.CONNECTED {
		return ErrNatsClientNotConnected
	}
	return nil
}
