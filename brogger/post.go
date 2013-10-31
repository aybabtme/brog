package brogger

import (
	"encoding/json"
	"fmt"
	"github.com/howeyc/fsnotify"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Post struct {
	filename string
	Title    string    `json:"title"`
	Date     time.Time `json:"date"`
	Author   string    `json:"author"`
	Abstract string    `json:"abstract"`
	Content  string
}

func NewPostFromFile(filename string) (*Post, error) {

	post := Post{filename: filename}

	postFile, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("opening file '%s', %v", filename, err)
	}
	defer postFile.Close()

	dec := json.NewDecoder(postFile)
	if err = dec.Decode(&post); err != nil {
		return nil, fmt.Errorf("reading json header of post '%s', %v", filename, err)
	}

	// Read content after that JSON header.  json.Decoder buffers some data, so
	// we need to ask for a buffered reader on it's current position.
	markdownContent, err := ioutil.ReadAll(dec.Buffered())
	if err != nil {
		return nil, fmt.Errorf("reading markdown content of post '%s', %v", post.Title, err)
	}

	htmlContent := blackfriday.MarkdownCommon(markdownContent)
	post.Content = string(htmlContent)

	return &post, nil
}

func WriteSamplePost(filename string) error {

	const samplePost = `{
    "title":"My First Post",
    "author":"Antoine Grondin",
    "date":"2013-10-30T23:59:59.000Z",
    "abstract":"My first post using Brog"
}
# Hello!!
This is my first Brog post.  I really like broging and feeling like I'm finally one of those Broggers.  At last, I'm part of a community!

` + "```" + `
Fenced code block?
` + "```" + `

Maybe!  __Who knows!!!__  Hopefully [this will be a link][1].

> Don't click it!!

_Shhhh_.

## Reasons why Antoine is great

* He has a nice beard.
* He has a nice way of mispeaking English.
* He prefers Star Trek to Star Wars.

[1]: en.wikipedia.org/wiki/Borg_(Star_Trek)
`

	return ioutil.WriteFile(filename, []byte(samplePost), 0640)
}

type PostManager struct {
	brog    *Brog
	path    string
	posts   map[string]*Post
	watcher *fsnotify.Watcher
}

func StartPostManager(brog *Brog, filepath string) (*PostManager, error) {

	os.MkdirAll(filepath, 0740)
	fileInfos, err := ioutil.ReadDir(filepath)
	if err != nil {
		return nil, fmt.Errorf("listing directory '%s', %v", filepath, err)
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("getting post watcher, %v", err)
	}

	postMngr := &PostManager{
		brog:    brog,
		path:    filepath,
		posts:   make(map[string]*Post),
		watcher: watcher,
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {

			fullpath := path.Clean(filepath) +
				string(os.PathSeparator) +
				fileInfo.Name()

			err := postMngr.loadFromFile(fullpath)
			if err != nil {
				brog.Warn("Loading post from file '%s' failed, %v",
					fileInfo.Name(), err)
			}
		}
	}

	postMngr.watchForChanges(filepath)

	return postMngr, nil
}

func (p *PostManager) close() error {
	return p.watcher.Close()
}

func (p *PostManager) watchForChanges(dirname string) error {

	go func() {
		for {
			select {
			case ev := <-p.watcher.Event:
				p.processPostEvent(ev)
			case err := <-p.watcher.Error:
				p.brog.Err("watching posts in '%s', %v", dirname, err)
			}
		}
	}()

	return p.watcher.Watch(dirname)

}

func (p *PostManager) processPostEvent(ev *fsnotify.FileEvent) {

	ext := filepath.Ext(ev.Name)
	switch ext {
	case ".md":
	case ".markdown":
	default:
		p.brog.Ok("Ignoring files in '%s': %s", ext, ev.Name)
		return
	}

	if ev.IsCreate() {
		p.processPostCreate(ev)
		return
	}

	if ev.IsModify() {
		p.processPostModify(ev)
		return
	}

	if ev.IsRename() {
		p.processPostRename(ev)
		return
	}

	if ev.IsDelete() {
		p.processPostDelete(ev)
		return
	}

	p.brog.Err("FileEvent '%s' is not recognized", ev.String())
}

func (p *PostManager) processPostRename(ev *fsnotify.FileEvent) {
	post, ok := p.posts[ev.Name]
	if !ok {
		p.brog.Warn("Renamed unknown file '%s', ignoring", ev.Name)
		return
	}
	delete(p.posts, ev.Name)
	p.brog.Ok("Post '%s': old filename '%s', deleting, %d posts total",
		post.Title, ev.Name, len(p.posts))

	return
}

func (p *PostManager) processPostDelete(ev *fsnotify.FileEvent) {
	post, ok := p.posts[ev.Name]
	if !ok {
		p.brog.Warn("Deleting unknown file '%s', ignoring", ev.Name)
		return
	}
	delete(p.posts, ev.Name)
	p.brog.Ok("Removing post '%s', %d posts left", post.Title, len(p.posts))
	return
}

func (p *PostManager) processPostCreate(ev *fsnotify.FileEvent) {
	p.brog.Ok("New file '%s'", ev.Name)
	err := p.loadFromFile(ev.Name)
	if err != nil {
		p.brog.Err("Error loading new post at '%s', %v", ev.Name, err)
	}
}

func (p *PostManager) processPostModify(ev *fsnotify.FileEvent) {
	p.brog.Ok("Modified file '%s'", ev.Name)

	post, ok := p.posts[ev.Name]
	if !ok {
		p.brog.Warn("File '%s' was unknown", ev.Name)
	}

	err := p.loadFromFile(ev.Name)
	if err != nil {
		p.brog.Err("Error loading new post at '%s', %v", ev.Name, err)
		if ok {
			delete(p.posts, ev.Name)
			p.brog.Warn("Removing related post '%s', %d posts left",
				post.Title, len(p.posts))

		}
	}
}

func (t *PostManager) loadFromFile(filename string) error {
	post, err := NewPostFromFile(filename)
	if err != nil {
		return fmt.Errorf("loading post from file '%s', %v", filename, err)
	}

	t.posts[filename] = post
	t.brog.Ok("Loaded post '%s' from file '%s', %d posts total", post.Title, filename, len(t.posts))

	return nil
}
