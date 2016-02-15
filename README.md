# gobundle

The goal of this project is to learn more about Go by implementing a
Javascript module loader for the web.

This tool can be used to generate a single JavaScript bundle containing all
dependencies installed using `npm` similar to [Browserify](http://browserify.org/).

## Usage

```
Usage:
  gobundle <entry_file> [-o <file>|--output=<file>]
  gobundle (-h | --help)
  gobundle --version
Options:
  <entry_file>                Entry file.
  -o <file> --output=<file>   Output file.
  -h --help                   Show this screen.
  --version                   Show version.
```
