package brogger

import (
	"fmt"
	"github.com/howeyc/fsnotify"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

type postManager struct {
	brog *Brog  // Reference to the Brog app for logging purpose
	path string // Path on which the manager watch for post changes

	watcher *fsnotify.Watcher // Listens on `path`
	die     chan struct{}     // To kill the watcher goroutine

	mu          sync.RWMutex     // Locks the `posts` and `sortedPosts`
	posts       map[string]*post // All the posts, accessed by filename
	sortedPosts []*post          // All the posts in most recent order
}

func startPostManager(brog *Brog, filepath string) (*postManager, error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("getting post watcher, %v", err)
	}

	postMngr := &postManager{
		mu:          sync.RWMutex{},
		brog:        brog,
		path:        filepath,
		posts:       make(map[string]*post),
		sortedPosts: []*post{},
		watcher:     watcher,
		die:         make(chan struct{}),
	}

	err = postMngr.loadAllPosts()
	if err != nil {
		return nil, fmt.Errorf("while loading all posts, %v", err)
	}

	postMngr.sortPosts()
	if err := postMngr.watchForChanges(filepath); err != nil {
		return nil, fmt.Errorf("starting watch for changes on '%s', %v", filepath, err)
	}

	return postMngr, nil
}

func (p *postManager) loadAllPosts() error {
	fileInfos, err := ioutil.ReadDir(p.path)
	if err != nil {
		return fmt.Errorf("listing directory '%s', %v", p.path, err)
	}

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() {

			fullpath := path.Clean(p.path) +
				string(os.PathSeparator) +
				fileInfo.Name()

			err := p.loadFromFile(fullpath)
			if err != nil {
				p.brog.Warn("Loading post from file '%s' failed, %v",
					fileInfo.Name(), err)
			}
		}
	}
	return nil
}

func (p *postManager) GetAllPosts() []*post {
	p.mu.RLock()
	defer p.mu.RUnlock()
	postCopy := make([]*post, len(p.sortedPosts))
	copy(postCopy, p.sortedPosts)
	return postCopy
}

func (p *postManager) GetPost(key string) (*post, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	post, ok := p.posts[key]
	if post.Invisible {
		return nil, false
	}
	return post, ok
}

func (p *postManager) DeletePostWithFilename(filename string) (*post, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, post := range p.posts {
		if post.filename == filename {
			delete(p.posts, post.GetID())
			return post, true
		}
	}
	return nil, false
}

func (p *postManager) SetPost(post *post) {
	p.mu.Lock()
	p.posts[post.GetID()] = post
	p.mu.Unlock()

	p.sortPosts()
}

func (p *postManager) DeletePost(key string) (*post, bool) {

	p.mu.Lock()
	post, ok := p.posts[key]
	if !ok {
		p.mu.Unlock()
		return nil, ok
	}
	delete(p.posts, key)
	p.mu.Unlock()

	p.sortPosts()

	return post, ok
}

func (p *postManager) sortPosts() {

	var postL postList

	p.mu.RLock()
	for _, val := range p.posts {
		if val.Invisible {
			continue
		}
		postL.posts = append(postL.posts, val)
	}
	p.mu.RUnlock()

	sort.Sort(postL)

	p.mu.Lock()
	p.sortedPosts = postL.posts
	p.mu.Unlock()
}

func (p *postManager) Close() error {
	p.die <- struct{}{}
	return p.watcher.Close()
}

func (p *postManager) watchForChanges(dirname string) error {

	go func() {
		for {
			select {
			case ev := <-p.watcher.Event:
				p.processPostEvent(ev)
			case err := <-p.watcher.Error:
				p.brog.Err("watching posts in '%s', %v", dirname, err)
			case <-p.die:
				return
			}
		}
	}()

	return p.watcher.Watch(dirname)

}

func (p *postManager) processPostEvent(ev *fsnotify.FileEvent) {

	ext := strings.ToLower(filepath.Ext(ev.Name))
	switch ext {
	case ".md":
	case ".markdown":
	case ".mkd":
	default:
		p.brog.Debug("Posts ignore files in '%s': %s", ext, ev.Name)
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

func (p *postManager) processPostRename(ev *fsnotify.FileEvent) {

	post, ok := p.DeletePost(ev.Name)

	if !ok {
		p.brog.Warn("Renamed unknown file '%s', ignoring", ev.Name)
		return
	}

	p.brog.Debug("Post '%s': old filename '%s', deleting, %d posts total",
		post.Title, ev.Name, len(p.posts))

	return
}

func (p *postManager) processPostDelete(ev *fsnotify.FileEvent) {

	post, ok := p.DeletePost(ev.Name)

	if !ok {
		p.brog.Warn("Deleting unknown file '%s', ignoring", ev.Name)
		return
	}

	p.brog.Debug("Removing post '%s', %d posts left", post.Title, len(p.posts))
	return
}

func (p *postManager) processPostCreate(ev *fsnotify.FileEvent) {
	p.brog.Debug("New file '%s'", ev.Name)
	err := p.loadFromFile(ev.Name)
	if err != nil {
		p.brog.Err("Error loading new post at '%s', %v", ev.Name, err)
	}
}

func (p *postManager) processPostModify(ev *fsnotify.FileEvent) {
	p.brog.Debug("Modified file '%s'", ev.Name)

	post, ok := p.DeletePostWithFilename(ev.Name)

	if !ok {
		p.brog.Warn("File '%s' was unknown", ev.Name)
	}

	p.brog.Warn("Removing related post '%s', %d posts left", post.Title, len(p.posts))

	err := p.loadFromFile(ev.Name)
	if err != nil {
		p.brog.Err("Error loading new post at '%s', %v", ev.Name, err)
	}
}

func (p *postManager) loadFromFile(filename string) error {
	post, err := newPostFromFile(filename)
	if err != nil {
		return fmt.Errorf("loading post from file '%s', %v", filename, err)
	}

	p.SetPost(post)

	p.brog.Debug("Loaded post '%s' from file '%s', %d posts total", post.Title, filename, len(p.posts))

	return nil
}
