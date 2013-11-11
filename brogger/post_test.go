package brogger

import (
	"os"
	"testing"
	"time"
)

func TestNewPostFromFile(t *testing.T) {
	/* Perhaps this should be a file in a test/ directory */
	os.Chdir("base")
	defer os.Chdir("..")
	post, err := newPostFromFile("posts" + string(os.PathSeparator) + "sample.md")
	if err != nil {
		t.Error("Can't read post because of error: %v", err)
	}
	if post.Title != "My First Post" {
		t.Error("newPostFromFile doesn't read title properly. Got", post.Title)
	}
	if post.Author != "Antoine Grondin" {
		t.Error("newPostFromFile doesn't read author properly. Got", post.Author)
	}
	if date, _ := time.Parse("2006-01-02T15:04:05.000Z", "2013-10-30T23:59:59.000Z"); post.Date != date {
		t.Error("newPostFromFile doesn't read date properly. Got", post.Date.Format("2006-01-02T15:04:05.000Z"))
	}
	if !post.Invisible {
		t.Error("Post which should be invisible parsed as not invisible")
	}
	if post.Abstract != "My first post using Brog" {
		t.Error("newPostFromFile doesn't read abstract properly. Got", post.Abstract)
	}
	if post.Language != "en" {
		t.Error("newPostFromFile doesn't read language properly. Got", post.Language)
	}
}

func TestGetAllPosts(t *testing.T) {
	pmgr, err := startPostManager(SetUpDefaultBrog(), "base"+string(os.PathSeparator)+DefaultPostPath)
	if err != nil {
		t.Error("Error encountered starting post manager: %v", err)
	}
	posts := pmgr.GetAllPosts()
	if len(posts) > 0 {
		t.Error("Invisible posts returned by Get All Posts")
	}
}
