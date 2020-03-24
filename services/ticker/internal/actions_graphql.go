package ticker

import (
	"github.com/hcnet/go/services/ticker/internal/gql"
	"github.com/hcnet/go/services/ticker/internal/tickerdb"
	hlog "github.com/hcnet/go/support/log"
)

func StartGraphQLServer(s *tickerdb.TickerSession, l *hlog.Entry, port string) {
	graphql := gql.New(s, l)

	graphql.Serve(port)
}
