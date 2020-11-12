# gosss

gosss is a tool for splitting/combining files using shamir algorithm [More on wikipedia](https://en.wikipedia.org/wiki/Shamir%27s_Secret_Sharing)
This project can be used either as a tool or as a library
It allows to split and combine large files rather than tiny strings

## Idea

The initial idea has been taken form HashiCorp shamir splitter and somewhere rewritten  
The main goal is to add ability to split/combine large files by slightly modified block-aware algorithm

## build

`$ make`

### cross build
You can cross-compile gosss for different platforms using command `$ make build-cross`
it will create `_dist` folder with binaries for various platforms

## install

`$ sudo make install`

## using

`$ gosss --help`
