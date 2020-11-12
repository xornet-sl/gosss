/*
Copyright © 2020 Vladimir Sukhonosov. Contacts: founder.sl@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	flags "github.com/jessevdk/go-flags"
)

type dirArguments struct {
	Pattern string `short:"p" long:"pattern" description:"filename pattern that is used to create or search parts. Use '%i' for part index replacement"`
	Dir     string `short:"d" long:"dir" description:"directory where parts are stored/searched. default is current dir"`
}

type commonArguments struct {
	BlockSize uint64 `short:"b" long:"block-size" description:"block size in bytes, default is 16k"`
}

type splitCommand struct {
	Dir            dirArguments
	Common         commonArguments
	In             string `short:"i" long:"in" description:"input secret filepath"`
	PartsCount     int    `short:"c" long:"count" description:"Parts count"`
	Threshold      int    `short:"t" long:"threshold" description:"Threshold for combining secret back"`
	PolynomPerByte bool   `short:"P" long:"polynom-per-byte" description:"generate new polynom for each byte"`
}

type combineCommand struct {
	Dir    dirArguments
	Common commonArguments
	Out    string `short:"o" long:"out" description:"output secret filepath"`
}

type versionCommand struct{}

// var dirArgs dirArguments
var splitCmd splitCommand
var combineCmd combineCommand
var versionCmd versionCommand

// ParseArgs parses command arguments and launches selected command
func parseArgs() {
	parser := flags.NewParser(nil, flags.PrintErrors|flags.HelpFlag)
	parser.AddCommand("version", "show version and exit", "", &versionCmd)
	parser.AddCommand("split", "split a secret", "split a secret to parts", &splitCmd)
	// cmd.AddGroup("Output parameters", "", &splitCmd.Dir)
	parser.AddCommand("combine", "combine parts", "combine splitted parts back to a secret", &combineCmd)
	// cmd.AddGroup("Input parameters", "", &combineCmd.Dir)
	parser.Parse()
}
