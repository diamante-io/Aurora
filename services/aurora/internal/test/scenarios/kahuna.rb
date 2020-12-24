# This is the big kahuna test scenario.  It aims to comprehensively use
# diamnet's features, exercising each feature and option at least once.
#
# As new features are added to diamnet, this scenario will get updated to
# exercise those new features.  This scenario is used during the aurora
# ingestion tests.
#
use_manual_close
KP = DiamNet::KeyPair
close_ledger #2


## Transaction exercises

# time bounds
  # Secret seed: SBQGG7PY4JZQT6F2MBXDDI4VNDKZYG2Y5TJLKNG7AG6ETNNTJT6MCBOF
  # Public: GAXI33UCLQTCKM2NMRBS7XYBR535LLEVAHL5YBN4FTCB4HZHT7ZA5CVK
  account :time_bounds, KP.from_seed("SBQGG7PY4JZQT6F2MBXDDI4VNDKZYG2Y5TJLKNG7AG6ETNNTJT6MCBOF")
  create_account :time_bounds do |env|
    env.tx.time_bounds = DiamNet::TimeBounds.new(min_time: 100, max_time: Time.parse("2020-01-01").to_i)
    env.signatures = [env.tx.sign_decorated(get_account :master)]
  end

  close_ledger #3

# multisig
  # Secret seed: SCUNXR4FVJGNH4CZBA5K3IIK264WBH7FURTLDMOT3FGWNZKOZCPBW3GS
  # Public: GDXFAGJCSCI4CK2YHK6YRLA6TKEXFRX7BMGVMQOBMLIEUJRJ5YQNLMIB
  account :multisig, KP.from_seed("SCUNXR4FVJGNH4CZBA5K3IIK264WBH7FURTLDMOT3FGWNZKOZCPBW3GS")
  # Secret seed: SBYRIKN5UMKMLVMVJOV2TC2FI3VIU572W4V5KIKR7EDNNORPBIW2RO3I
  # Public: GD3E7HKMRNT6HGBGHBT6I6JE4N2S4W5KZ246TGJ4KQSXJ2P4BXCUPQMP
  multisig_2 = KP.from_seed("SBYRIKN5UMKMLVMVJOV2TC2FI3VIU572W4V5KIKR7EDNNORPBIW2RO3I")

  create_account :multisig
  close_ledger #4

  set_master_signer_weight :multisig, 1
  add_signer :multisig, multisig_2, 1
  set_thresholds :multisig, low: 2, medium: 2, high: 2

  close_ledger #5

  set_master_signer_weight :multisig, 2 do |env|
    env.signatures << env.tx.sign_decorated(multisig_2)
  end

  close_ledger #6

# memo
  # Secret seed: SACUGRIIBI3WTZTVYZNS7KM4JSJQJKUE2AQ4GDE3MEZRAT5QEJ76YWFE
  # Public: GA46VRKBCLI2X6DXLX7AIEVRFLH3UA7XBE3NGNP6O74HQ5LXHMGTV2JB
  account :memo, KP.from_seed("SACUGRIIBI3WTZTVYZNS7KM4JSJQJKUE2AQ4GDE3MEZRAT5QEJ76YWFE")
  create_account :memo
  close_ledger #7

  payment :memo, :master, [:native, "1.0"], memo: [:id, 123]
  payment :memo, :master, [:native, "1.0"], memo: [:text, "hello"]
  payment :memo, :master, [:native, "1.0"], memo: [:hash, "\x01" * 32]
  payment :memo, :master, [:native, "1.0"], memo: [:return, "\x02" * 32]
  close_ledger #8

# multiop
  # Secret seed: SD7MLW2LH2PJOU5ZS2AVOZ5OWTH47MXYHOY6JSTNJU4RORU5RLNUTM7V
  # Public: GAG52TW6QAB6TGNMOTL32Y4M3UQQLNNNHPEHYAIYRP6SFF6ZAVRF5ZQY
  account :multiop, KP.from_seed("SD7MLW2LH2PJOU5ZS2AVOZ5OWTH47MXYHOY6JSTNJU4RORU5RLNUTM7V")
  create_account :multiop
  close_ledger #9

  payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations * 2
    env.tx.fee = 200
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end

  close_ledger #10

