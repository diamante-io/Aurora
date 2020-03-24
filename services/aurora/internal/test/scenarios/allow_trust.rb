run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway
create_account :scott
create_account :andrew

close_ledger

require_trust_auth :usd_gateway
set_flags :usd_gateway, [:auth_revocable_flag]

close_ledger

trust :scott,  :usd_gateway, "USD"              ; close_ledger
change_trust :andrew, :usd_gateway, "USD", 4000 ; close_ledger
allow_trust :usd_gateway, :scott, "USD"         ; close_ledger
allow_trust :usd_gateway, :andrew, "USD"        ; close_ledger
revoke_trust :usd_gateway, :andrew, "USD"       ; close_ledger

close_ledger
