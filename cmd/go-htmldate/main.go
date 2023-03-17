// This file is part of go-htmldate, Go package for extracting publication dates from a web page.
// Source available in <https://github.com/markusmobius/go-htmldate>.
// Copyright (C) 2022 Markus Mobius
//
// This program is free software: you can redistribute it and/or modify it under the terms of
// the GNU General Public License as published by the Free Software Foundation, either version 3
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program.
// If not, see <https://www.gnu.org/licenses/>.

// Code in this file is ported from <https://github.com/adbar/htmldate> which available under
// GNU GPL v3 license.

package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"net/http"
	nurl "net/url"
	"os"
	fp "path/filepath"
	"strings"
	"time"

	"github.com/markusmobius/go-htmldate"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"golang.org/x/net/html"
)

const (
	// defaultUserAgent is the default user agent to use, which is Firefox's.
	defaultUserAgent = "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0"

	// defaultFormat is the default output format
	defaultFormat = "2006-01-02"
)

var (
	resultZero = htmldate.Result{}

	log = zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: "2006-01-02 15:04",
	}).With().Timestamp().Logger()
)

func main() {
	// Create root command
	rootCmd := &cobra.Command{
		Use:   "go-htmldate [flags] [source]",
		Run:   rootCmdHandler,
		Short: "Extract publish date from a HTML file or url",
		Args:  cobra.ExactArgs(1),
	}

	// Register persistent flags
	flags := rootCmd.PersistentFlags()
	flags.Bool("time", false, "extract publish time as well")
	flags.Bool("ori", false, "extract original date instead of the the most recent one")
	flags.BoolP("verbose", "v", false, "enable log message")
	flags.IntP("timeout", "t", 30, "timeout for downloading web page in seconds")
	flags.Bool("skip-tls", false, "skip X.509 (TLS) certificate verification")
	flags.StringP("format", "f", defaultFormat, "set custom date output format")
	flags.StringP("user-agent", "u", defaultUserAgent, "set custom user agent")

	// Execute
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}

func rootCmdHandler(cmd *cobra.Command, args []string) {
	// Process source
	source := args[0]
	opts := createExtractorOptions(cmd)
	httpClient := createHttpClient(cmd)
	userAgent, _ := cmd.Flags().GetString("user-agent")
	outputFormat, _ := cmd.Flags().GetString("format")

	var err error
	var result htmldate.Result

	if fileExists(source) {
		result, err = processFile(source, opts)
	} else if parsedURL, isValid := validateURL(source); isValid {
		result, err = processURL(httpClient, userAgent, parsedURL, opts)
	}

	if err != nil {
		log.Fatal().Msgf("failed to extract %s: %v", source, err)
	}

	if result.IsZero() {
		log.Fatal().Msgf("failed to extract %s: no date found", source)
	}

	// Print result
	fmt.Println(result.Format(outputFormat))
}

func processFile(path string, opts htmldate.Options) (htmldate.Result, error) {
	// Open file
	f, err := os.Open(path)
	if err != nil {
		return resultZero, err
	}
	defer f.Close()

	// Make sure it's html
	var fReader io.Reader
	mimeType := mime.TypeByExtension(fp.Ext(path))
	if strings.Contains(mimeType, "text/html") {
		fReader = f
	} else {
		buffer := bytes.NewBuffer(nil)
		tee := io.TeeReader(f, buffer)

		_, err := html.Parse(tee)
		if err != nil {
			return resultZero, fmt.Errorf("%s is not a valid html file: %v", path, err)
		}

		fReader = buffer
	}

	// Extract
	return htmldate.FromReader(fReader, opts)
}

func processURL(client *http.Client, userAgent string, url *nurl.URL, opts htmldate.Options) (htmldate.Result, error) {
	// Download URL
	strURL := url.String()
	log.Info().Msgf("downloading %s", strURL)

	resp, err := download(client, userAgent, strURL)
	if err != nil {
		return resultZero, err
	}
	defer resp.Body.Close()

	// Make sure it's html
	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") {
		return resultZero, fmt.Errorf("page is not html: \"%s\"", contentType)
	}

	// Extract
	opts.URL = strURL
	return htmldate.FromReader(resp.Body, opts)
}

func createExtractorOptions(cmd *cobra.Command) htmldate.Options {
	var opts htmldate.Options

	flags := cmd.Flags()
	opts.ExtractTime, _ = flags.GetBool("time")
	opts.UseOriginalDate, _ = flags.GetBool("ori")
	opts.EnableLog, _ = flags.GetBool("verbose")
	return opts
}

func createHttpClient(cmd *cobra.Command) *http.Client {
	flags := cmd.Flags()
	timeout, _ := flags.GetInt("timeout")
	skipTls, _ := flags.GetBool("skip-tls")

	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: skipTls,
			},
		},
	}
}

func download(client *http.Client, userAgent string, url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func validateURL(url string) (*nurl.URL, bool) {
	parsedURL, err := nurl.ParseRequestURI(url)
	if err != nil {
		return nil, false
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return nil, false
	}

	return parsedURL, true
}