## Operation exercises

# create account
  # Secret seed: SDM2YMHCCJWEOSDA26XV7OAKP4DPS5MZXBM7RHUR5N7XMKVDOMCQDINF
  # Public: GDCVTBGSEEU7KLXUMHMSXBALXJ2T4I2KOPXW2S5TRLKDRIAXD5UDHAYO
  account :first_create, KP.from_seed("SDM2YMHCCJWEOSDA26XV7OAKP4DPS5MZXBM7RHUR5N7XMKVDOMCQDINF")
  # Secret seed: SCHWJBPUXWXYXQ5GCTDSGZUQWNRX7IO5DDD3ULPTR5JQPK6AG7YOKMFF
  # Public: GCB7FPYGLL6RJ37HKRAYW5TAWMFBGGFGM4IM6ERBCZXI2BZ4OOOX2UAY
  account :second_create, KP.from_seed("SCHWJBPUXWXYXQ5GCTDSGZUQWNRX7IO5DDD3ULPTR5JQPK6AG7YOKMFF")

  # default create from root account
  create_account :first_create
  close_ledger #11

  # create with custom starting balance
  create_account :second_create, :first_create, "50.00"

  close_ledger #12

# payment
  # Secret seed: SCYYD7ZVS4UNOOIEQA2W77ZEXLOVTGQXA3Z6WCP3KD7YLMT3GFTTTMMO
  # Public: GCHC4D2CS45CJRNN4QAHT2LFZAJIU5PA7H53K3VOP6WEJ6XWHNSNZKQG
  account :payer, KP.from_seed("SCYYD7ZVS4UNOOIEQA2W77ZEXLOVTGQXA3Z6WCP3KD7YLMT3GFTTTMMO")
  # Secret seed: SAZRXWWS6BZ5G7TTW22CXDPJQC2PHYVDBQJBDADS4MHN4NRJQYPW7JFU
  # Public: GANZGPKY5WSHWG5YOZMNG52GCK5SCJ4YGUWMJJVGZSK2FP4BI2JIJN2C
  account :payee, KP.from_seed("SAZRXWWS6BZ5G7TTW22CXDPJQC2PHYVDBQJBDADS4MHN4NRJQYPW7JFU")

  create_account :payer
  create_account :payee
  close_ledger #13

  # native payment
  payment :payer, :payee,  [:native, "10.00"]

  # non-native payment
  trust :payee, :payer, "USD"
  close_ledger #14
  payment :payer, :payee,  ["USD", :payer, "10.00"]

  close_ledger #15

# path payment

  # Secret seed: SBVM6Q7LG23HQGK6P56RY4UMI24DB6DSYPH6QSBUMD7FM3YOAO4JTZOE
  # Public: GDRW375MAYR46ODGF2WGANQC2RRZL7O246DYHHCGWTV2RE7IHE2QUQLD
  account :path_payer, KP.from_seed("SBVM6Q7LG23HQGK6P56RY4UMI24DB6DSYPH6QSBUMD7FM3YOAO4JTZOE")
  # Secret seed: SCKTG6NCSP5JMZBXCI7UQUDB2X3UOFIKDHB4Z3RZ7ZP4UOIHBGEU6VDA
  # Public: GACAR2AEYEKITE2LKI5RMXF5MIVZ6Q7XILROGDT22O7JX4DSWFS7FDDP
  account :path_payee, KP.from_seed("SCKTG6NCSP5JMZBXCI7UQUDB2X3UOFIKDHB4Z3RZ7ZP4UOIHBGEU6VDA")
  # Secret seed: SB6E22ZX7QOUNZCINQGNXCFQXZNSLU3WTUIWDRFNXNNMYLIWQLX2IIMB
  # Public: GAXMF43TGZHW3QN3REOUA2U5PW5BTARXGGYJ3JIFHW3YT6QRKRL3CPPU
  account :path_gateway, KP.from_seed("SB6E22ZX7QOUNZCINQGNXCFQXZNSLU3WTUIWDRFNXNNMYLIWQLX2IIMB")

  create_account :path_payer
  create_account :path_payee
  create_account :path_gateway
  close_ledger #16

  trust :path_payer,  :path_gateway, "USD"
  trust :path_payee,  :path_gateway, "EUR"
  close_ledger #17

  payment :path_gateway, :path_payer,  ["USD", :path_gateway, "100.00"]
  offer :path_gateway, {buy:["USD", :path_gateway], with: :native}, "200.0", "2.0"
  offer :path_gateway, {sell:["EUR", :path_gateway], for: :native}, "300.0", "1.0"
  close_ledger #18

  payment :path_payer, :path_payee,
    ["EUR", :path_gateway, "200.0"],
    with: ["USD", :path_gateway, "100.0"],
    path:[:native]
  close_ledger #19

  payment :path_payer, :path_payee,
    ["EUR", :path_gateway, "100.0"],
    with: [:native, "100.0"],
    path:[]
  close_ledger #20

