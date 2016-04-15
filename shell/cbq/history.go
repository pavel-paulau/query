//  Copyright (c) 2015-2016 Couchbase, Inc.
//  Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file
//  except in compliance with the License. You may obtain a copy of the License at
//    http://www.apache.org/licenses/LICENSE-2.0
//  Unless required by applicable law or agreed to in writing, software distributed under the
//  License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
//  either express or implied. See the License for the specific language governing permissions
//  and limitations under the License.

package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/couchbase/query/errors"
	"github.com/couchbase/query/shell/cbq/command"
	"github.com/peterh/liner"
)

func LoadHistory(liner *liner.State, dir string) (int, string) {
	if dir != "" {

		path := command.GetPath(dir, command.HISTFILE)
		err_code, err_str := ReadHistoryFromFile(liner, path)

		if err_code != 0 {
			return err_code, err_str
		}
		//Print path to histfile on startup.
		if !command.QUIET {
			io.WriteString(command.W, "Path to stored history for the shell : "+path+" \n")
		}
	}
	return 0, ""
}

func UpdateHistory(liner *liner.State, dir, line string) (int, string) {
	liner.AppendHistory(line)
	if dir != "" {
		path := command.GetPath(dir, command.HISTFILE)
		err_code, err_str := WriteHistoryToFile(liner, path)

		if err_code != 0 {
			return err_code, err_str
		}
	}
	return 0, ""
}

func WriteHistoryToFile(liner *liner.State, path string) (int, string) {

	var err error
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return errors.FILE_OPEN, err.Error()
	}

	defer f.Close()

	writer := bufio.NewWriter(f)
	_, err = liner.WriteHistory(writer)
	if err != nil {
		return errors.WRITE_FILE, err.Error()
	} else {
		err = writer.Flush()
		if err != nil {
			return errors.WRITER_OUTPUT, err.Error()
		}
	}
	return 0, ""

}

func ReadHistoryFromFile(liner *liner.State, path string) (int, string) {

	var err error
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			f, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				return errors.FILE_OPEN, err.Error()
			}

		} else {
			return errors.FILE_OPEN, err.Error()
		}
	}

	defer f.Close()

	reader := bufio.NewReader(f)
	_, err = liner.ReadHistory(reader)

	//Check for line too long errors. If the line didnt fit into the buffer
	//then dont report the error
	if err != nil && strings.Contains(err.Error(), "too long") {
		err = nil
	}

	if err != nil {
		return errors.READ_FILE, err.Error()
	}

	return 0, ""
}
