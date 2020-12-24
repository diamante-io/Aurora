#address: GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN
account :gateway, DiamNet::KeyPair.from_seed("SDHZO6NSO3OXIOORMZF4CAMYM37OK7E2OB3JCHT2ZD273ELK2QJRVNDR")
#address: GARSFJNXJIHO6ULUBK3DBYKVSIZE7SC72S5DYBCHU7DKL22UXKVD7MXP
account :payer, DiamNet::KeyPair.from_seed("SC2RRVCKKDT5HTUVJLHB4YCY4BXRI35S2R2XC2X3S3NPR46ITKB5E4G7")
#address: GAEDTJ4PPEFVW5XV2S7LUXBEHNQMX5Q2GM562RJGOQG7GVCE5H3HIB4V
account :payee, DiamNet::KeyPair.from_seed("SBEJ6AEIE2374O3YFBSVV7XI7QAWEKBYBWFGHHLMMPRH7W2O2NVB5NGU")
#address: GA2NC4ZOXMXLVQAQQ5IQKJX47M3PKBQV2N5UV5Z4OXLQJ3CKMBA2O2YL
account :trader, DiamNet::KeyPair.from_seed("SAAHOOTVIZJVXOEPCTNKYTYOTKZA3MFXJ3AIWEVCL4EH4HOPQBBOUTAA")

use_manual_close

create_account :gateway, :master, 10000
create_account :payer
create_account :payee
create_account :trader

close_ledger

trust :payer, :gateway, "USD"
trust :payee, :gateway, "EUR"

trust :payer, :gateway, "AAA"
trust :payee, :gateway, "CCC"

trust :trader, :gateway, "USD"
trust :trader, :gateway, "EUR"

# one hop path
trust :trader, :gateway, "1"
# two hop path
trust :trader, :gateway, "21"
trust :trader, :gateway, "22"

# three hop path
trust :trader, :gateway, "31"
trust :trader, :gateway, "32"
trust :trader, :gateway, "33"

# order test
trust :trader, :gateway, "AAA"
trust :trader, :gateway, "BBB"
trust :trader, :gateway, "CCC"

close_ledger

payment :gateway, :payer,   ["USD", :gateway, 5000]
payment :gateway, :trader,  ["EUR", :gateway, 5000]
payment :gateway, :trader,  ["1",   :gateway, 5000]
payment :gateway, :trader,  ["21", :gateway, 5000]
payment :gateway, :trader,  ["22", :gateway, 5000]
payment :gateway, :trader,  ["31", :gateway, 5000]
payment :gateway, :trader,  ["32", :gateway, 5000]
payment :gateway, :trader,  ["33", :gateway, 5000]
payment :gateway, :trader,  ["AAA", :gateway, 5000]
payment :gateway, :trader,  ["BBB", :gateway, 5000]
payment :gateway, :trader,  ["CCC", :gateway, 5000]

close_ledger

offer :trader, {for:["USD", :gateway], sell:["EUR", :gateway]}, 10, 0.5
offer :gateway, {for:["USD", :gateway], sell:["EUR", :gateway]}, 10, 1.0
offer :gateway, {for:["USD", :gateway], sell:["EUR", :gateway]}, 10, 0.5

offer :trader, {for:["USD", :gateway], sell:["1", :gateway]}, 20, 1.0
offer :trader, {for:["1", :gateway], sell:["EUR", :gateway]}, 20, 1.0

offer :trader, {for:["USD", :gateway], sell:["21", :gateway]}, 30, 1.0
offer :trader, {for:["21", :gateway], sell:["22", :gateway]}, 30, 1.0
offer :trader, {for:["22", :gateway], sell:["EUR", :gateway]}, 30, 1.0

offer :trader, {for:["USD", :gateway], sell:["31", :gateway]}, 40, 2.0
offer :trader, {for:["31", :gateway], sell:["32", :gateway]}, 40, 2.0
offer :trader, {for:["32", :gateway], sell:["33", :gateway]}, 40, 2.0
offer :trader, {for:["33", :gateway], sell:["EUR", :gateway]}, 40, 2.0

offer :gateway, {for:["USD", :gateway], sell: :native}, 1000, 0.1

offer :trader, {for:["AAA", :gateway], sell:["BBB", :gateway]}, 1, 11
offer :trader, {for:["BBB", :gateway], sell:["CCC", :gateway]}, 10, 0.1

close_ledger
