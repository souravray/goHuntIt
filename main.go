/*
* @Author: Sourav Ray <me@raysourav.com>
* @Date:   2014-08-15 01:24:47
* @Last Modified by:   souravray
* @Last Modified time: 2014-08-15 23:14:45
* @License: BEER-WARE
 */

package goHuntIt

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	s "strconv"
)

/*
   ProductHunt cilent
*/
type Client struct {
	APIKey      string
	APISecret   string
	APIVersion  string
	Host        string
	rootAddress string
	grantType   string
	BearerToken string
	ExpirySec   int
}

func NewClient() *Client {
	client := new(Client)
	client.APIVersion = "v1"
	client.rootAddress = "https://api.producthunt.com"
	client.grantType = "client_credentials"
	return client
}

func mergeParam(to url.Values, from url.Values) url.Values {
	if from != nil {
		for key, _ := range from {
			to.Set(key, from.Get(key))
		}
	}
	return to
}

func (self *Client) request(method string, endpoint string, getParams url.Values, postParams url.Values, headers map[string]string, data interface{}) error {
	var res *http.Response
	var req *http.Request
	var err error

	if method != "GET" && method != "POST" {
		return fmt.Errorf("Method not supported %s\n", method)
	}

	uri, err := url.ParseRequestURI(self.rootAddress)
	if err != nil {
		return err
	}

	finalEndpoint := self.APIVersion + "/" + endpoint + "/"
	uri.Path = finalEndpoint

	if getParams != nil {
		uri.RawQuery = getParams.Encode()
	}

	client := &http.Client{}
	urlStr := fmt.Sprintf("%v", uri)
	req, err = http.NewRequest(method, urlStr, bytes.NewBufferString(postParams.Encode()))
	if err != nil {
		return err
	}

	for key, value := range headers {
		req.Header.Add(key, value)
	}

	res, err = client.Do(req)
	if err != nil {
		return err
	}

	if res.Body != nil {
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if res.StatusCode == 200 {

			if err != nil {
				return err
			}

			err = json.Unmarshal(body, data)
			if err != nil {
				return err
			}

			return nil
		} else if res.StatusCode == 401 {
			// request bearer token renewal
			return fmt.Errorf("Auth-challenge")
		} else {
			return fmt.Errorf("ERROR: %s returned status %d  %s", finalEndpoint, res.StatusCode, body)
		}
	}

	return fmt.Errorf("Uncought error.\n")
}

func (self *Client) headers() map[string]string {
	headers := make(map[string]string)
	headers["Accept"] = "application/json"
	headers["Host"] = self.Host
	return headers
}

//Authentication

type authResponseBody struct {
	Access_token string `json: "access_token"`
	Token_type   string `json: "token_type"`
	Expires_in   int    `json: "expires_in"`
	Scope        string `json: "scope"`
}

/* API method: OAuth Client Only Authentication
 *  Doc: https://www.producthunt.com/v1/docs/oauth_client_only_authentication/oauthtoken__ask_for_client_level_token
 */
func (self *Client) ClientOnlyAuth() error {
	endpoint := "oauth/token"

	data := &authResponseBody{}

	headers := self.headers()

	postParams := url.Values{}
	postParams.Set("client_id", self.APIKey)
	postParams.Add("client_secret", self.APISecret)
	postParams.Add("grant_type", self.grantType)

	err := self.request("POST", endpoint, nil, postParams, headers, data)

	self.BearerToken = data.Access_token
	self.ExpirySec = data.Expires_in

	return err
}

// USER and POSTS

type UserResponseBody struct {
	Id         int
	Name       string
	Headline   string
	Created_at string
	Username   string
	// Image_url  []string  /*current jso schema has a non nocompatable key*/
}

type PostResponceBody struct {
	Id             int
	Name           string
	Tagline        string
	Created_at     string
	Day            string
	Comments_count int
	Votes_count    int
	Discussion_url string
	Redirect_url   string
	// Screenshot_url []string /*current jso schema has a non nocompatable key*/
	Maker_inside bool
	User         UserResponseBody
}

type PostsResponceBody struct {
	Posts []PostResponceBody
}

/* API method: Today's Posts
 *  Doc: https://www.producthunt.com/v1/docs/posts/postsindex__get_the_posts_of_today
 */
func (self *Client) PostsOfTheDay() error {
	endpoint := "posts"

	data := &PostsResponceBody{}

	headers := self.headers()
	headers["Authorization"] = "Bearer " + self.BearerToken

	err := self.request("GET", endpoint, nil, nil, headers, data)
	return err
}

/* API method: Historical Posts
 *  Doc: https://www.producthunt.com/v1/docs/posts/postsindex__get_the_posts_of_today
 */
func (self *Client) PostsOn(daysBack int) (*PostsResponceBody, error) {
	endpoint := "posts"

	data := &PostsResponceBody{}

	headers := self.headers()
	headers["Authorization"] = "Bearer " + self.BearerToken

	getParams := url.Values{}
	getParams.Set("days_ago", s.Itoa(daysBack))

	err := self.request("GET", endpoint, getParams, nil, headers, data)
	return data, err
}
