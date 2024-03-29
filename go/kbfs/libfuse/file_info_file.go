// Copyright 2018 Keybase Inc. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.
//
//go:build !windows
// +build !windows

package libfuse

import (
	"time"

	"github.com/adamwalz/keybase-client/go/kbfs/libfs"
	"github.com/adamwalz/keybase-client/go/kbfs/libkbfs"
	"golang.org/x/net/context"
)

// NewFileInfoFile returns a special file that contains a text
// representation of a file's KBFS metadata.
func NewFileInfoFile(
	fs *FS, dir libkbfs.Node, name string,
	entryValid *time.Duration) *SpecialReadFile {
	*entryValid = 0
	return &SpecialReadFile{
		read: func(ctx context.Context) ([]byte, time.Time, error) {
			return libfs.GetFileInfo(ctx, fs.config, dir, name)
		},
	}
}