# manage offer

  # Secret seed: SDK24P4CD2ILEMNEDAWEZIJS6TZYTQQH7VQRPNVXXIJVYDI7IBUSUG2K
  # Public: GBOK7BOUSOWPHBANBYM6MIRYZJIDIPUYJPXHTHADF75UEVIVYWHHONQC
  account :manage_trader, KP.from_seed("SDK24P4CD2ILEMNEDAWEZIJS6TZYTQQH7VQRPNVXXIJVYDI7IBUSUG2K")
  # Secret seed: SCPIDWDRQCS2CNEXDQBJ7WLXJKG2D3WMPWZ74YXDHGBKX2BZZL625UE7
  # Public: GB2QIYT2IAUFMRXKLSLLPRECC6OCOGJMADSPTRK7TGNT2SFR2YGWDARD
  account :manage_gateway, KP.from_seed("SCPIDWDRQCS2CNEXDQBJ7WLXJKG2D3WMPWZ74YXDHGBKX2BZZL625UE7")
  create_account :manage_trader
  create_account :manage_gateway
  close_ledger #21

  trust :manage_trader,  :manage_gateway, "USD"
  close_ledger #22

  # make offer
  offer :manage_trader, {buy:["USD", :manage_gateway], with: :native}, "20.0", "1.0"
  close_ledger #23

  # offer that consumes another
  offer :manage_gateway, {sell:["USD", :manage_gateway], for: :native}, "30.0", "1.0"
  close_ledger #24

# create passive offer

  # Secret seed: SDLEIPR3JYEPVGISRN7WK5NRK667Q3HHXUYVBSSIIV46AHQ7QOEOP7XY
  # Public: GB6GN3LJUW6JYR7EDOJ47VBH7D45M4JWHXGK6LHJRAEI5JBSN2DBQY7Q
  account :passive_trader, KP.from_seed("SDLEIPR3JYEPVGISRN7WK5NRK667Q3HHXUYVBSSIIV46AHQ7QOEOP7XY")
  create_account :passive_trader
  close_ledger #25

  passive_offer :passive_trader, {sell:["USD", :passive_trader], for: :native}, "200.0", "1.0"
  passive_offer :passive_trader, {buy:["USD", :passive_trader], with: :native}, "200.0", "1.0"
  close_ledger #26


# set options
  # Secret seed: SDTS7HKJ4TON3G66U4PN2P6QPHMCFYDQHXDZ5BPCZSZQ62OGZDGPO3TX
  # Public: GCIFFRQKHMH6JD7CK5OI4XVCYCMNRNF6PYA7JTCR3FPHPJZQTYYFB5ES
  account :optioneer, KP.from_seed("SDTS7HKJ4TON3G66U4PN2P6QPHMCFYDQHXDZ5BPCZSZQ62OGZDGPO3TX")
  # Secret seed: SBLKCWLFDZUF2F7WK43DZFPBHXA3N33LCNJ2U7UDSWG6NIUVV35TTLIY
  # Public: GB6J3WOLKYQE6KVDZEA4JDMFTTONUYP3PUHNDNZRWIKA6JQWIMJZATFE
  option_kp = KP.from_seed("SBLKCWLFDZUF2F7WK43DZFPBHXA3N33LCNJ2U7UDSWG6NIUVV35TTLIY")

  create_account :optioneer
  close_ledger #27

  set_inflation_dest        :optioneer, :master
  set_flags                 :optioneer, [:auth_required_flag]
  set_flags                 :optioneer, [:auth_revocable_flag]
  set_master_signer_weight  :optioneer, 2
  set_thresholds            :optioneer, low: 0, medium: 2, high: 2
  set_home_domain           :optioneer, "example.com"
  add_signer                :optioneer, option_kp, 1
  close_ledger #28

  # no-op change of master weight
  set_master_signer_weight  :optioneer, 2
  close_ledger #29

  # no-op change of weight
  add_signer                :optioneer, option_kp, 1
  close_ledger #30

  # change weight
  add_signer                :optioneer, option_kp, 5
  close_ledger #31

  clear_flags               :optioneer, [:auth_required_flag, :auth_revocable_flag]
  remove_signer             :optioneer, option_kp
  close_ledger #32


