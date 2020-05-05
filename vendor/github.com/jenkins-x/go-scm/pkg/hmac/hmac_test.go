// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hmac

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"hash"
	"testing"
)

func TestValidatePrefix(t *testing.T) {
	tests := []struct {
		msg string
		key string
		sig string
		res bool
	}{
		//
		// sha256
		//
		{
			msg: "bonjour monde",
			key: "topsecret",
			sig: "sha256=8ca57e2afbad9fea8860404575c2d61827995c62aacd4c514eae4c404896390b",
			res: true,
		},
		{
			msg: "hello world",
			key: "topsecret",
			sig: "sha256=8ca57e2afbad9fea8860404575c2d61827995c62aacd4c514eae4c404896390b",
			res: false,
		},
		//
		// sha1
		//
		{
			msg: "bonjour monde",
			key: "topsecret",
			sig: "sha1=f25bad540601ff3131736e24a48dd928fa9ccc93",
			res: false,
		},
		{
			msg: "hello world",
			key: "topsecret",
			sig: "sha1=f25bad540601ff3131736e24a48dd928fa9ccc93",
			res: true,
		},
		//
		// algorithm not supported
		//
		{
			msg: "bonjour monde",
			key: "topsecret",
			sig: "md5=f7e48000ca338292875859154295d2fe",
			res: false,
		},
		//
		// algorithm not specified
		//
		{
			msg: "bonjour monde",
			key: "topsecret",
			sig: "8ca57e2afbad9fea8860404575c2d61827995c62aacd4c514eae4c404896390b",
			res: false,
		},
	}

	for _, test := range tests {
		res := ValidatePrefix(
			[]byte(test.msg),
			[]byte(test.key),
			test.sig,
		)
		if res != test.res {
			t.Errorf("Want valid %v for message %q",
				test.res, test.msg)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		alg func() hash.Hash
		msg string
		key string
		sig string
		res bool
	}{
		//
		// sha256
		//
		{
			alg: sha256.New,
			msg: "bonjour monde",
			key: "topsecret",
			sig: "8ca57e2afbad9fea8860404575c2d61827995c62aacd4c514eae4c404896390b",
			res: true,
		},
		{
			alg: sha256.New,
			msg: "hello world",
			key: "topsecret",
			sig: "8ca57e2afbad9fea8860404575c2d61827995c62aacd4c514eae4c404896390b",
			res: false,
		},
		//
		// sha1
		//
		{
			alg: sha1.New,
			msg: "hello world",
			key: "topsecret",
			sig: "f25bad540601ff3131736e24a48dd928fa9ccc93",
			res: true,
		},
		{
			alg: sha1.New,
			msg: "bonjour monde",
			key: "topsecret",
			sig: "f25bad540601ff3131736e24a48dd928fa9ccc93",
			res: false,
		},
		//
		// md5
		//
		{
			alg: md5.New,
			msg: "hello world",
			key: "topsecret",
			sig: "223a982e3a9eeaf1ebae1b458464d90b",
			res: true,
		},
		{
			alg: md5.New,
			msg: "bonjour monde",
			key: "topsecret",
			sig: "223a982e3a9eeaf1ebae1b458464d90b",
			res: false,
		},
		//
		// invalid hex
		//
		{
			alg: sha1.New,
			msg: "hello world",
			key: "topsecret",
			sig: "f25bad540601ff3131736e24a48dd928fa9ccc93==",
			res: false,
		},
	}

	for _, test := range tests {
		res := Validate(
			test.alg,
			[]byte(test.msg),
			[]byte(test.key),
			test.sig,
		)
		if res != test.res {
			t.Errorf("Want valid %v for message %q",
				test.res, test.msg)
		}
	}
}
