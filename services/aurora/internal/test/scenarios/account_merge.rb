use_manual_close

run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"


create_account :scott,  :master
create_account :bartek, :master

close_ledger

merge_account :scott, :bartek

close_ledger
