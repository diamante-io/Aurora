use_manual_close

# Secret seed: SBZWG33UOQQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAPSA
# Address: GCXKG6RN4ONIEPCMNFB732A436Z5PNDSRLGWK7GBLCMQLIFO4S7EYWVU
account :anchor, HcNet::KeyPair.from_seed("SBZWG33UOQQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAPSA")
# Secret seed: SBQW4ZDSMV3SAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCA65I
# Address: GBXGQJWVLWOYHFLVTKWV5FGHA3LNYY2JQKM7OAJAUEQFU6LPCSEFVXON
account :user1, HcNet::KeyPair.from_seed("SBQW4ZDSMV3SAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCA65I")
# Secret seed: SBRGC4TUMVVSAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCBDHV
# Address: GA5WBPYA5Y4WAEHXWR2UKO2UO4BUGHUQ74EUPKON2QHV4WRHOIRNKKH2
account :user2, HcNet::KeyPair.from_seed("SBRGC4TUMVVSAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCAIBAEAQCBDHV")

create_account :anchor,  :master
create_account :user1, :master
create_account :user2, :master

close_ledger

trust :user1, :anchor, "USD"
trust :user2, :anchor, "USD"

close_ledger

payment :anchor, :user1,  ["USD", :anchor, "100.00"]
payment :anchor, :user2,  ["USD", :anchor, "100.00"]
offer :anchor, {buy:["USD", :anchor], with: :native}, "200.0", "2.0"

close_ledger

# this should fail
payment :user1, :user2,
  ["USD", :anchor, "200.0"],
  path:[:native]

close_ledger
