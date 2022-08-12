package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const API_BLOGPOST_URL = "https://api.hatchways.io"

//The struct definition for the JSON response that comes from the Hatchways API
type BlogPosts []struct {
	ID         int      `json:"id"`
	Author     string   `json:"author"`
	AuthorID   int      `json:"authorId"`
	Likes      int      `json:"likes"`
	Popularity float64  `json:"popularity"`
	Reads      int      `json:"reads"`
	Tags       []string `json:"tags"`
}

//A simple cache struct and map that stores the time the API was called and the response
type BlogPostsCacheEntry struct {
	CalledAt int64
	Response []byte
}

var blogPostsCache = make(map[string]BlogPostsCacheEntry) //maps the tag that was queried to the above struct

const cacheValidSeconds = 5 //a cache entry is valid for 5 seconds or the external API should be hit again

//The acceptable fields for the sortBy and direction paramaters
var acceptedSortByParams = map[string]bool{
	"id":         true,
	"reads":      true,
	"likes":      true,
	"popularity": true,
}
var acceptedDirectionParams = map[string]bool{
	"desc": true,
	"asc":  true,
}

var wg sync.WaitGroup

//Sorts the blog posts
func sortBlogPosts(blogPosts BlogPosts, sortBy string, direction string) {
	sort.Slice(blogPosts, func(i, j int) bool {
		if direction == "asc" {
			switch sortBy {
			case "id":
				return blogPosts[i].ID < blogPosts[j].ID
			case "reads":
				return blogPosts[i].Reads < blogPosts[j].Reads
			case "likes":
				return blogPosts[i].Likes < blogPosts[j].Likes
			case "popularity":
				return blogPosts[i].Popularity < blogPosts[j].Popularity
			}
		} else if direction == "desc" {
			switch sortBy {
			case "id":
				return blogPosts[i].ID > blogPosts[j].ID
			case "reads":
				return blogPosts[i].Reads > blogPosts[j].Reads
			case "likes":
				return blogPosts[i].Likes > blogPosts[j].Likes
			case "popularity":
				return blogPosts[i].Popularity > blogPosts[j].Popularity
			}
		}
		return true
	})
}

//strips the root
func stripRoot(html []byte) ([]byte, error) {
	m := make(map[string]interface{})
	err := json.Unmarshal(html, &m)
	strippedRoot, err := json.Marshal(m["posts"])

	return strippedRoot, err
}

//Function that hits the hatchways api
func hitHatchwaysAPI(blogPosts *BlogPosts, tag string) error {
	defer wg.Done()

	//check if the cache already contains the result we are looking for
	if cachedResult, exist := blogPostsCache[tag]; exist {
		if cachedResult.CalledAt+cacheValidSeconds > time.Now().Unix() {
			fmt.Println("Cache hit!")
			err := json.Unmarshal(cachedResult.Response, &blogPosts)
			return err
		}
	}

	//make the request, defer closing the reponse body, read the response
	res, err := http.Get(API_BLOGPOST_URL + "/assessment/blog/posts?tag=" + tag)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	html, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	//Unmarshalls the response from the hatchways API into the struct definition
	strippedRoot, err := stripRoot(html)
	errU := json.Unmarshal(strippedRoot, blogPosts)

	//cache the result
	blogPostsCache[tag] = BlogPostsCacheEntry{time.Now().Unix(), strippedRoot}

	return errU
}

//This API listens on 0.0.0.0:8080 (windows: localhost:8080)
func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	//Route 1
	router.GET("/api/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})

	//Route 2
	router.GET("/api/posts", func(c *gin.Context) {
		tags := c.Query("tags")
		sortBy := c.Query("sortBy")
		direction := c.Query("direction")

		//Check if tags parameter is not present
		if tags == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tags paramater is required",
			})
			return
		}
		splittedTags := strings.Split(tags, ",")

		//set the default values for sortBy and direction
		if sortBy == "" {
			sortBy = "id"
		}
		if direction == "" {
			direction = "asc"
		}

		//Check if sortBy and direction parameter are invalid values
		if !acceptedSortByParams[sortBy] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "sortBy parameter is invalid",
			})
			return
		}
		if !acceptedDirectionParams[direction] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "direction parameter is invalid",
			})
			return
		}

		//Hit the hatchways API concurently
		blogPosts := BlogPosts{}
		for _, tag := range splittedTags {
			wg.Add(1)
			go hitHatchwaysAPI(&blogPosts, tag)
		}

		wg.Wait()
		fmt.Println(fmt.Sprint(len(blogPosts)) + " blogposts retrieved")

		//sort the blog posts and sent a response
		sortBlogPosts(blogPosts, sortBy, direction)
		c.JSON(http.StatusOK, blogPosts)

	})

	router.Run()
}
