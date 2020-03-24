run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :scott
create_account :bartek
create_account :andrew
create_account :usd_gateway
close_ledger

# :scott has no flags set
set_home_domain           :scott, "example.com"

set_flags                 :bartek, [:auth_required_flag]
set_home_domain           :bartek, "abc.com"

set_flags                 :andrew, [:auth_revocable_flag]
set_home_domain           :andrew, ""

set_flags                 :usd_gateway, [:auth_required_flag, :auth_revocable_flag]
# no home domain set

close_ledger
