package main

import (
    "os"
    "log"
    "docopt"
    "gobundle/gobundle"
)

const Version = "0.1"
const Usage = `
Usage:
  gobundle <entry_file> [-o <file>|--output=<file>]
  gobundle (-h | --help)
  gobundle --version
Options:
  <entry_file>                Entry file.
  -o <file> --output=<file>   Output file.
  -h --help                   Show this screen.
  --version                   Show version.`

func main() {
    args, _ := docopt.Parse(Usage, nil, true, Version, false)

    entryFile := args["<entry_file>"].(string)

    outputFile, ok := args["--output"].(string)
    if !ok {
        outputFile = ""
    }

    writer := os.Stdout
    if len(outputFile) > 0 {
        fp, err := os.Create(outputFile)
        if err != nil {
            log.Fatalln(err)
        }
        defer fp.Close()
        writer = fp
    }

    bundle := gobundle.Bundle(entryFile)
    gobundle.WriteBundle(writer, bundle)
}
