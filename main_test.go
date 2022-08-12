package main

import (
	"sort"
	"testing"
)

func checkIfFailed(t *testing.T, got interface{}, want interface{}) {
	if got != want {
		t.Errorf("got %q, wanted %q", got, want)
	}
}

//Test the stripRoot() function which should strip away the root component
func TestStripRoot(t *testing.T) {
	mockResponse := []byte(`{"posts":[{"author":"Rylee Paul","authorId":9,"id":1,"likes":960,"popularity":0.13,"reads":50361,"tags":["tech","health"]}]}`)
	ret, _ := stripRoot(mockResponse)
	got := string(ret)
	want := `[{"author":"Rylee Paul","authorId":9,"id":1,"likes":960,"popularity":0.13,"reads":50361,"tags":["tech","health"]}]`
	checkIfFailed(t, got, want)
}

//Test the hitHatchwaysAPI() function
func TestHitHatchwaysAPI(t *testing.T) {
	blogPosts := BlogPosts{}
	got := hitHatchwaysAPI(&blogPosts, "science")
	checkIfFailed(t, got, nil)
}

//Test the sortBlogPosts() function
func TestSortBlogPosts(t *testing.T) {
	blogPosts := BlogPosts{}
	hitHatchwaysAPI(&blogPosts, "science")

	sortBlogPosts(blogPosts, "id", "asc")
	got := sort.SliceIsSorted(blogPosts, func(i, j int) bool {
		return blogPosts[i].ID < blogPosts[j].ID
	})
	want := true
	checkIfFailed(t, got, want)

	sortBlogPosts(blogPosts, "likes", "desc")
	got = sort.SliceIsSorted(blogPosts, func(i, j int) bool {
		return blogPosts[i].Likes > blogPosts[j].Likes
	})
	checkIfFailed(t, got, want)

	sortBlogPosts(blogPosts, "popularity", "asc")
	got = sort.SliceIsSorted(blogPosts, func(i, j int) bool {
		return blogPosts[i].Popularity < blogPosts[j].Popularity
	})
	checkIfFailed(t, got, want)
}
