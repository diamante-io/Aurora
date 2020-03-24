run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :scott,       :master, 100
create_account :usd_gateway, :master, 100
create_account :andrew,      :master, 100

close_ledger

trust :scott,  :usd_gateway, "USD"
trust :andrew, :usd_gateway, "USD"

close_ledger

payment :usd_gateway, :scott,  ["USD", :usd_gateway, 100]

close_ledger

payment :scott, :andrew, ["USD", :usd_gateway, 50]

close_ledger
