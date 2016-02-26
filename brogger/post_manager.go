package brogger

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/aybabtme/log"
	"github.com/howeyc/fsnotify"
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

			fullpath := filepath.Clean(p.path) +
				string(os.PathSeparator) +
				fileInfo.Name()

			err := p.loadFromFile(fullpath)
			if err != nil {
				log.Err(err).KV("file.name", fileInfo.Name()).Error("can't load post from file")
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

func (p *postManager) GetAllPostsWithLanguage(lang string) []*post {
	if lang == "" {
		return p.GetAllPosts()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	var postCopy []*post
	for _, val := range p.sortedPosts {
		if val.Language == lang {
			postCopy = append(postCopy, val)
		}
	}
	return postCopy
}

func (p *postManager) GetPost(key string) (*post, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	post, ok := p.posts[key]
	if ok && post.Invisible {
		return nil, false
	}
	return post, ok
}

func (p *postManager) DeletePostWithFilename(filename string) (*post, bool) {

	p.mu.RLock()

	ll := log.KV("file.name", filename)
	ll.KV("post.count", len(p.posts)).Info("deleting post by filename")

	for _, post := range p.posts {
		if post.filename == filename {
			p.mu.RUnlock()
			p.DeletePost(post)
			p.sortPosts()
			return post, true
		}
	}
	p.mu.RUnlock()

	ll.Error("couldn't find post to delete")
	return nil, false
}

func (p *postManager) SetPost(post *post) {
	p.mu.Lock()
	p.posts[post.GetID()] = post
	p.mu.Unlock()

	p.sortPosts()
}

func (p *postManager) DeletePost(post *post) {
	p.mu.Lock()

	delete(p.posts, post.GetID())
	p.mu.Unlock()

	p.sortPosts()
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
		ll := log.KV("dir.name", dirname)
		for {
			select {
			case ev := <-p.watcher.Event:
				p.processPostEvent(ev)
			case err := <-p.watcher.Error:
				ll.Err(err).Error("error watching posts")
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
	case p.brog.Config.PostFileExt:
	case ".md":
	case ".markdown":
	case ".mkd":
	default:
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

	log.KV("file.event", ev.String()).Error("unknown file event")
}

func (p *postManager) processPostRename(ev *fsnotify.FileEvent) {
	ll := log.KV("post.name", ev.Name)
	ll.Info("post name changed")

	_, ok := p.DeletePostWithFilename(ev.Name)

	if !ok {
		ll.Error("unknown post, ignoring the rename")
	}
}

func (p *postManager) processPostDelete(ev *fsnotify.FileEvent) {
	ll := log.KV("post.name", ev.Name)
	ll.Info("post name changed")

	post, ok := p.DeletePostWithFilename(ev.Name)

	if !ok {
		ll.Error("unknown post, ignoring the deletion")
		return
	}

	ll.
		KV("post.title", post.Title).
		KV("post.count", len(p.posts)).
		Info("deleted post")
	return
}

func (p *postManager) processPostCreate(ev *fsnotify.FileEvent) {
	ll := log.KV("post.name", ev.Name)
	ll.Info("new post detected")
	err := p.loadFromFile(ev.Name)
	if err != nil {
		ll.Err(err).Error("can't load new post")
		return
	}
	ll.Info("new post has been assimilated")
}

func (p *postManager) processPostModify(ev *fsnotify.FileEvent) {
	ll := log.KV("post.name", ev.Name)
	ll.Info("modified post detected")

	post, ok := p.DeletePostWithFilename(ev.Name)

	if ok {
		ll = ll.KV("post.title", post.Title)
		ll.Info("reloading post")
	}

	if err := p.loadFromFile(ev.Name); err != nil {
		ll.Err(err).Error("couldn't load modified post")
	}
}

func (p *postManager) loadFromFile(filename string) error {
	post, err := newPostFromFile(filename)
	if err != nil {
		return fmt.Errorf("loading post from file '%s', %v", filename, err)
	}

	p.SetPost(post)

	if post.Invisible {
		log.KV("post.title", post.Title).Info("post is invisible!")
	}
	return nil
}
