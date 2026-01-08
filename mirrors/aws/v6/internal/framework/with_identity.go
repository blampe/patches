// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	inttypes "github.com/blampe/patches/mirrors/aws/v6/internal/types"
)

type Identityer interface {
	SetIdentitySpec(identity inttypes.Identity)
}

var _ Identityer = &WithIdentity{}

type WithIdentity struct {
	identity inttypes.Identity
}

func (w *WithIdentity) SetIdentitySpec(identity inttypes.Identity) {
	w.identity = identity
}

func (w WithIdentity) IdentitySpec() inttypes.Identity {
	return w.identity
}
