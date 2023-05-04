# diamnet-key-gen

Generate Diamnet keys.

## Usage

Run the command with no options to get a public and private key:
```
diamnet-key-gen
GB2QRDI4FY2KERQBGPDS36XVWBJ4JBY3KW376H3KVF6YTNB2ROFNYN5L
SCGP6ZACCIPZXLGSMLNC3DE5VFZMS6GZJRCA4E524WFD5SHYQEE7NMK6
```

Run the command with a format option to change the output:
```
diamnet-key-gen -f '{{.SecretKey}}'
SCGP6ZACCIPZXLGSMLNC3DE5VFZMS6GZJRCA4E524WFD5SHYQEE7NMK6
```

Help:
```
$ diamnet-key-gen -h
Generate a Diamnet key.

Usage:
  diamnet-key-gen [flags]

Flags:
  -f, --format string   Format of output (default "{{.PublicKey}}\n{{.SecretKey}}\n")
```
