// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package enum

import (
	tfslices "github.com/blampe/patches/mirrors/aws/v6/internal/slices"
	inttypes "github.com/blampe/patches/mirrors/aws/v6/internal/types"
)

type Valueser[T ~string] interface {
	~string
	Values() []T
}

func EnumValues[T Valueser[T]]() []T {
	return inttypes.Zero[T]().Values()
}

func Values[T Valueser[T]]() []string {
	return tfslices.Strings(EnumValues[T]())
}

func EnumSlice[T ~string](l ...T) []T {
	return l
}

func Slice[T ~string](l ...T) []string {
	return tfslices.Strings(l)
}
