run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close

create_account :scott
create_account :bartek

close_ledger
kp = DiamNet::KeyPair.from_seed("SB2XGZC7M5QXIZLXMF4SAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCBV6K")

set_inflation_dest :scott, :bartek                ; close_ledger
set_flags :scott, [:auth_required_flag]           ; close_ledger
set_master_signer_weight :scott, 2                ; close_ledger
set_thresholds :scott, low: 0, medium: 2, high: 2 ; close_ledger
set_home_domain :scott, "nullstyle.com"           ; close_ledger
add_signer :scott, kp, 1                          ; close_ledger
add_signer :scott, kp, 5                          ; close_ledger
clear_flags :scott, [:auth_required_flag]         ; close_ledger
remove_signer :scott, kp                          ; close_ledger

close_ledger
