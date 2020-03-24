run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway,  :master, 100
create_account :scott,  :master, 100

close_ledger

trust :scott,  :usd_gateway, "USD"

close_ledger

payment :usd_gateway, :scott, ["USD", :usd_gateway, 10_000_000]

close_ledger

payment :scott, :usd_gateway, ["USD", :usd_gateway, 10_000]

close_ledger