# change trust
  # Secret seed: SCCNDKYXPFSHP3IPOQ2DGNJ6TKJY5KPRMB4JDMLISICHBMPMBF42ZEQU
  # Public: GCJKJXPKBFIHOO3455WXWG5CDBZXQNYFRRGICYMPUQ35CPQ4WVS3KZLG
  account :change_trustor, KP.from_seed("SCCNDKYXPFSHP3IPOQ2DGNJ6TKJY5KPRMB4JDMLISICHBMPMBF42ZEQU")
  create_account :change_trustor
  close_ledger #33


  trust :change_trustor,  :master, "USD"
  close_ledger #34

  change_trust :change_trustor,  :master, "USD", 100
  close_ledger #35

  # change trust operation that doesn't change limit
  change_trust :change_trustor,  :master, "USD", 100
  close_ledger #36

  change_trust :change_trustor,  :master, "USD", 0
  close_ledger #37


# allow trust

  # Secret seed: SCSV7463UPUPU5T2WJVT7TUK3V3IC4AMQTR6QSRAP2J2XS37JSEDY2J4
  # Public: GCVW5LCRZFP7PENXTAGOVIQXADDNUXXZJCNKF4VQB2IK7W2LPJWF73UG
  account :allow_trustor, KP.from_seed("SCSV7463UPUPU5T2WJVT7TUK3V3IC4AMQTR6QSRAP2J2XS37JSEDY2J4")
  # Secret seed: SAJTZHT2P73SI3U7VPYZTXVWRXKWUPTW2Q4GDUT56NTJ5PXAYUYNVQ3Q
  # Public: GD4SMOE3VPSF7ZR3CTEQ3P5UNTBMEJDA2GLXTHR7MMARANKKJDZ7RPGF
  account :allow_trustee, KP.from_seed("SAJTZHT2P73SI3U7VPYZTXVWRXKWUPTW2Q4GDUT56NTJ5PXAYUYNVQ3Q")
  create_account :allow_trustor
  create_account :allow_trustee
  close_ledger #38

  set_flags :allow_trustee, [:auth_required_flag, :auth_revocable_flag]
  close_ledger  #39

  # start trust
  trust :allow_trustor, :allow_trustee, "USD"
  trust :allow_trustor, :allow_trustee, "EUR"
  close_ledger  #40

  # allow trust
  allow_trust :allow_trustee, :allow_trustor, "USD"
  allow_trust :allow_trustee, :allow_trustor, "EUR"
  close_ledger  #41

  # revoke trust
  allow_trust :allow_trustee, :allow_trustor, "EUR", false
  close_ledger #42

# account merge
  # Secret seed: SCCLMTKRVHN2GSPJ7IP3VXI2NATH6QQTE5ZDMJCIZYWMZESSF5RKKBHT
  # Public: GCHPXGVDKPF5KT4CNAT7X77OXYZ7YVE4JHKFDUHCGCVWCL4K4PQ67KKZ
  account :merger, KP.from_seed("SCCLMTKRVHN2GSPJ7IP3VXI2NATH6QQTE5ZDMJCIZYWMZESSF5RKKBHT")
  create_account :merger
  close_ledger #43

  merge_account :merger, :master
  close_ledger #44

# inflation with payouts
# Secret seed: SBLDQ4ZOUCR4ZOIS4VHMSK5CZTZ7EXIPBNM2LP2HQPMW4T2F5EBTP4MF
# Public: GDR53WAEIKOU3ZKN34CSHAWH7HV6K63CBJRUTWUDBFSMY7RRQK3SPKOS
account :inflatee, KP.from_seed("SBLDQ4ZOUCR4ZOIS4VHMSK5CZTZ7EXIPBNM2LP2HQPMW4T2F5EBTP4MF")

