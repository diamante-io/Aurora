ENTRIES=(accounts accountdata offers trustlines)

echo "Sorting dump-ledger-state output files..."
for i in "${ENTRIES[@]}"
do
  if test -f "${i}_sorted.csv"; then
    echo "Skipping, ${i}_sorted.csv exists (remove if out of date to sort again)"
    continue
  fi
  wc -l ${i}.csv
  sort -o ${i}_sorted.csv ${i}.csv
done

echo "Sorting hcnet-core output files..."
for i in "${ENTRIES[@]}"
do
  if test -f "${i}_core_sorted.csv"; then
    echo "Skipping, ${i}_core_sorted.csv exists (remove if out of date to sort again)"
    continue
  fi
  wc -l ${i}_core.csv
  sort -o ${i}_core_sorted.csv ${i}_core.csv
done

echo "Checking diffs..."
for i in "${ENTRIES[@]}"
do
  diff -q ${i}_core_sorted.csv ${i}_sorted.csv
  if [ "$?" -ne "0" ]
  then
    echo "ERROR: $i does NOT match";
  else
    echo "$i OK";
  fi
done
