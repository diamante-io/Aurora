run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :usd_gateway
create_account :scott, :master, "10000.0"
create_account :andrew, :master, "10000.0"

close_ledger

trust :scott,  :usd_gateway, "USD"
trust :andrew, :usd_gateway, "USD"

trust :scott,  :usd_gateway, "BTC"
trust :andrew, :usd_gateway, "BTC"

close_ledger

payment :usd_gateway, :scott,   ["USD", :usd_gateway, "5000.0"]
payment :usd_gateway, :andrew,  ["USD", :usd_gateway, "5000.0"]

payment :usd_gateway, :scott,   ["BTC", :usd_gateway, "6000.0"]
payment :usd_gateway, :andrew,  ["BTC", :usd_gateway, "6000.0"]

close_ledger

offer :scott, {buy:["USD", :usd_gateway], with: :native}, "10", "10.0" # price = 1/10; amount = 10*10 = 100
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "100", "9.0" # price = 1/9; amount = 100*9 = 900
offer :scott, {buy:["USD", :usd_gateway], with: :native}, "1000", "5.0" # price = 1/5; amount = 1000*5 = 5000

offer :andrew, {sell:["USD", :usd_gateway], for: :native}, "10", "15.0"
offer :andrew, {sell:["USD", :usd_gateway], for: :native}, "100", "20.0"
offer :andrew, {sell:["USD", :usd_gateway], for: :native}, "1000", "50.0"

close_ledger

offer :scott, {buy:["USD", :usd_gateway], with: ["BTC", :usd_gateway]}, 10, "10.0" # price = 1/10; amount = 10*10 = 100
offer :scott, {buy:["USD", :usd_gateway], with: ["BTC", :usd_gateway]}, 100, "9.0" # price = 1/9; amount = 100*9 = 900
offer :scott, {buy:["USD", :usd_gateway], with: ["BTC", :usd_gateway]}, 1000, "5.0" # price = 1/5; amount = 1000*5 = 5000

offer :andrew, {sell:["USD", :usd_gateway], for: ["BTC", :usd_gateway]}, 10, "15.0"
offer :andrew, {sell:["USD", :usd_gateway], for: ["BTC", :usd_gateway]}, 100, "20.0"
offer :andrew, {sell:["USD", :usd_gateway], for: ["BTC", :usd_gateway]}, 1000, "50.0"

close_ledger
