// Copyright 2016 HcNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"io"
	"io/ioutil"

	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

type XdrStream struct {
	buf        bytes.Buffer
	rdr        io.ReadCloser
	rdr2       io.ReadCloser
	sha256Hash hash.Hash

	validateHash bool
	expectedHash [sha256.Size]byte
}

func NewXdrStream(in io.ReadCloser) *XdrStream {
	// We write all we read from in to sha256Hash that can be later
	// compared with `expectedHash` using SetExpectedHash and Close.
	sha256Hash := sha256.New()
	teeReader := io.TeeReader(in, sha256Hash)

	return &XdrStream{
		rdr: struct {
			io.Reader
			io.Closer
		}{bufio.NewReader(teeReader), in},
		sha256Hash: sha256Hash,
	}
}

func NewXdrGzStream(in io.ReadCloser) (*XdrStream, error) {
	rdr, err := gzip.NewReader(bufReadCloser(in))
	if err != nil {
		in.Close()
		return nil, err
	}

	stream := NewXdrStream(rdr)
	stream.rdr2 = in
	return stream, nil
}

func HashXdr(x interface{}) (Hash, error) {
	var msg bytes.Buffer
	_, err := xdr.Marshal(&msg, x)
	if err != nil {
		var zero Hash
		return zero, err
	}
	return Hash(sha256.Sum256(msg.Bytes())), nil
}

// SetExpectedHash sets expected hash that will be checked in Close().
// This (obviously) needs to be set before Close() is called.
func (x *XdrStream) SetExpectedHash(hash [sha256.Size]byte) {
	x.validateHash = true
	x.expectedHash = hash
}

// Close closes all internal readers and checks if the expected hash
// (if set by SetExpectedHash) matches the actual hash of the stream.
func (x *XdrStream) Close() error {
	if x.validateHash {
		// Read all remaining data from rdr
		_, err := io.Copy(ioutil.Discard, x.rdr)
		if err != nil {
			return errors.Wrap(err, "Error reading remaining bytes from rdr")
		}

		actualHash := x.sha256Hash.Sum([]byte{})

		if !bytes.Equal(x.expectedHash[:], actualHash[:]) {
			return errors.New("Stream hash does not match expected hash!")
		}
	}

	if x.rdr != nil {
		if err := x.rdr.Close(); err != nil {
			return err
		}
	}
	if x.rdr2 != nil {
		if err := x.rdr2.Close(); err != nil {
			return err
		}
	}

	return nil
}

func (x *XdrStream) ReadOne(in interface{}) error {
	var nbytes uint32
	err := binary.Read(x.rdr, binary.BigEndian, &nbytes)
	if err != nil {
		x.rdr.Close()
		if err == io.EOF {
			// Do not wrap io.EOF
			return err
		}
		return errors.Wrap(err, "binary.Read error")
	}
	nbytes &= 0x7fffffff
	x.buf.Reset()
	if nbytes == 0 {
		x.rdr.Close()
		return io.EOF
	}
	x.buf.Grow(int(nbytes))
	read, err := x.buf.ReadFrom(io.LimitReader(x.rdr, int64(nbytes)))
	if err != nil {
		x.rdr.Close()
		return err
	}
	if read != int64(nbytes) {
		x.rdr.Close()
		return errors.New("Read wrong number of bytes from XDR")
	}

	readi, err := xdr.Unmarshal(&x.buf, in)
	if err != nil {
		x.rdr.Close()
		return err
	}
	if int64(readi) != int64(nbytes) {
		return fmt.Errorf("Unmarshalled %d bytes from XDR, expected %d)",
			readi, nbytes)
	}
	return nil
}

func WriteFramedXdr(out io.Writer, in interface{}) error {
	var tmp bytes.Buffer
	n, err := xdr.Marshal(&tmp, in)
	if err != nil {
		return err
	}
	un := uint32(n)
	if un > 0x7fffffff {
		return fmt.Errorf("Overlong write: %d bytes", n)
	}

	un = un | 0x80000000
	binary.Write(out, binary.BigEndian, &un)
	k, err := tmp.WriteTo(out)
	if int64(n) != k {
		return fmt.Errorf("Mismatched write length: %d vs. %d", n, k)
	}
	return err
}
