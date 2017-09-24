// Copyright 2017 Brian Starkey <stark3y@gmail.com>

package ratelimit

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	http.Client
	ticker *time.Ticker
	nreqs int
	period time.Duration
	throttle chan time.Time
}

func (r *Client) Do(req *http.Request) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Do(req)
}

func (r *Client) Get(url string) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Get(url)
}

func (r *Client) Head(url string) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Head(url)
}

func (r *Client) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.Post(url, contentType, body)
}

func (r *Client) PostForm(url string, data url.Values) (resp *http.Response, err error) {
	<-r.throttle
	return r.Client.PostForm(url, data)
}

func NewClient(nreqs int, period time.Duration) *Client {
	r := Client{
		Client: http.Client{},
		nreqs: nreqs,
		period: period,
		throttle: make(chan time.Time, nreqs),
	}

	r.ticker = time.NewTicker(r.period)
	n := r.nreqs
	for n > 0 {
		r.throttle <- time.Now()
		n--
	}

	go func() {
		for {
			select {
			case <-r.ticker.C:
				n := r.nreqs
				for n > 0 {
					select {
					case r.throttle <- time.Now():
					default:
					n--
					}
				}
			}
		}
	}()

	return &r
}
