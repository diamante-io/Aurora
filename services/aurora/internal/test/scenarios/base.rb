run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :scott,  :master, 100
create_account :bartek, :master, 100
create_account :andrew, :master, 100

close_ledger

payment :scott, :andrew,  [:native, 5]

close_ledger
