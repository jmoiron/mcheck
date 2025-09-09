## mcheck & tsmc

This repo contains a command line tool that uses `SpyglassMC`'s "mcdoc" format to verify mincraft datapacks. It requires
a copy of the vanilla-mcdocs to run. You can get a copy by either:

```sh
# 1. Cloning the SpyglassMC/vanilla-mcdoc github repo into this directory
$ git clone https://github.com/SpyglassMC/vanilla-mcdoc

# 2. Using the SpyglassMC API to download a tarball of the mcdocs
$ curl -o mcdoc.tar https://api.spyglassmc.com/vanilla-mcdoc/tarball
```

### Where is it?

The `go-version` directory contains an attempt to generate a clean-room parser for mcdocs in Go using a custom peg parser.
One of the goals of pursuing this project was as a way to learn how to use LLM development tools, specifically CLI tools like
claude code and codex-cli.

There are two Go versions within the git history, one where claude tried to do this all with regexp, and one where I made it
use peg. The second version is what is on `HEAD`. I abandoned it when it became clear that utilizing mcdocs correctly
required a lot more secondary code than a parser; it required parsing an entire collection of mcdocs, doing type resolution,
etc.

At the point of abandonment, I decided to see if I could make a version that used the spyglass vscode extension's code
directly to do perform datapack validation. From here, leaning on LLMs was a must for me as I have no typescript experience.
This version is the TypeScript MCheck, or `tsmc`, a name that thankfully has no clashes with the world's most important
semiconductor fab.

## tsmc

The directory `tsmc` contains a tool, also called `tsmc`, which is documented fully in its own README.md.

`tsmc` works enough for me to use it as a quick CI/test script for local datapack development. My main personal goal was to
be able to have some guardrails for a local LLM to iterate on performing updates and making modifications to the datapack for
my modpack.

### Slop Warning

`tsmc` has not known the love of human touch. Its code is probably terrible. It might not really work properly. It might not be
suitable as a starting point for a fork to make a real CLI.
