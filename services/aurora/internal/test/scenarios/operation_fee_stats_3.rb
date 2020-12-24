run_recipe File.dirname(__FILE__) + "/_common_accounts.rb"

use_manual_close
KP = DiamNet::KeyPair

close_ledger #1

account :multiop, KP.from_seed("SD7MLW2LH2PJOU5ZS2AVOZ5OWTH47MXYHOY6JSTNJU4RORU5RLNUTM7V")
create_account :multiop, :master

close_ledger #2

payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations * 2
    env.tx.fee = 200
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end

close_ledger #3

payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations
    env.tx.fee = 200
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end

close_ledger #4

payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations
    env.tx.fee = 400
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end

close_ledger #5

payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations
    env.tx.fee = 400
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end

close_ledger #6

payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations
    env.tx.fee = 300
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end

close_ledger #7

payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations
    env.tx.fee = 400
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end
payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations
    env.tx.fee = 400
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end
payment :multiop, :master,  [:native, "10.00"] do |env|
    env.tx.operations = env.tx.operations
    env.tx.fee = 400
    env.signatures = [env.tx.sign_decorated(get_account :multiop)]
  end


close_ledger #8
