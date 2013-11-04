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
	p.brog.Debug("Getting all posts, got %d", len(p.sortedPosts))
	postCopy := make([]*post, len(p.sortedPosts))
	copy(postCopy, p.sortedPosts)
	p.brog.Debug("Getting all posts, returning %d", len(postCopy))
	return postCopy
}

func (p *postManager) GetAllPostsWithLanguage(lang string) []*post {
	var i int
	var j int
	p.mu.RLock()
	defer p.mu.RUnlock()
	for j=0;j<len(p.sortedPosts);j++ {
		if p.sortedPosts[j].Language == lang {
			i++
		}
	}
	postCopy := make([]*post, i)
	i = 0
	for j=0;j<len(p.sortedPosts);j++ {
		if p.sortedPosts[j].Language == lang {
			postCopy[i] = p.sortedPosts[j]
			i++
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

	p.brog.Watch("Deleting post by filename '%s', %d posts before", filename, len(p.posts))

	for _, post := range p.posts {
		p.brog.Debug("Checking if match '%s'", post.filename)
		if post.filename == filename {
			p.brog.Debug("Found post with ID '%s'", post.GetID())
			p.mu.RUnlock()
			p.DeletePost(post)
			p.sortPosts()
			return post, true
		}
	}
	p.mu.RUnlock()

	p.brog.Warn("Couldn't find post with filename '%s', not deleted", filename)
	return nil, false
}

func (p *postManager) SetPost(post *post) {
	p.mu.Lock()
	p.brog.Debug("SetPost with id %s, %d posts before", post.GetID(), len(p.posts))
	p.posts[post.GetID()] = post
	p.brog.Debug("SetPost, %d posts after", len(p.posts))
	p.mu.Unlock()

	p.sortPosts()
}

func (p *postManager) DeletePost(post *post) {
	p.mu.Lock()

	p.brog.Debug("Deleting post '%s', %d posts before", post.filename, len(p.posts))
	delete(p.posts, post.GetID())
	p.brog.Debug("Deleted '%s', %d posts left", post.filename, len(p.posts))
	p.mu.Unlock()

	p.sortPosts()
}

func (p *postManager) sortPosts() {
	p.brog.Debug("Sorting posts")
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

	p.brog.Debug("Sorted %d posts", len(postL.posts))
	p.mu.Lock()
	p.sortedPosts = postL.posts
	p.mu.Unlock()
}

func (p *postManager) Close() error {
	p.brog.Debug("PostManager closing now")
	p.die <- struct{}{}
	return p.watcher.Close()
}

func (p *postManager) watchForChanges(dirname string) error {
	go func() {
		for {
			select {
			case ev := <-p.watcher.Event:
				p.processPostEvent(ev)
				p.brog.Debug("PostManager processed event '%s'", ev.String())
			case err := <-p.watcher.Error:
				p.brog.Err("PostManager watching posts in '%s', %v", dirname, err)
			case <-p.die:
				return
			}
		}
	}()
	p.brog.Debug("PostManager watching for changes on '%s'", dirname)
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

	p.brog.Watch("Renamed post '%s'", ev.Name)

	post, ok := p.DeletePostWithFilename(ev.Name)

	if !ok {
		p.brog.Warn("Renamed unknown file '%s', ignoring", ev.Name)
		return
	}

	p.brog.Debug("Renamed post '%s': old filename '%s', deleting this copy, %d posts total",
		post.Title, ev.Name, len(p.posts))

	return
}

func (p *postManager) processPostDelete(ev *fsnotify.FileEvent) {
	post, ok := p.DeletePostWithFilename(ev.Name)

	if !ok {
		p.brog.Warn("Deleting unknown file '%s', ignoring", ev.Name)
		return
	}

	p.brog.Watch("Deleted post at '%s', %d posts left", post.Title, len(p.posts))
	return
}

func (p *postManager) processPostCreate(ev *fsnotify.FileEvent) {
	p.brog.Watch("New post at '%s'", ev.Name)
	err := p.loadFromFile(ev.Name)
	if err != nil {
		p.brog.Err("Error loading new post at '%s', %v", ev.Name, err)
	}
	p.brog.Watch("Assimilation completed. '%s' has become one with the brog.", ev.Name)
}

func (p *postManager) processPostModify(ev *fsnotify.FileEvent) {
	p.brog.Watch("Modified file '%s'", ev.Name)

	post, ok := p.DeletePostWithFilename(ev.Name)

	if !ok {
		fmt.Printf("Listing known posts.\n")
		for key := range p.posts {
			fmt.Printf("Knows of key %s\n", key)
		}
		p.brog.Warn("Post at '%s' was unknown", ev.Name)
	} else {
		p.brog.Watch("Reloading post titled '%s', %d posts before", post.Title, len(p.posts))
	}

	err := p.loadFromFile(ev.Name)
	if err != nil {
		p.brog.Err("Error loading new post at '%s', %v", ev.Name, err)
	}
}

func (p *postManager) loadFromFile(filename string) error {
	p.brog.Debug("Loading post from file '%s'", filename)
	post, err := newPostFromFile(filename)
	if err != nil {
		return fmt.Errorf("loading post from file '%s', %v", filename, err)
	}

	p.SetPost(post)

	p.brog.Debug("Loaded post '%s' from file '%s', %d posts total", post.Title, filename, len(p.posts))

	if post.Invisible {
		p.brog.Watch("'%s' is invisible.", post.Title)
	}
	return nil
}
