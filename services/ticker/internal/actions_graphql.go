package ticker

import (
	"github.com/diamnet/go/services/ticker/internal/gql"
	"github.com/diamnet/go/services/ticker/internal/tickerdb"
	hlog "github.com/diamnet/go/support/log"
)

func StartGraphQLServer(s *tickerdb.TickerSession, l *hlog.Entry, port string) {
	graphql := gql.New(s, l)

	graphql.Serve(port)
}
