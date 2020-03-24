run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :scott,       :master, 100
create_account :andrew,      :master, 100

close_ledger #1

payment :andrew, :scott,  [:native, "10.00"]

close_ledger #2

payment :scott, :andrew, [:native, "10.00"]

close_ledger #3

payment :andrew, :scott,  [:native, "10.00"]
payment :andrew, :scott,  [:native, "10.00"]
payment :andrew, :scott,  [:native, "10.00"]
payment :andrew, :scott,  [:native, "10.00"]

close_ledger #4

payment :scott, :andrew, [:native, "10.00"]
payment :andrew, :scott,  [:native, "10.00"]

close_ledger #5

payment :andrew, :scott,  [:native, "10.00"]
payment :andrew, :scott,  [:native, "10.00"]
payment :andrew, :scott,  [:native, "10.00"]

close_ledger #6
