The vanilla-mcdoc directory contains a nested set of schemas for Minecraft datapack objects in the "mcdoc" format, specified https://spyglassmc.com/user/mcdoc/specification.html

These schemas correspond to the requirements for json files in Minecraft datapacks, and have version info in them such as `#[until="1.19.1"]` which means that a particular atribute or feature is only valid for versions 1.19.1 and earlier.

I want you to create a command line tool in Go that is capable of checking a json file against a given schema. The tool should:

* parse the required `mcdata` files in order to perform validation
* validate against a user-specified minecraft version
* determine which kind of datapack item it is based on the current directory path, eg a file in `worldgen/noise_settings` should be validated against the `data/worldgen/noise_settings.mcdoc` file
