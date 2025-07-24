"mcdoc" is a format used to define schemas for minecraft data files. It is specified in the html file `docs/specification.html`.

The `vanilla-mcdocs` directory contains a nested set of mcdoc files describing various things, including minecraft datapack json files.

I want a command line tool written in Go that check whether a datapack json file is correct for a given version of Minecraft according to the schemas in the `vanilla-mcdocs` directory.

To build this tool, we need:

- a parser for the `mcdoc` format
- a way to match a json file to the appropriate schema
- a way to validate the json structure against the parsed schema

For the parser, we should create a PEG (parser expression grammar) parser. The `github.com/pointlander/peg` Go package can generate parsers in Go from peg files. The peg file format is described in `docs/peg-file-syntax.md` and there are example packages in `peg/grammars/`.

We can match json files to schema files using the directory context of the file. A json file in a directory like `worldgen/noise_settings` should use the `data/worldgen/noise_settings.mcdoc` schema.

There is already code in this directory that features an attempt to parse mcdoc files using regexp, but the regexp approach is not sophisticated enough for the mcdoc format.
