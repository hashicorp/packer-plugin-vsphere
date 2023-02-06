// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import "testing"

func TestFolderAcc(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	d := newTestDriver(t)
	f, err := d.FindFolder("folder1/folder2")
	if err != nil {
		t.Fatalf("Cannot find the default folder '%v': %v", "folder1/folder2", err)
	}
	path, err := f.Path()
	if err != nil {
		t.Fatalf("Cannot read folder name: %v", err)
	}
	if path != "folder1/folder2" {
		t.Errorf("Wrong folder. expected: 'folder1/folder2', got: '%v'", path)
	}
}
