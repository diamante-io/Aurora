run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway
create_account :scott
close_ledger

trust :scott, :usd_gateway, "USD1"
trust :scott, :usd_gateway, "USD2"
close_ledger
