package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/levigross/grequests"
	"gopkg.in/urfave/cli.v2"
)

func action(ctx *cli.Context) error {
	var url string
	if len(ctx.String("url")) > 0 {
		url = ctx.String("url")
	} else if len(ctx.Args().Get(0)) > 0 {
		url = ctx.Args().Get(0)
	} else {
		return errors.New("No URL specified")
	}

	ro := &grequests.RequestOptions{}
	ro.InsecureSkipVerify = ctx.Bool("insecure")
	ro.UserAgent = ctx.String("user-agent")

	headers := ctx.StringSlice("header")
	if len(headers) > 0 {
		ro.Headers = make(map[string]string)
		for _, header := range headers {
			hs := strings.SplitN(header, ":", 2)
			if len(hs) == 2 {
				ro.Headers[hs[0]] = hs[1]
			}
		}
	}

	data := ctx.String("data")
	if len(data) > 0 {
		if data[:1] == "@" {
			dataFile := data[1:]
			if _, err := os.Stat(dataFile); os.IsNotExist(err) {
				fmt.Println("Warning: Couldn't read data from file \"" +
					dataFile + "\", this makes an empty POST.")
			} else {
				fileReader, err := os.Open(dataFile)
				if err != nil {
					return err
				}
				ro.RequestBody = fileReader
			}
		} else {
			ro.RequestBody = bytes.NewReader([]byte(data))
		}
	}

	method := ctx.String("request")
	if !ctx.IsSet("request") && ctx.IsSet("head") {
		method = "HEAD"
	}
	if !ctx.IsSet("request") && ctx.IsSet("data") {
		method = "POST"
	}

	if method == "POST" {
		if ro.Headers == nil {
			ro.Headers = make(map[string]string, 0)
		}
		if _, ok := ro.Headers["Content-Type"]; !ok {
			ro.Headers["Content-Type"] = "application/x-www-form-urlencoded"
		}
	}

	user := ctx.String("user")
	if len(user) > 0 {
		ro.Auth = strings.Split(user, ":")
	}

	ro.HTTPClient = grequests.BuildHTTPClient(*ro)
	if !ctx.Bool("location") {
		ro.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	var rsp *grequests.Response
	var err error
	switch method {
	case "GET":
		rsp, err = grequests.Get(url, ro)
	case "HEAD":
		rsp, err = grequests.Head(url, ro)
	case "POST":
		rsp, err = grequests.Post(url, ro)
	case "PUT":
		rsp, err = grequests.Put(url, ro)
	case "OPTIONS":
		rsp, err = grequests.Options(url, ro)
	case "PATCH":
		rsp, err = grequests.Patch(url, ro)
	case "DELETE":
		rsp, err = grequests.Delete(url, ro)
	default:
		return errors.New("Unsupported COMMAND")
	}
	if err == nil {
		if ctx.Bool("include") || ctx.Bool("head") {
			rawRsp := rsp.RawResponse
			fmt.Println(rawRsp.Proto + " " + rawRsp.Status)
			for k, v := range rsp.Header {
				fmt.Println(k + ": " + strings.Join(v, ";"))
			}
			fmt.Println("")
		}
		if !ctx.Bool("head") {
			fmt.Print(rsp)
		}
		return nil
	}
	return err
}

func main() {
	cli.AppHelpTemplate = `usage: {{.Name}} [options...] <url>
author: {{range .Authors}}{{ . }}{{end}}
options:
    {{range .VisibleFlags}}{{.}}
    {{end}}`

	fetch := &cli.App{
		Name:    "fetch",
		Version: "0.0.2",
		Authors: []*cli.Author{
			&cli.Author{Name: "qiyi", Email: "bphanzhu@gmail.com"}},
		Description: "URL fetch application",
		Action:      action,
	}
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:    "data",
			Aliases: []string{"d"},
			Usage:   "HTTP POST `DATA`",
		},
		&cli.StringSliceFlag{
			Name:    "header",
			Aliases: []string{"H"},
			Usage:   "Custom header `LINE` to pass to server",
		},
		&cli.BoolFlag{
			Name:    "head",
			Aliases: []string{"I"},
			Usage:   "Show document info only",
		},
		&cli.BoolFlag{
			Name:    "include",
			Aliases: []string{"i"},
			Usage:   "Include protocol headers in the output",
		},
		&cli.BoolFlag{
			Name:    "insecure",
			Aliases: []string{"k"},
			Usage:   "Allow connections to SSL sites without certs",
		},
		&cli.BoolFlag{
			Name:    "location",
			Aliases: []string{"L"},
			Usage:   "Follow redirects",
		},
		&cli.StringFlag{
			Name:    "request",
			Aliases: []string{"X"},
			Value:   "GET",
			Usage:   "Specify request `COMMAND` to use",
		},
		&cli.StringFlag{
			Name:  "url",
			Usage: "`URL` to work with",
		},
		&cli.StringFlag{
			Name:    "user",
			Aliases: []string{"u"},
			Usage:   "`USER`[:PASSWORD][;OPTIONS]  Server user, password and login options",
		},
		&cli.StringFlag{
			Name:    "user-agent",
			Aliases: []string{"A"},
			Value:   fetch.Name + "/" + fetch.Version,
			Usage:   "User-Agent to send to server",
		},
	}
	fetch.Flags = flags
	fetch.Run(os.Args)
}
