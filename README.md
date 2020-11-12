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
It will create `_dist` folder with binaries for various platforms

## install

`$ sudo make install`

## using

`$ gosss --help`

### Usage examples

Split file `test` to six parts with combine threshold 3:  
`$ gosss split -i test -d parts -p 'test-%i.part' -s 6 -t 3`

Combine parts back to one file:  
`$ gosss combine -d parts -p 'test-%i.part' -o combined`

Check that checksums are equal:  
`$ sha256sum test combined`

--input / --output args can be omitted or equal to '-'. In that case input/output file will be read/written from/to stdin/stdout:  
`$ gosss combine -d parts -p 'test-%i.part'` will output file to stdout

Combine tries to search all files matching the specified --pattern. Operation will be successful only if there are enough files found (more then or equal to threshold)  
Default --pattern is just `%i`. Default --dir is a current dir
