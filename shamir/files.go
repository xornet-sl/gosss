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

package shamir

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"strings"
)

const defaultBlockSize = 16384

// SplitFile splits inputFile to parts
func SplitFile(inputFilename, outDir, filePattern string, partsCount, threshold int, blockSize uint64, polynomPerByte bool) error {
	if blockSize == 0 {
		blockSize = defaultBlockSize
	}
	var (
		err       error
		partsBuf  [][]byte
		xCoords   []uint8
		bytesRead int
		inputFile *os.File
		outFiles  = make([]*os.File, partsCount)
		readBuf   = make([]byte, blockSize)
	)
	err = validateSplitArgs(partsCount, threshold)
	if err != nil {
		return err
	}
	if inputFilename == "" {
		inputFile = os.Stdin
	} else {
		inputFile, err = os.Open(inputFilename)
		if err != nil {
			return err
		}
		defer inputFile.Close()
	}
	partsBuf, xCoords, err = prepareSplit(blockSize, uint8(partsCount))
	if err != nil {
		return err
	}
	for partIdx := 0; partIdx < partsCount; partIdx++ {
		filename := fmt.Sprint(partIdx + 1)
		if len(filePattern) > 0 {
			filename = strings.ReplaceAll(filePattern, "%i", fmt.Sprint(partIdx+1))
			if filename == filePattern {
				return errors.New("You should use '%i' in filename pattern in order to distinguish different parts")
			}
		}
		outFiles[partIdx], err = os.OpenFile(path.Join(outDir, filename), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return err
		}
		defer outFiles[partIdx].Close()
		outFiles[partIdx].Write(xCoords[partIdx : partIdx+1])
	}
	for {
		bytesRead, err = inputFile.Read(readBuf)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		err = splitBlock(readBuf[:bytesRead], partsBuf, xCoords, uint8(threshold), polynomPerByte)
		if err != nil {
			return err
		}
		for partIdx := range partsBuf {
			_, err = outFiles[partIdx].Write(partsBuf[partIdx][:bytesRead])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CombineFiles searches and combines files
func CombineFiles(searchDir, filePattern, outputFilename string, blockSize uint64) (int, error) {
	if blockSize == 0 {
		blockSize = defaultBlockSize
	}
	var (
		xSamples        []uint8
		ySamplesBuf     []uint8
		partsBuf        [][]byte
		partsCombineBuf [][]byte
		outFile         *os.File
		buf             = make([]byte, blockSize)
		partFiles       = make([]*os.File, 0, 256)
	)
	if len(filePattern) > 0 && strings.ReplaceAll(filePattern, "%i", "") == filePattern {
		return 0, errors.New("You should use '%i' in filename pattern in order to distinguish different parts")
	}
	files, err := ioutil.ReadDir(searchDir)
	if err != nil {
		return 0, err
	}
	rxNumber := regexp.MustCompile("[0-9]+")
	if filePattern == "" {
		filePattern = "%i"
	}
	partLen := int64(-1)
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if rxNumber.ReplaceAllString(file.Name(), "%i") != filePattern {
			continue
		}
		partFile, err := os.Open(path.Join(searchDir, file.Name()))
		if err != nil {
			return len(partFiles), err
		}
		defer partFile.Close()
		partStat, err := partFile.Stat()
		if err != nil {
			return len(partFiles), err
		}
		partFiles = append(partFiles, partFile)
		// dat, err := ioutil.ReadFile(path.Join(searchDir, file.Name()))
		// if err != nil {
		// 	return 0, err
		// }
		if partLen < 0 {
			partLen = partStat.Size()
		} else if partLen != partStat.Size() {
			return 0, errors.New("Parts have different size")
		}
	}
	partsCount := len(partFiles)
	if outputFilename == "" {
		outFile = os.Stdout
	} else {
		outFile, err = os.OpenFile(outputFilename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return partsCount, err
		}
		defer outFile.Close()
	}
	xSamples = make([]uint8, partsCount)
	ySamplesBuf = make([]uint8, partsCount)
	partsBuf = make([][]byte, partsCount)
	partsCombineBuf = make([][]byte, partsCount)
	for partIdx, partFile := range partFiles {
		partsBuf[partIdx] = make([]byte, blockSize)
		_, err = partFile.Read(xSamples[partIdx : partIdx+1])
		if err != nil {
			return partsCount, err
		}
	}
	for {
		currentRead := -1
		bytesRead := -1
		for partIdx, partFile := range partFiles {
			bytesRead, err = partFile.Read(partsBuf[partIdx])
			if err == io.EOF {
				if partIdx != 0 {
					return partsCount, errors.New("Unexpected EOF found in part file")
				}
				break
			} else if err != nil {
				return partsCount, err
			}
			if currentRead < 0 {
				currentRead = bytesRead
			} else if currentRead != bytesRead {
				return partsCount, errors.New("Unable to read parts evenly")
			}
			partsCombineBuf[partIdx] = partsBuf[partIdx][:bytesRead]
		}
		if err == io.EOF {
			break
		}
		err = combineBlock(buf, partsCombineBuf, xSamples, ySamplesBuf)
		if err != nil {
			return partsCount, err
		}
		_, err = outFile.Write(buf[:currentRead])
		if err != nil {
			return partsCount, err
		}
	}
	return partsCount, nil
}
