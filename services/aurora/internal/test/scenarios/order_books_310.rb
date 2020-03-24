# This scenario attempts to emulate the scenario that gave rise to https://github.com/hcnet/aurora/issues/310

run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway
create_account :scott, :master, "1000000.0"
create_account :andrew, :master, "6000.0"

close_ledger

trust :scott,  :usd_gateway, "USD"
trust :andrew, :usd_gateway, "USD"

trust :scott,  :usd_gateway, "BTC"
trust :andrew, :usd_gateway, "BTC"

close_ledger

payment :usd_gateway, :scott,   ["USD", :usd_gateway, "5000.0"]
payment :usd_gateway, :andrew,  ["USD", :usd_gateway, "5000.0"]

payment :usd_gateway, :scott,   ["BTC", :usd_gateway, "5000.0"]
payment :usd_gateway, :andrew,  ["BTC", :usd_gateway, "5000.0"]

close_ledger


# polate an order book with more than 20 price levels, inserted into the db out
# of order.
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "10.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "9.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "8.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "7.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "6.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "5.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "4.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "3.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "2.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "1.0"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "0.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "10.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "9.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "8.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "7.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "6.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "5.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "4.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "3.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "2.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "1.1"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "0.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "10.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "9.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "8.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "7.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "6.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "5.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "4.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "3.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "2.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "1.2"
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "0.3"

close_ledger
