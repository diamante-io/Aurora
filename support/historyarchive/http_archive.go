// Copyright 2016 HcNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/hcnet/go/support/errors"
)

type HttpArchiveBackend struct {
	client http.Client
	base   url.URL
}

func checkResp(r *http.Response) error {
	if r.StatusCode >= 200 && r.StatusCode < 400 {
		return nil
	} else {
		return fmt.Errorf("Bad HTTP response '%s' for GET '%s'",
			r.Status, r.Request.URL.String())
	}
}

func (b *HttpArchiveBackend) GetFile(pth string) (io.ReadCloser, error) {
	var derived url.URL = b.base
	derived.Path = path.Join(derived.Path, pth)
	resp, err := b.client.Get(derived.String())
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	err = checkResp(resp)
	if err != nil {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
		return nil, err
	}
	return resp.Body, nil
}

func (b *HttpArchiveBackend) Exists(pth string) (bool, error) {
	var derived url.URL = b.base
	derived.Path = path.Join(derived.Path, pth)
	resp, err := b.client.Head(derived.String())
	if err != nil {
		return false, err
	}

	if resp.Body != nil {
		resp.Body.Close()
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return false, errors.Errorf("Unkown status code=%d", resp.StatusCode)
	}
}

func (b *HttpArchiveBackend) PutFile(pth string, in io.ReadCloser) error {
	in.Close()
	return errors.New("PutFile not available over HTTP")
}

func (b *HttpArchiveBackend) ListFiles(pth string) (chan string, chan error) {
	ch := make(chan string)
	er := make(chan error)
	close(ch)
	er <- errors.New("ListFiles not available over HTTP")
	close(er)
	return ch, er
}

func (b *HttpArchiveBackend) CanListFiles() bool {
	return false
}

func makeHttpBackend(base *url.URL, opts ConnectOptions) ArchiveBackend {
	return &HttpArchiveBackend{
		base: *base,
	}
}
