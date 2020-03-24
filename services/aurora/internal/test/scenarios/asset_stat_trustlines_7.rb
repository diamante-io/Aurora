run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway
create_account :scott
create_account :bartek
close_ledger

trust :scott, :usd_gateway, "USD"
trust :bartek, :usd_gateway, "USD"
close_ledger

# issue asset to :scott
payment :usd_gateway, :scott,  ["USD", :usd_gateway, "101.2345"]
close_ledger

# :scott pays :bartek
payment :scott, :bartek,  ["USD", :usd_gateway, "10.123"]
close_ledger
