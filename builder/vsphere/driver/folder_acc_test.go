// Copyright IBM Corp. 2013, 2025
// SPDX-License-Identifier: MPL-2.0

package driver

import "testing"

func TestFolderAcc(t *testing.T) {
	t.Skip("Acceptance tests not configured yet.")
	d := newTestDriver(t)
	f, err := d.FindFolder("folder1/folder2")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	path, err := f.Path()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	if path != "folder1/folder2" {
		t.Errorf("unexpected result: expected '%s', but returned '%s'", "folder1/folder2", path)
	}
}
