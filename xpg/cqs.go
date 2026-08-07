package xpg

import (
	"github.com/sptGabriel/gb-libs/cqs"
)

// TransactionCommandInterceptor wraps the command handler execution in a
// database transaction. If no querier exists in the context, the
// transactioner falls back to its internal pool connection automatically.
//
// Deprecated: use cqs.TransactionCommandInterceptor. The body never had
// anything to do with Postgres — it only ever touched command types and a
// one-method interface. This stays as a thin forward so existing call sites
// keep compiling.
func TransactionCommandInterceptor[C cqs.Command](tx Tx) cqs.CommandInterceptor[C] {
	return cqs.TransactionCommandInterceptor[C](tx)
}
