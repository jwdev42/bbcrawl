package download

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type NoFilenameInContentDisposition struct {
	url *string
}

func newNoFilenameInContentDisposition(url string) NoFilenameInContentDisposition {
	return NoFilenameInContentDisposition{url: &url}
}

func (e NoFilenameInContentDisposition) Error() string {
	return fmt.Sprintf("URL %q: No filename found in Content-disposition header", *e.url)
}

type DownloadFailure struct {
	url *string
}

func (e DownloadFailure) Error() string {
	return fmt.Sprintf("Download failed: %q", *e.url)
}

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
	dir           string
	file          string
	AllowOverride bool
	Err           error
}

func (dl *Download) Dir() string {
	return dl.dir
}

//Exists returns true if there already is a file named the same as specified in the Download struct.
//Panics if the field "dir" or "file" is zero-valued.
func (dl *Download) Exists() (bool, error) {
	f, err := os.Open(dl.Path())
	defer f.Close()
	if err == nil {
		//file exists
		return true, nil
	} else if os.IsNotExist(err) {
		//file does not exist
		return false, nil
	}
	//file status is not obtainable
	return false, err
}

func (dl *Download) File() string {
	return dl.file
}

//Path returns the absolute path to the (to be) downloaded file. Panics if the field "dir" or "file" is zero-valued.
func (dl *Download) Path() string {
	if dl.dir == "" || dl.file == "" {
		panic("a Download struct must have valid \"dir\" and \"file\" fields before Path() can be called")
	}
	return filepath.Join(dl.dir, dl.file)
}

//SetDir sets the download directory. Panics if dir is not an absolute path.
func (dl *Download) SetDir(dir string) error {
	if !filepath.IsAbs(dir) {
		panic("dir must be an absolute path")
	}
	dirtest, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer dirtest.Close()
	info, err := dirtest.Stat()
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", dir)
	}
	dl.dir = dir
	return nil
}

func (dl *Download) SetFile(file string) {
	if strings.IndexByte(file, os.PathSeparator) >= 0 {
		panic(fmt.Errorf("Filename %q is not allowed to contain the directory separator \"%c\"", file, os.PathSeparator))
	}
	dl.file = file
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

func (r *DownloadDispatcher) nameFromHeader(dl *Download) error {
	var filename string
	resp, err := dl.Client.Head(dl.Addr.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	header := resp.Header
	for _, v := range header.Values("Content-disposition") {
		field := strings.TrimSpace(v)
		if isHttpFileNameField(field) {
			filename = httpFileNameFieldValue(field)
			if filename == "" {
				return fmt.Errorf("malformed filename in Content-disposition header: %s", field)
			}
			break
		}
	}
	if filename == "" {
		return newNoFilenameInContentDisposition(dl.Addr.String())
	}
	dl.file = filename
	return nil
}

func (r *DownloadDispatcher) prepareJob(dl *Download) error {
	//safety check
	if len(dl.dir) == 0 {
		panic("directory not set")
	}

	//obtain filename from http header if no file was given
	if len(dl.file) == 0 {
		if err := r.nameFromHeader(dl); err != nil {
			return err
		}
	}

	//safety check
	if len(dl.file) == 0 {
		panic("filename not set")
	}

	//prevent an overridden file if not allowed
	if !dl.AllowOverride {
		if ex, err := dl.Exists(); ex {
			return fmt.Errorf("file already exists: %s", dl.Path())
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (r *DownloadDispatcher) downloadJob(dl *Download) {
	defer func() {
		r.resc <- dl
	}()

	//prepare the download
	if err := r.prepareJob(dl); err != nil {
		dl.Err = err
		return
	}

	//open connection
	resp, err := dl.Client.Get(dl.Addr.String())
	if err != nil {
		dl.Err = err
		return
	}
	defer resp.Body.Close()

	//create the local file
	f, err := os.Create(dl.Path())
	if err != nil {
		dl.Err = err
		return
	}
	defer f.Close()

	//copy received content to the local file
	if _, err := io.Copy(f, resp.Body); err != nil {
		dl.Err = err
		return
	}
}

func isHttpFileNameField(input string) bool {
	if strings.Index(input, "filename=\"") == 0 {
		return true
	}
	return false
}

//httpFileNameValue returns the filename of a "Content-disposition" HTTP header field, if it is not a filename, it returns "".
func httpFileNameFieldValue(input string) string {
	splitted := strings.Split(input, "=")
	if len(splitted) != 2 {
		return ""
	}
	if splitted[0] != "filename" {
		return ""
	}
	name := strings.Trim(splitted[1], "\"")
	if strings.IndexByte(name, os.PathSeparator) >= 0 {
		return ""
	}
	return name
}
