// SPDX-License-Identifier: GPL-3.0-or-later

// Package closepool allows pooling [io.Closer] instances
// and closing them in a single operation.
//
// Deprecated: use `github.com/rbmk-project/common/closepool` instead.
package closepool

import "github.com/rbmk-project/common/closepool"

// Pool is an alias for [closepool.Pool].
type Pool = closepool.Pool
