We have a working PEG parser for mcdoc using the pointlander/peg package.

Lets modify our validation tool to use this parser. This tool reads a json file and compares it to an mcdoc. The logic to determine which mcdoc file to use is already present and functional.

Using the parser's output, we will need to:

- a set of types similar to the expression types that are capable of validation, eg. a 'Range' type which validates that a numeric value falls within the range, or a 'Struct' validator that checks that attributes are allowed
- construct a tree of these types from the PEG parser's output (we can modify the grammar to do this during the parse step) to use for validation
- test the requested version against the version expressions in the mcdoc and only apply rules that match that version