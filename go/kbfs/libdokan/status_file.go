// Copyright 2016 Keybase Inc. All rights reserved.
// Use of this source code is governed by a BSD
// license that can be found in the LICENSE file.

package libdokan

import (
	"time"

	"github.com/adamwalz/keybase-client/go/kbfs/libfs"
	"golang.org/x/net/context"
)

// NewNonTLFStatusFile returns a special read file that contains a
// text representation of the global KBFS status.
func NewNonTLFStatusFile(fs *FS) *SpecialReadFile {
	return &SpecialReadFile{
		read: func(ctx context.Context) ([]byte, time.Time, error) {
			return libfs.GetEncodedStatus(ctx, fs.config)
		},
		fs: fs,
	}
}
