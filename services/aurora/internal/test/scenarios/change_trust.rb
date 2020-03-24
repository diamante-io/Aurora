run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway
create_account :scott

close_ledger

trust :scott, :usd_gateway, "USD"              ; close_ledger
change_trust :scott, :usd_gateway, "USD", 4000 ; close_ledger
change_trust :scott, :usd_gateway, "USD", 0    ; close_ledger

close_ledger
