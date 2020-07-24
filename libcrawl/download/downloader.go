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
	id            uint64 //the id is assigned by the DownloadDispatcher
	dir           string
	file          string
	tempname      bool
	header        http.Header
	AllowOverride bool
	Err           error
	AfterDownload func(*Download)
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

func (dl *Download) NameFromHeader() (string, error) {
	var filename string
	header := dl.header
	if header == nil {
		panic("NameFromHeader is not supposed to be called until the download has finished")
	}
	for _, v := range header.Values("Content-disposition") {
		field := strings.TrimSpace(v)
		if isHttpFileNameField(field) {
			filename = httpFileNameFieldValue(field)
			if filename == "" {
				return "", fmt.Errorf("malformed filename in Content-disposition header: %s", field)
			}
			break
		}
	}
	return filename, nil
}

//Path returns the absolute path to the (to be) downloaded file. Panics if the field "dir" or "file" is zero-valued.
func (dl *Download) Path() string {
	if dl.dir == "" || dl.file == "" {
		panic("a Download struct must have valid \"dir\" and \"file\" fields before Path() can be called")
	}
	return filepath.Join(dl.dir, dl.file)
}

//Rename changes the file name of a download. If the download already exists, it will be renamed on the file system.
func (dl *Download) Rename(name string) error {
	if strings.IndexByte(name, os.PathSeparator) >= 0 {
		panic(fmt.Errorf("Filename %q is not allowed to contain the directory separator \"%c\"", name, os.PathSeparator))
	}
	file, err := os.Open(dl.Path())
	if os.IsNotExist(err) {
		dl.SetFile(name)
		return nil
	} else if err != nil {
		return err
	}
	file.Close()
	newpath := filepath.Join(dl.Dir(), name)
	if err := os.Rename(dl.Path(), newpath); err != nil {
		return err
	}
	dl.SetFile(name)
	return nil
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

func (dl *Download) SetFile(name string) {
	if strings.IndexByte(name, os.PathSeparator) >= 0 {
		panic(fmt.Errorf("Filename %q is not allowed to contain the directory separator \"%c\"", name, os.PathSeparator))
	}
	dl.file = name
	dl.tempname = false
}

type DownloadDispatcher struct {
	max       int
	counter   *threadcounter
	dlcounter *DownloadCounter
	resc      chan *Download //yields the state of finished download routines
}

func NewDownloadDispatcher(downloads int) *DownloadDispatcher {
	if downloads < 1 {
		panic("parameter downloads must be > 0")
	}
	dd := DownloadDispatcher{
		max:       downloads,
		counter:   &threadcounter{m: new(sync.Mutex), max: downloads},
		dlcounter: NewDownloadCounter(),
		resc:      make(chan *Download, downloads),
	}
	return &dd
}

func (r *DownloadDispatcher) ChooChoo() {
	fmt.Println("🚂")
}

func (r *DownloadDispatcher) Dispatch(dl *Download) {
	dl.id = r.dlcounter.Count()
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

func (r *DownloadDispatcher) Collect() *Download {
	dl := <-r.resc
	if dl != nil {
		r.counter.dec()
	}
	return dl
}

func (r *DownloadDispatcher) prepareJob(dl *Download) error {
	//safety check
	if len(dl.dir) == 0 {
		panic("directory not set")
	}
	//use an automatically generated file name if none was provided
	if len(dl.file) == 0 {
		dl.file = fmt.Sprintf("%d.download", dl.id)
		dl.tempname = true
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

	//copy http header fields
	dl.header = resp.Header.Clone()

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

	//call AfterDownload routine if available
	if dl.AfterDownload != nil {
		dl.AfterDownload(dl)
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
