# Go-HtmlDate [![Go Reference][ref-badge]][ref-link]

Go-HtmlDate is a Go package and command-line tool to extract the original and updated publication dates
of web pages. This package is based on [`htmldate`][0], a Python package by [Adrien Barbaresi][1].

The structure of this package is arranged following the structure of original Python code. This way, both
libraries should give similar performance and any improvements from the original can be ported easily.

## Table of Contents

- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Status](#status)
- [Usage as Go package](#usage-as-go-package)
- [Usage as CLI Application](#usage-as-cli-application)
- [Performance](#performance)
- [Additional Notes](#additional-notes)
- [Acknowledgements](#acknowledgements)
- [License](#license)

## Features

- Extracts original or updated publication date of web pages;
- **EXPERIMENTAL**: Extracts original or updated publication time (and its timezone) as well;

Just like the original, Go-HtmlDate has two mode: fast and extensive. The differences are:

- In fast mode, the HTML page is cleaned and precise patterns are targeted;
- In extensive mode, Go-HtmlDate will also collects all potential dates and uses a disambiguation
  algorithm to determines the best one to use.

By default Go-HtmlDate will run in extensive mode, and usually there are no reasons to use the fast
mode. This is because unlike in the original, in our Go port the extraction speed between the fast
and extensive mode is negligible, so might as well use the extensive mode.

## Status

This package is stable enough for use and up to date with the original `htmldate` (commit [c0b2aee][3]).
However, since time extraction is a brand new feature which doesn't exist in the original, use it with
care. So far it works quite nicely on most news sites that I've tried, but it still needs more testing.

When time extraction is enabled, there a some behaviors that I'd like to note:

- If time is not found or not specified in the web page, the time will be set into `00:00:00` (it will
  only returns the date).
- If timezone is not found or not specified in the web page, the timezone will be set into `time.UTC`.

In future I hope we could improve the comparison script to check the accuracy for time extraction as well.

## Usage as Go package

Run following command inside your Go project :

```
go get -u -v github.com/markusmobius/go-htmldate@master
```

Next, include it in your application :

```go
import "github.com/markusmobius/go-htmldate"
```

Now you can use Trafilatura to extract date of a web page. For basic usage you can check the
[examples](examples).

## Usage as CLI Application

To use CLI, you need to build it from source. Make sure you use `go >= 1.16` then run following commands :

```
go install github.com/markusmobius/go-htmldate/cmd/go-htmldate@master
```

Once installed, you can use it from your terminal:

```
$ go-htmldate -h
Extract publish date from a HTML file or url

Usage:
  go-htmldate [flags] [source]

Flags:
  -f, --format string       set custom date output format (default "2006-01-02")
  -h, --help                help for go-htmldate
      --ori                 extract original date instead of the the most recent one
      --skip-tls            skip X.509 (TLS) certificate verification
      --time                extract publish time as well
  -t, --timeout int         timeout for downloading web page in seconds (default 30)
  -u, --user-agent string   set custom user agent (default "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
  -v, --verbose             enable log message
```

## Performance

Here we compare the extraction performance between the fast and extensive mode. To reproduce this test,
clone this repository then run:

```
go run scripts/comparison/*.go
```

For the test, we use 500 documents which taken from two sources:

- 225 documents from BBAW collection by Adrien Barbaresi and Lukas Kozmus.
- 275 documents from [Data Culture Group][dcg] from Northeastern University for additional English news.

Here is the result when tested in my PC (Intel i7-8550U @ 4.000GHz, RAM 16 GB):

|         Package         | Precision | Recall | Accuracy | F-Score | Speed (s) |
| :---------------------: | :-------: | :----: | :------: | :-----: | :-------: |
|   `go-htmldate` fast    |   0.842   | 0.920  |  0.786   |  0.879  |   1.429   |
| `go-htmldate` extensive |   0.831   | 0.993  |  0.826   |  0.905  |   3.035   |

So, from the table above, this port has a similar performance with the original `htmldate`.

## Additional Notes

Despite the impressive score above, there is a little caveat: the performance test is only used to
measure the accuracy of publish date extraction, and **not** the modified date. This issue is occured in
the original `htmldate` as well since we use the comparison data from there.

With that said, if you use this package for extracting the modified date, the performance might not be
as good as the performance table above. However, it should be still good enough to use.

> The weird thing is the default behavior for original `htmldate` is to extract the modified date
> instead of the original, so ideally the performance test is done for modified date as well.
> To be fair, collecting the modified date seems harder than collecting the original date though.

## Acknowledgements

This package won't be exist without effort by Adrien Barbaresi, the author of the original Python package. He created `htmldate` as part of effort to build [text databases for research][k-web]. There are web pages for which neither the URL nor the server response provide a reliable way to find out when a document was published or modified. For more information:

```
@article{barbaresi-2020-htmldate,
  title = {{htmldate: A Python package to extract publication dates from web pages}},
  author = "Barbaresi, Adrien",
  journal = "Journal of Open Source Software",
  volume = 5,
  number = 51,
  pages = 2439,
  url = {https://doi.org/10.21105/joss.02439},
  publisher = {The Open Journal},
  year = 2020,
}
```

- Barbaresi, A. ["htmldate: A Python package to extract publication dates from web pages"][paper-1], Journal of Open Source Software, 5(51), 2439, 2020. DOI: 10.21105/joss.02439
- Barbaresi, A. ["Generic Web Content Extraction with Open-Source Software"][paper-2], Proceedings of KONVENS 2019, Kaleidoscope Abstracts, 2019.
- Barbaresi, A. ["Efficient construction of metadata-enhanced web corpora"][paper-3], Proceedings of the [10th Web as Corpus Workshop (WAC-X)][wac-x], 2016.

## License

Like the original, `go-htmldate` is distributed under the [GNU General Public License v3.0](LICENSE).

[0]: https://github.com/adbar/htmldate
[1]: https://github.com/adbar
[2]: https://github.com/adbar/htmldate/tree/v1.1.1
[3]: https://github.com/adbar/htmldate/commit/c0b2aee
[dcg]: https://dataculturegroup.org
[ref-badge]: https://pkg.go.dev/badge/github.com/markusmobius/go-htmldate.svg
[ref-link]: https://pkg.go.dev/github.com/markusmobius/go-htmldate
[paper-1]: https://doi.org/10.21105/joss.02439
[paper-2]: https://hal.archives-ouvertes.fr/hal-02447264/document
[paper-3]: https://hal.archives-ouvertes.fr/hal-01371704v2/document
[wac-x]: https://www.sigwac.org.uk/wiki/WAC-X
[k-web]: https://www.dwds.de/d/k-web
