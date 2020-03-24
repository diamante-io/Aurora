# Get state from hcnet-core DB, colums match CSV printer
echo "Fetching accounts from hcnet-core DB..."
psql -d core -t -A -F"," -c "select accountid, balance, seqnum, numsubentries, inflationdest, homedomain, thresholds, flags, COALESCE(buyingliabilities, 0), COALESCE(sellingliabilities, 0), signers from accounts" > accounts_core.csv
rm accounts_core_sorted.csv

echo "Fetching accountdata from hcnet-core DB..."
psql -d core -t -A -F"," -c "select accountid, dataname, datavalue from accountdata" > accountdata_core.csv
rm accountdata_core_sorted.csv

echo "Fetching offers from hcnet-core DB..."
psql -d core -t -A -F"," -c "select sellerid, offerid, sellingasset, buyingasset, amount, pricen, priced, flags from offers" > offers_core.csv
rm offers_core_sorted.csv

echo "Fetching trustlines from hcnet-core DB..."
psql -d core -t -A -F"," -c "select accountid, assettype, issuer, assetcode, tlimit, balance, flags, COALESCE(buyingliabilities, 0), COALESCE(sellingliabilities, 0) from trustlines" > trustlines_core.csv
rm trustlines_core_sorted.csv
