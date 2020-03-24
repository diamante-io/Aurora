run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

# create accounts
create_account :usd_gateway
create_account :scott
create_account :bartek
close_ledger

set_home_domain :usd_gateway, "test.com"
set_flags :usd_gateway, [:auth_required_flag]
set_flags :scott, [:auth_revocable_flag]
close_ledger

# add trustlines
trust :scott, :usd_gateway, "USD"
trust :scott, :usd_gateway, "BTC"
trust :bartek, :usd_gateway, "USD"
trust :bartek, :scott, "SCOT"
close_ledger

# allow trust
allow_trust :usd_gateway, :scott, "USD"
allow_trust :usd_gateway, :scott, "BTC"
allow_trust :usd_gateway, :bartek, "USD"
close_ledger

# issue currencies
payment :usd_gateway, :scott,  ["USD", :usd_gateway, "100000.12"]
payment :usd_gateway, :scott,  ["BTC", :usd_gateway, "100.9876"]
payment :usd_gateway, :bartek,  ["USD", :usd_gateway, "200000.9234"]
payment :scott, :bartek,  ["SCOT", :scott, "1000.00"]
close_ledger

# change trust
change_trust :bartek, :usd_gateway, "USD", 1000000000

# make payments 1
payment :scott, :bartek,  ["USD", :usd_gateway, "89.95"]
close_ledger

# make payments 2
payment :scott, :bartek,  ["USD", :usd_gateway, "31.657768"]
payment :bartek, :scott,  ["USD", :usd_gateway, "1.3623"]
close_ledger
