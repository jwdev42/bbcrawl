package download

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

type threadcounter struct {
	m       *sync.Mutex
	counter int
	max     int
}

func (r *threadcounter) inc() bool {
	r.m.Lock()
	defer r.m.Unlock()
	if r.counter >= r.max {
		return false
	}
	r.counter++
	return true
}

func (r *threadcounter) dec() {
	r.m.Lock()
	defer r.m.Unlock()
	if r.counter < 1 {
		panic("Attempted to decrease a thread counter < 1")
	}
	r.counter--
}

func (r *threadcounter) zero() bool {
	r.m.Lock()
	defer r.m.Unlock()
	if r.counter == 0 {
		return true
	}
	return false
}

type Download struct {
	Client        *http.Client
	Addr          *url.URL
	File          string
	AllowOverride bool
	Err           error
}

type DownloadDispatcher struct {
	max      int
	counter  *threadcounter
	resc     chan *Download //yields the state of finished download routines
	finished []*Download    //collector will receive messages from resc and write it into finished
}

func NewDownloadDispatcher(downloads int) *DownloadDispatcher {
	dd := DownloadDispatcher{
		max:      downloads,
		counter:  &threadcounter{m: new(sync.Mutex), max: downloads},
		resc:     make(chan *Download, downloads),
		finished: make([]*Download, 0, downloads*2),
	}
	go dd.collector()
	return &dd
}

func (r *DownloadDispatcher) ChooChoo() {
	fmt.Println("🚂")
}

func (r *DownloadDispatcher) Dispatch(dl *Download) {
	for !r.counter.inc() {
		time.Sleep(time.Millisecond * 50)
	}
	go r.downloadJob(dl)
}

func (r *DownloadDispatcher) Close() {
	for !r.counter.zero() {
		time.Sleep(time.Millisecond * 50)
	}
	close(r.resc)
}

func (r *DownloadDispatcher) Collect() []*Download {
	return r.finished
}

func (r *DownloadDispatcher) collector() {
	for dl := range r.resc {
		r.finished = append(r.finished, dl)
		r.counter.dec()
	}
}

func (r *DownloadDispatcher) downloadJob(dl *Download) {
	defer func() {
		r.resc <- dl
	}()
	resp, err := dl.Client.Get(dl.Addr.String())
	if err != nil {
		dl.Err = err
		return
	}
	defer resp.Body.Close()
	if !dl.AllowOverride {
		exists := func() error {
			readtest, err := os.Open(dl.File)
			defer readtest.Close()
			if err == nil {
				return fmt.Errorf("File already exists: %q", dl.File)
			} else if !os.IsNotExist(err) {
				return err
			}
			return nil
		}
		if err := exists(); err != nil {
			dl.Err = err
			return
		}
	}
	f, err := os.Create(dl.File)
	if err != nil {
		dl.Err = err
		return
	}
	defer f.Close()
	if _, err := io.Copy(f, resp.Body); err != nil {
		dl.Err = err
		return
	}
}
