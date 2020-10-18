# bbcrawl

## synopsis

bbcrawl is a tool that can crawl through bulletin board threads and scrape data off it.

## usage

bbcrawl is operated via the command line. A successful command has to satisfy the following scheme:  
> **bbcrawl** *\[global_options\]* **-pager** *pager_name* *\[pager_options\]* **-crawler** *crawler_name* *\[crawler_options\]* *URL*

#### global options
> **-o** *PATH*  
> o sets the output directory where the downloads should be placed into. Default value is the current workdir.

> **-cookie-file** *PATH*  
> cookie-file loads cookies from the given file. The file must be in the same format as the one used by
> [curl](https://curl.haxx.se/docs/http-cookies.html).

## pagers
A pager generates the URLs that will be sent to the crawler module. Every pager takes the URL from the end of the bbcrawl command
as a blueprint. A manipulated URL based on that blueprint will be sent to the crawler everytime when it requests a new page.
The blueprint url typically refers to a bulletin board thread.

#### common pager options
These options work on every pager

> **-end** *INT*  
> end tells the pager the number of the last page.

> **-start** *INT*  
> start tells the pager the number of the first page.

### cutter
cutter generates URLs by cutting away a user-defined part of the blueprint URL. That cut out part will then be replaced
by the page number.

#### options for cutter
> **-adjust** *INT*  
> adjust adjusts the page number that is reported to the crawler. 

> **-cut** *INT*,*INT*  
> cut cuts out a part of the blueprint url, this part will be replaced with a page number
> everytime the crawler requests a new page.
> The first argument is the index of the first character to be cut out,
> the second argument is the amount of characters to be cut out.
> The index starts with 1.

> **-digits** *INT*  
> digits tells the pager the amount of the leading zeros the page number should have. *0* is automatic,
> which means no leading zeros (default). If the integer argument of *digits* is smaller than length of the
> largest page number, bbcrawl will abort.

> **-startpage** *URL*  
> startpage takes an url that is sent to the crawler only at the first iteration.
> When the crawler requests the second page, the normal logic will be used.
> Please note that the pager doesn't increase the page number at its first iteration if *startpage* is set.

> **-step** *INT*  
> the page number is multiplied by *step* before generating the URL. Default value is 1.

### query
query generates URLs by manipulating the query string of a blueprint URL.

#### options for query
> **-name** *STRING*  
> name sets the variable identifier that is responsible for selecting a page in the url query string.
> The default value for name is *page*.

### vb4
vb4 generates URLs for vbulletin 4 threads.

## crawlers

#### common crawler options
These options work on every crawler 

> **-exclude** *URL\{,URL\}*  
> exclude accepts a list of urls that are ignored and not downloaded.

> **-redirect** *BOOLEAN*  
> if redirect is true (default), the crawler will follow http redirects. If redirect is false, the crawler will produce an error
> if it encounters a http redirect.

### file
file is a crawler that treats every received page as a file for download.

### src
src downloads sources from audio, img and video tags.

#### options for src
> **-attrs** *ATTRIBUTES*  
> attrs filters img tags for the given html attributes. For the specification, see [attr_spec.txt](attr_spec.txt).
>> Example:

>> *-attrs width=500/alt=lorem ipsum*  
>> This will only download images that have a width attribute with value 500 an an alt attribute with value "lorem ipsum"

> **-tags** *TAG\{,TAG\}*  
> tags defines the tags the crawler will download sources from. Tags are supplied as a comma-separated list,
> valid tags are "audio", "img", "video".

### vb-attachments
vb-attachments downloads every vbulletin attachment found on a page. It currently supports vbulletin versions 3 and 4.

#### options for vb-attachments

> **-names-from-header** *BOOLEAN*  
> if true, the downloader will use the file names sent via the http header. False by default.

## examples

	bbcrawl -o /home/test -pager query -start 3 -end 15 -crawler img https://example.net/thread1293?page=5
This will download all images of thread1293 from page 3 to 15 to the directory /home/test

## issues

- passwords provided in URLs like https://user:pass@example.net may leak to the logger as there is no filter implemented yet.

bbcrawl is in an early state so expect bugs and changes.
This software is developed and tested under Linux, platform specific bugs may occur on other platforms.
You are welcome to report any bugs on plattforms that are supported by go.

bbcrawl is provided \*as is\* without any warranty.
© 2020 Jörg Walter
