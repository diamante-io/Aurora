run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway, :master, 100
create_account :eur_gateway, :master, 100
create_account :andrew, :master, 100
create_account :bartek, :master, 100
create_account :scott, :master, 100

close_ledger

trust :scott,  :usd_gateway, "USD"
trust :bartek, :eur_gateway, "EUR"
trust :andrew, :usd_gateway, "USD"
trust :andrew, :eur_gateway, "EUR"

close_ledger

payment :usd_gateway, :scott,  ["USD", :usd_gateway, 100]
payment :usd_gateway, :andrew, ["USD", :usd_gateway, 20]
payment :eur_gateway, :andrew, ["EUR", :eur_gateway, 20]
payment :eur_gateway, :bartek, ["EUR", :eur_gateway, 100]

close_ledger

offer :andrew, {buy:["USD", :usd_gateway], with:["EUR", :eur_gateway]}, 20, 1.0

close_ledger

# fix path pathment
payment :scott, :bartek, ["EUR", :eur_gateway, 10], with: ["USD", :usd_gateway, 10], path:[]

close_ledger
