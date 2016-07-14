# anom
[![GoDoc](https://godoc.org/github.com/Bogdanp/anom?status.svg)](http://godoc.org/github.com/Bogdanp/anom)
[![Build Status](https://travis-ci.org/Bogdanp/anom.svg?branch=master)](https://travis-ci.org/Bogdanp/anom)

`go get github.com/Bogdanp/anom`

Package anom is a simple "object mapper" for appengine datastore
and that provides some convenience functions for dealing with
datastore entities.

## Usage

Declare your models:

``` go
package models

import (
	"github.com/Bogdanp/anom"
)

type Post struct {
	anom.Meta

	Title string
	Content string
}
```

Then use them:

``` go
package app

import (
	"net/http"

	"github.com/Bogdanp/anom"

	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
)

func createPostHandler(rw http.ResponseWriter, req *http.Request) {
	ctx := appengine.NewContext(req)
	p := &models.Post{
		Title: req.PostFormValue("title"),
		Content: req.PostFormValue("content"),
	}
	if err := anom.Put(ctx, p); err != nil {
		panic(err)
	}
	log.Infof(ctx, "post: %v", p)
}

func getPostHandler(rw http.ResponseWriter, req *http.Request) {
	var p models.Post
	ctx := appengine.NewContext(req)
	if err := anom.Get(ctx, &p); err != nil {
		panic(err)
	}
	log.Infof(ctx, "post: %v", p)
}
```
