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
	"fmt"

	"github.com/xornet-sl/gosss/shamir"
)

var version = "dev"
var commit = ""

func check(err error) {
	if err != nil {
		panic(err)
	}
}

// Split
func (x *splitCommand) Execute(args []string) error {
	inFile := x.In
	if inFile == "-" {
		inFile = ""
	}
	return shamir.SplitFile(x.In, x.Dir.Dir, x.Dir.Pattern, x.PartsCount, x.Threshold, x.Common.BlockSize, x.PolynomPerByte)
}

// Combine
func (x *combineCommand) Execute(args []string) error {
	dir := "."
	if len(x.Dir.Dir) > 0 {
		dir = x.Dir.Dir
	}
	outFile := x.Out
	if outFile == "-" {
		outFile = ""
	}
	_, err := shamir.CombineFiles(dir, x.Dir.Pattern, outFile, x.Common.BlockSize)
	// fmt.Printf("Found %d parts\n", foundParts)
	return err
}

// Version
func (x *versionCommand) Execute(args []string) error {
	s := fmt.Sprintf("gosss version %s (%s)", version, commit)
	fmt.Println(s)
	return nil
}

func main() {
	parseArgs()
}
