// Copyright 2017 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gitlab

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/jenkins-x/go-scm/scm"

	"github.com/google/go-cmp/cmp"
	"github.com/h2non/gock"
)

func TestContentFind(t *testing.T) {
	defer gock.Off()

	gock.New("https://gitlab.com").
		Get("/api/v4/projects/diaspora/diaspora/repository/files/app/models/key.rb").
		MatchParam("ref", "7fd1a60b01f91b314f59955a4e4d4e80d8edf11d").
		Reply(200).
		Type("application/json").
		SetHeaders(mockHeaders).
		File("testdata/content.json")

	client := NewDefault()
	got, res, err := client.Contents.Find(
		context.Background(),
		"diaspora/diaspora",
		"app/models/key.rb",
		"7fd1a60b01f91b314f59955a4e4d4e80d8edf11d",
	)
	if err != nil {
		t.Error(err)
		return
	}

	want := new(scm.Content)
	raw, _ := ioutil.ReadFile("testdata/content.json.golden")
	json.Unmarshal(raw, want)

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("Unexpected Results")
		t.Log(diff)
	}

	t.Run("Request", testRequest(res))
	t.Run("Rate", testRate(res))
}

func TestContentCreate(t *testing.T) {
	content := new(contentService)
	_, err := content.Create(context.Background(), "octocat/hello-world", "README", nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestContentUpdate(t *testing.T) {
	content := new(contentService)
	_, err := content.Update(context.Background(), "octocat/hello-world", "README", nil)
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

func TestContentDelete(t *testing.T) {
	content := new(contentService)
	_, err := content.Delete(context.Background(), "octocat/hello-world", "README", "master")
	if err != scm.ErrNotSupported {
		t.Errorf("Expect Not Supported error")
	}
}

var fileContent = []byte(`require 'digest/md5'

class Key < ActiveRecord::Base
  include Gitlab::CurrentSettings
  include Sortable

  belongs_to :user

  before_validation :generate_fingerprint

  validates :title,
    presence: true,
    length: { maximum: 255 }

  validates :key,
    presence: true,
    length: { maximum: 5000 },
    format: { with: /\A(ssh|ecdsa)-.*\Z/ }

  validates :fingerprint,
    uniqueness: true,
    presence: { message: 'cannot be generated' }

  validate :key_meets_restrictions

  delegate :name, :email, to: :user, prefix: true

  after_commit :add_to_shell, on: :create
  after_create :post_create_hook
  after_create :refresh_user_cache
  after_commit :remove_from_shell, on: :destroy
  after_destroy :post_destroy_hook
  after_destroy :refresh_user_cache

  def key=(value)
    value&.delete!("\n\r")
    value.strip! unless value.blank`)
