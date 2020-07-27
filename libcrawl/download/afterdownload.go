/* This file is part of bbcrawl, ©2020 Jörg Walter
 *  This software is licensed under the "GNU General Public License version 3" */

package download

import "fmt"

//RenameError is thrown if renaming a downloaded file fails.
type RenameError struct {
	oldname *string
	newname *string
	err     error
}

func NewRenameError(name string, newname string, err error) RenameError {
	return RenameError{oldname: &name, newname: &newname, err: err}
}

func (e RenameError) Error() string {
	return fmt.Sprintf("Cannot rename file %q to %q", *e.oldname, *e.newname)
}

func (e RenameError) Unwrap() error {
	return e.err
}

func ADNameFromHeader(prefix string) func(*Download) {
	f := func(dl *Download) {
		var newname string
		name, err := dl.NameFromHeader()
		if err != nil {
			dl.Err = NewRenameError(dl.File(), name, err)
			return
		}
		if len(prefix) > 0 {
			newname = fmt.Sprintf("%s - %s", prefix, name)
		} else {
			newname = name
		}
		if err := dl.Rename(newname); err != nil {
			dl.Err = NewRenameError(dl.File(), newname, err)
			return
		}
	}
	return f
}
