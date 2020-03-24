package processors

import (
	"context"
	"encoding/base64"
	"encoding/csv"
	"fmt"
	stdio "io"
	"os"
	"strconv"

	"github.com/hcnet/go/exp/ingest/io"
	ingestpipeline "github.com/hcnet/go/exp/ingest/pipeline"
	"github.com/hcnet/go/exp/support/pipeline"
	"github.com/hcnet/go/support/errors"
	"github.com/hcnet/go/xdr"
)

func (p *CSVPrinter) fileHandle() (*os.File, error) {
	if p.Filename == "" {
		return os.Stdout, nil
	}

	return os.Create(p.Filename)
}

func (p *CSVPrinter) ProcessState(ctx context.Context, store *pipeline.Store, r io.StateReader, w io.StateWriter) error {
	defer r.Close()
	defer w.Close()

	f, err := p.fileHandle()
	if err != nil {
		return err
	}

	if f != os.Stdout {
		defer f.Close()
	}

	csvWriter := csv.NewWriter(f)

	for {
		entryChange, err := r.Read()
		if err != nil {
			if err == stdio.EOF {
				break
			} else {
				return err
			}
		}

		if entryChange.Type != xdr.LedgerEntryChangeTypeLedgerEntryState {
			return errors.New("CSVPrinter requires LedgerEntryChangeTypeLedgerEntryState changes only")
		}

		entry := entryChange.State
		switch entry.Data.Type {
		case xdr.LedgerEntryTypeAccount:
			account := entry.Data.MustAccount()

			inflationDest := ""
			if account.InflationDest != nil {
				inflationDest = entry.Data.Account.InflationDest.Address()
			}

			var err error
			var signers string
			if len(account.Signers) > 0 {
				signers, err = xdr.MarshalBase64(account.Signers)
				if err != nil {
					return err
				}
			}

			var buyingLiabilities, sellingLiabilities int64

			if account.Ext.V1 != nil {
				buyingLiabilities = int64(account.Ext.V1.Liabilities.Buying)
				sellingLiabilities = int64(account.Ext.V1.Liabilities.Selling)
			}

			csvWriter.Write([]string{
				account.AccountId.Address(),
				strconv.FormatInt(int64(account.Balance), 10),
				strconv.FormatInt(int64(account.SeqNum), 10),
				strconv.FormatInt(int64(account.NumSubEntries), 10),
				inflationDest,
				base64.StdEncoding.EncodeToString([]byte(account.HomeDomain)),
				base64.StdEncoding.EncodeToString(account.Thresholds[:]),
				strconv.FormatInt(int64(account.Flags), 10),
				strconv.FormatInt(buyingLiabilities, 10),
				strconv.FormatInt(sellingLiabilities, 10),
				signers,
			})
		case xdr.LedgerEntryTypeTrustline:
			trustline := entry.Data.MustTrustLine()

			var assetType, assetCode, assetIssuer string
			trustline.Asset.MustExtract(&assetType, &assetCode, &assetIssuer)

			var buyingLiabilities, sellingLiabilities int64

			if trustline.Ext.V1 != nil {
				buyingLiabilities = int64(trustline.Ext.V1.Liabilities.Buying)
				sellingLiabilities = int64(trustline.Ext.V1.Liabilities.Selling)
			}

			csvWriter.Write([]string{
				trustline.AccountId.Address(),
				strconv.FormatInt(int64(trustline.Asset.Type), 10),
				assetIssuer,
				assetCode,
				strconv.FormatInt(int64(trustline.Limit), 10),
				strconv.FormatInt(int64(trustline.Balance), 10),
				strconv.FormatInt(int64(trustline.Flags), 10),
				strconv.FormatInt(buyingLiabilities, 10),
				strconv.FormatInt(sellingLiabilities, 10),
			})
		case xdr.LedgerEntryTypeOffer:
			offer := entry.Data.MustOffer()

			var err error
			selling, err := xdr.MarshalBase64(offer.Selling)
			if err != nil {
				return err
			}

			buying, err := xdr.MarshalBase64(offer.Buying)
			if err != nil {
				return err
			}

			csvWriter.Write([]string{
				offer.SellerId.Address(),
				strconv.FormatInt(int64(offer.OfferId), 10),
				selling,
				buying,
				strconv.FormatInt(int64(offer.Amount), 10),
				strconv.FormatInt(int64(offer.Price.N), 10),
				strconv.FormatInt(int64(offer.Price.D), 10),
				strconv.FormatInt(int64(offer.Flags), 10),
			})
		case xdr.LedgerEntryTypeData:
			csvWriter.Write([]string{
				entry.Data.Data.AccountId.Address(),
				base64.StdEncoding.EncodeToString([]byte(entry.Data.Data.DataName)),
				base64.StdEncoding.EncodeToString(entry.Data.Data.DataValue),
			})
		default:
			return fmt.Errorf("Invalid LedgerEntryType: %d", entryChange.Type)
		}

		err = csvWriter.Error()
		if err != nil {
			return errors.Wrap(err, "Error during csv.Writer.Write")
		}

		csvWriter.Flush()
		err = csvWriter.Error()
		if err != nil {
			return errors.Wrap(err, "Error during csv.Writer.Flush")
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			continue
		}
	}

	return nil
}

func (p *CSVPrinter) ProcessLedger(ctx context.Context, store *pipeline.Store, r io.LedgerReader, w io.LedgerWriter) error {
	defer r.Close()
	defer w.Close()

	f, err := p.fileHandle()
	if err != nil {
		return err
	}

	if f != os.Stdout {
		defer f.Close()
	}

	for {
		transaction, err := r.Read()
		if err != nil {
			if err == stdio.EOF {
				break
			} else {
				return err
			}
		}

		fmt.Fprintf(
			f,
			"%d,%t,%s,%d,%d\n",
			transaction.Index,
			transaction.Result.Result.Result.Code == xdr.TransactionResultCodeTxSuccess,
			transaction.Envelope.Tx.SourceAccount.Address(),
			len(transaction.Envelope.Tx.Operations),
			transaction.Envelope.Tx.Fee,
		)

		select {
		case <-ctx.Done():
			return nil
		default:
			continue
		}
	}

	return nil
}

func (p *CSVPrinter) Name() string {
	return "CSVPrinter"
}

var _ ingestpipeline.StateProcessor = &CSVPrinter{}
var _ ingestpipeline.LedgerProcessor = &CSVPrinter{}
