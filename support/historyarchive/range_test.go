// Copyright 2016 DiamNet Development Foundation and contributors. Licensed
// under the Apache License, Version 2.0. See the COPYING file at the root
// of this distribution or at http://www.apache.org/licenses/LICENSE-2.0

package historyarchive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFmtRangeList(t *testing.T) {

	assert.Equal(t,
		"",
		fmtRangeList([]uint32{}))

	assert.Equal(t,
		"0x0000003f",
		fmtRangeList([]uint32{0x3f}))

	assert.Equal(t,
		"[0x0000003f-0x0000007f]",
		fmtRangeList([]uint32{0x3f, 0x7f}))

	assert.Equal(t,
		"[0x0000003f-0x000000bf]",
		fmtRangeList([]uint32{0x3f, 0x7f, 0xbf}))

	assert.Equal(t,
		"[0x0000003f-0x0000007f], 0x000000ff",
		fmtRangeList([]uint32{0x3f, 0x7f, 0xff}))

	assert.Equal(t,
		"[0x0000003f-0x0000007f], [0x000000ff-0x0000017f]",
		fmtRangeList([]uint32{0x3f, 0x7f, 0xff, 0x13f, 0x17f}))

	assert.Equal(t,
		"[0x0000003f-0x0000007f], 0x000000ff, [0x0000017f-0x000001bf]",
		fmtRangeList([]uint32{0x3f, 0x7f, 0xff, 0x17f, 0x1bf}))
}
