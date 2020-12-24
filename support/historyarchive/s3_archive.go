// Copyright 2016 DiamNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"bytes"
	"io"
	"net/http"
	"path"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/diamnet/go/support/errors"
)

type S3ArchiveBackend struct {
	svc              *s3.S3
	bucket           string
	prefix           string
	unsignedRequests bool
}

func (b *S3ArchiveBackend) GetFile(pth string) (io.ReadCloser, error) {
	params := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(path.Join(b.prefix, pth)),
	}

	req, resp := b.svc.GetObjectRequest(params)
	if b.unsignedRequests {
		req.Handlers.Sign.Clear() // makes this request unsigned
	}
	err := req.Send()
	if err != nil {
		return nil, err
	}

	return resp.Body, nil
}

func (b *S3ArchiveBackend) Exists(pth string) (bool, error) {
	params := &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(path.Join(b.prefix, pth)),
	}

	req, _ := b.svc.HeadObjectRequest(params)
	if b.unsignedRequests {
		req.Handlers.Sign.Clear() // makes this request unsigned
	}
	err := req.Send()
	if err != nil {
		return false, err
	}

	if req.HTTPResponse.StatusCode >= 200 && req.HTTPResponse.StatusCode < 400 {
		return true, nil
	} else if req.HTTPResponse.StatusCode == http.StatusNotFound {
		return false, nil
	} else {
		return false, errors.Errorf("Unkown status code=%d", req.HTTPResponse.StatusCode)
	}
}

func (b *S3ArchiveBackend) PutFile(pth string, in io.ReadCloser) error {
	var buf bytes.Buffer
	_, err := buf.ReadFrom(in)
	in.Close()
	if err != nil {
		return err
	}
	params := &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(path.Join(b.prefix, pth)),
		ACL:    aws.String(s3.ObjectCannedACLPublicRead),
		Body:   bytes.NewReader(buf.Bytes()),
	}
	req, _ := b.svc.PutObjectRequest(params)
	if b.unsignedRequests {
		req.Handlers.Sign.Clear() // makes this request unsigned
	}
	err = req.Send()

	in.Close()
	return err
}

func (b *S3ArchiveBackend) ListFiles(pth string) (chan string, chan error) {
	prefix := path.Join(b.prefix, pth)
	ch := make(chan string)
	errs := make(chan error)

	params := &s3.ListObjectsInput{
		Bucket:  aws.String(b.bucket),
		MaxKeys: aws.Int64(1000),
		Prefix:  aws.String(prefix),
	}
	req, resp := b.svc.ListObjectsRequest(params)
	if b.unsignedRequests {
		req.Handlers.Sign.Clear() // makes this request unsigned
	}
	err := req.Send()
	if err != nil {
		errs <- err
		close(ch)
		close(errs)
		return ch, errs
	}
	go func() {
		for {
			for _, c := range resp.Contents {
				params.Marker = c.Key
				ch <- *c.Key
			}
			if *resp.IsTruncated {
				req, resp = b.svc.ListObjectsRequest(params)
				if b.unsignedRequests {
					req.Handlers.Sign.Clear() // makes this request unsigned
				}
				err := req.Send()
				if err != nil {
					errs <- err
				}
			} else {
				break
			}
		}
		close(ch)
		close(errs)
	}()
	return ch, errs
}

func (b *S3ArchiveBackend) CanListFiles() bool {
	return true
}

func makeS3Backend(bucket string, prefix string, opts ConnectOptions) (ArchiveBackend, error) {
	cfg := aws.Config{
		Region:   aws.String(opts.S3Region),
		Endpoint: aws.String(opts.S3Endpoint),
	}

	sess, err := session.NewSession(&cfg)
	if err != nil {
		return nil, err
	}

	backend := S3ArchiveBackend{
		svc:              s3.New(sess),
		bucket:           bucket,
		prefix:           prefix,
		unsignedRequests: opts.UnsignedRequests,
	}
	return &backend, nil
}