create_account :inflatee,  :master, "20000000000.0"
close_ledger #45

set_inflation_dest :master, :master
set_inflation_dest :inflatee, :inflatee
close_ledger #46

inflation
close_ledger #47


# manage_data
  # Secret seed: SCHZL45S64JBNP7V6K7IM35PM7MFJ3REWRVMDRSJIH63JIYSW44VUOLN
  # Public: GAYSCMKQY6EYLXOPTT6JPPOXDMVNBWITPTSZIVWW4LWARVBOTH5RTLAD
  account :dataman, KP.from_seed("SCHZL45S64JBNP7V6K7IM35PM7MFJ3REWRVMDRSJIH63JIYSW44VUOLN")
  create_account :dataman
  close_ledger #48

  set_data :dataman, "name1", "1234"
  set_data :dataman, "name2", "5678"
  set_data :dataman, "name ", "its got spaces!"
  close_ledger #49

  clear_data :dataman, "name2"
  close_ledger #50

  # no-op change
  set_data :dataman, "name1", "1234"
  close_ledger #51

  set_data :dataman, "name1", "0000"
  close_ledger #52

# different source account
  # Secret seed: SAKHJAR6DZPSQMKR5EFGQBWH4RZYUCLXBXA3MMJ5PK7YWV2LKVWEQMYA
  # Public: GACJPE4YUR22VP4CM2BDFDAHY3DLEF3H7NENKUQ53DT5TEI2GAHT5N4X
  account :different_source, KP.from_seed("SAKHJAR6DZPSQMKR5EFGQBWH4RZYUCLXBXA3MMJ5PK7YWV2LKVWEQMYA")
  create_account :different_source
  close_ledger #53

  payment :master, :different_source,  [:native, "10.00"] do |env|
    newop = DiamNet::Operation.from_xdr env.tx.operations[0].to_xdr

    newop.source_account         = env.tx.operations[0].body.value.destination
    newop.body.value.destination = env.tx.source_account

    env.tx.operations << newop
    env.tx.fee = 200
    env.signatures = [
      env.tx.sign_decorated(get_account :master),
      env.tx.sign_decorated(get_account :different_source),
    ]
  end
  close_ledger #54

# self-pay
  # Secret seed: SAN5MUUVD2B3WJPIFDT7FQRLGNTD7LYFT7S7ULOKYBFC6ZUFIOSC2YRP
  # Public: GANFZDRBCNTUXIODCJEYMACPMCSZEVE4WZGZ3CZDZ3P2SXK4KH75IK6Y

  account :selfpay, KP.from_seed("SAN5MUUVD2B3WJPIFDT7FQRLGNTD7LYFT7S7ULOKYBFC6ZUFIOSC2YRP")
  create_account :selfpay
  close_ledger #55

  payment :selfpay, :selfpay, [:native, "10.0"]
  close_ledger #56

# bumpseq
  # Public Key	GCQZP3IU7XU6EJ63JZXKCQOYT2RNXN3HB5CNHENNUEUHSMA4VUJJJSEN
  # Secret Key	SB7M6CWQLVEAMCJMYNVURRVXQDBBEDEAYFXUSP5B4XCG4FRMALD7CWCG

  account :bumper, KP.from_seed("SB7M6CWQLVEAMCJMYNVURRVXQDBBEDEAYFXUSP5B4XCG4FRMALD7CWCG")
  create_account :bumper
  close_ledger #57

  bump_sequence :bumper, 300000000000 #should work
  close_ledger #58

  #all the following should ingest as operations but have no effects
  bump_sequence :bumper, 100
  close_ledger #59

  bump_sequence :bumper, next_sequence(get_account(:bumper))-1
  close_ledger #60

  bump_sequence :bumper, next_sequence(get_account(:bumper))
  close_ledger #61

# ====================================================================
# NOTE:  Do not add any additional commands to this recipe.  Due to an 
# issue, we can't easily go beyond 64 ledgers in a single recipe.
return
# ====================================================================
