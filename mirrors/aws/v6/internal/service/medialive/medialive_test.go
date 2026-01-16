// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package medialive_test

import (
	"testing"

	"github.com/blampe/patches/mirrors/aws/v6/internal/acctest"
)

func TestAccMediaLive_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Multiplex": {
			acctest.CtBasic:      testAccMultiplex_basic,
			acctest.CtDisappears: testAccMultiplex_disappears,
			"update":             testAccMultiplex_update,
			"tags":               testAccMediaLiveMultiplex_tagsSerial,
			"start":              testAccMultiplex_start,
		},
		"MultiplexProgram": {
			acctest.CtBasic:      testAccMultiplexProgram_basic,
			"update":             testAccMultiplexProgram_update,
			acctest.CtDisappears: testAccMultiplexProgram_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
