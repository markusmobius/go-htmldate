# Go-HtmlDate

Go-HtmlDate is a Go package and command-line tool to extract the original and updated publication dates
of web pages. This package is based on [`htmldate`][0], a Python package by [Adrien Barbaresi][1].

The structure of this package is arranged following the structure of original Python code. This way, both
libraries should give similar performance and any improvements from the original can be ported easily.

## Table of Contents

- [Features](#features)
- [Status](#status)
- [Usage as Go package](#usage-as-go-package)
- [Usage as CLI Application](#usage-as-cli-application)
- [Comparison with Original](#comparison-with-original)
- [License](#license)

## Features

Go-HtmlDate extracts original or updated publication dates of web pages using several heuristics on HTML
code and linguistic patterns. There are four steps on extracting the dates:

- Extract the date from URL (if specified);
- Look for dates in the metadata, which done by parsing `<meta>` elements in HTML header;
- Look in time related HTML elements like `<time>` and `<abbr>`;
- Finally, when nothing found, scan the entire document's text to find potential dates.

Just like the original, Go-HtmlDate has two mode: fast and extensive. The differences are:

- in fast mode, the HTML page is cleaned and precise patterns are targeted;
- in extensive mode, all potential dates are collected and a disambiguation algorithm determines
  the best one to use.

By default Go-HtmlDate will run in extensive mode, and usually there are no reasons to use the fast mode.
This is because unlike in the original, in our Go port the extraction speed between the fast and extensive
mode is negligible, so might as well use the extensive mode.

## Status

This package is stable enough for use and up to date with the original `htmldate` commit [d6d34d3][2].
However, there are some difference between this port and the original Trafilatura.

First, in the original they have two kind of date parser: the fast and slow one (for extensive mode). For
the slow one, they use [`scrapinghub/dateparser`][3], a powerful date parser which can parse date from
almost any string in many languages. Unfortunately, it hasn't been ported to Go and as far as we know
there are no package as powerful as it. So, in this port we just modify their fast parser to make our
extensive mode has similar performance as the original.

We also added several month name translations for French and Indonesian language, so it should works
to certain extend for web pages from those country.

## Usage as Go package

Run following command inside your Go project :

```
go get -u -v github.com/markusmobius/go-htmldate
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
  -t, --timeout int         timeout for downloading web page in seconds (default 30)
  -u, --user-agent string   set custom user agent (default "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0")
  -v, --verbose             enable log message
```

## Comparison with Original

Here we compare the extraction performance between the fast and extensive mode. To reproduce this test,
clone this repository then run:

```
go run scripts/comparison/*.go
```

For the test, we use 255 documents which were selected from large collections of web pages in German.
For the sake of completeness a few documents in other languages were added (mostly in English and French
but also in other European languages, Chinese, Japanese and Arabic). Here is the result when tested in my
PC (Intel i7-8550U @ 4.000GHz, RAM 16 GB):

|             Package            | Precision | Recall | Accuracy | F-Score | Speed (s) |
|:------------------------------:|:---------:|:------:|:--------:|:-------:|:---------:|
|      `go-htmldate` fast        |   0.919   |  0.933 |   0.862  |  0.926  |   0.344   |
|    `go-htmldate` extensive     |   0.911   |  1.000 |   0.911  |  0.953  |   0.488   |

For comparison, here is the result of the original htmldate:

|             Package            | Precision | Recall | Accuracy | F-Score | Speed (s) |
|:------------------------------:|:---------:|:------:|:--------:|:-------:|:---------:|
|        `htmldate` fast         |   0.899   |  0.917 |   0.831  |  0.908  |   1.241   |
|      `htmldate` extensive      |   0.893   |  1.000 |   0.893  |  0.944  |   2.129   |

From the tables above, our port has slightly better performance than the original. I believe this is
mostly because of the improvements that I done to the fast parser. Regarding speed, our port is much
faster compared than the original. This is as expected considering Go is famous for its performance.

> By the way, to calculate the speed of the original htmldate, I've made a little modification for
> their comparison script. This is because in their original script they don't reuse the html document,
> and instead parsing the document everytime they want to extract the date. Thanks to this, most of the
> time in their script is used for parsing instead of extracting, so I decided to modify it a bit to make
> it similar with our port.

## License

Like the original, `go-htmldate` is distributed under the [GNU General Public License v3.0](LICENSE).

[0]: https://github.com/adbar/htmldate
[1]: https://github.com/adbar
[2]: https://github.com/adbar/htmldate/commit/d6d34d3ae82d43ce7a2549a51d62584a5afb078f
[3]: https://github.com/scrapinghub/dateparser