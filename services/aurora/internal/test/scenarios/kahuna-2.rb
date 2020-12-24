# This is the 2nd big kahuna test scenario.  It continues where the first big kahuna recipe left off `kahuna.rb` due to an unfortunate edge case.
#
use_manual_close
KP = DiamNet::KeyPair
close_ledger #2


# one-time signer being consumed

  # Public Key	GD6NTRJW5Z6NCWH4USWMNEYF77RUR2MTO6NP4KEDVJATTCUXDRO3YIFS
  # Secret Key	SAOCKUHEWRFMHSDBG33XFOUZDOXRZPKIXR3FUBS5QJ4IVKBGZDZPMLT3

  account :onetime, KP.from_seed("SAOCKUHEWRFMHSDBG33XFOUZDOXRZPKIXR3FUBS5QJ4IVKBGZDZPMLT3")
  create_account :onetime
  close_ledger #3

  # add one time signer

  require 'digest'
  x = Digest::SHA256.digest("hello world")
  key = DiamNet::SignerKey.hash_x(x)
  set_options :onetime, signer: DiamNet::Signer.new({
    key: key, 
    weight: 1,
  })
  close_ledger #4

  # consume one time signer
  account = get_account :onetime
  tx = DiamNet::Transaction.manage_data({
    account:  account,
    sequence: next_sequence(account),
    name:     "done",
    value:    "true",
  })
  env = tx.to_envelope
  env.signatures << DiamNet::DecoratedSignature.new({
    hint:      key.to_xdr.slice(-4, 4),
    signature: x
  })
  
  submit_transaction env

  close_ledger #5
