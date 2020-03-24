// hcnet-sign is a small interactive utility to help you contribute a
// signature to a transaction envelope.
//
// It prompts you for a key
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/howeyc/gopass"
	"github.com/hcnet/go/build"
	"github.com/hcnet/go/xdr"
)

var in *bufio.Reader

var infile = flag.String("infile", "", "transaction envelope")

func main() {
	flag.Parse()
	in = bufio.NewReader(os.Stdin)

	var (
		env string
		err error
	)

	if *infile == "" {
		// read envelope
		env, err = readLine("Enter envelope (base64): ", false)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		var file *os.File
		file, err = os.Open(*infile)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		var raw []byte
		raw, err = ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}

		env = string(raw)
	}

	// parse the envelope
	var txe xdr.TransactionEnvelope
	err = xdr.SafeUnmarshalBase64(env, &txe)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("")
	fmt.Println("Transaction Summary:")
	fmt.Printf("  source: %s\n", txe.Tx.SourceAccount.Address())
	fmt.Printf("  ops: %d\n", len(txe.Tx.Operations))
	fmt.Printf("  sigs: %d\n", len(txe.Signatures))
	fmt.Println("")

	// TODO: add operation details

	// read seed
	seed, err := readLine("Enter seed: ", true)
	if err != nil {
		log.Fatal(err)
	}

	// sign the transaction
	b := &build.TransactionEnvelopeBuilder{E: &txe}
	b.Init()
	err = b.MutateTX(build.PublicNetwork)
	if err != nil {
		log.Fatal(err)
	}
	err = b.Mutate(build.Sign{seed})
	if err != nil {
		log.Fatal(err)
	}

	newEnv, err := xdr.MarshalBase64(b.E)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("\n==== Result ====\n\n")
	fmt.Print("```\n")
	fmt.Println(newEnv)
	fmt.Print("```\n")

}

func readLine(prompt string, private bool) (string, error) {
	fmt.Println(prompt)
	var line string
	var err error

	if private {
		str, err := gopass.GetPasswdMasked()
		if err != nil {
			return "", err
		}
		line = string(str)
	} else {
		line, err = in.ReadString('\n')
		if err != nil {
			return "", err
		}
	}
	return strings.Trim(line, "\n"), nil
}
