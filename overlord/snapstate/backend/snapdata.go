// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2014-2016 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

package backend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/snapcore/snapd/osutil"
	"github.com/snapcore/snapd/snap"
)

// RemoveSnapData removes the data for the given version of the given snap.
func (b Backend) RemoveSnapData(snap *snap.Info) error {
	dirs, err := snapDataDirs(snap)
	if err != nil {
		return err
	}

	return removeDirs(dirs)
}

// RemoveSnapCommonData removes the data common between versions of the given snap.
func (b Backend) RemoveSnapCommonData(snap *snap.Info) error {
	dirs, err := snapCommonDataDirs(snap)
	if err != nil {
		return err
	}

	return removeDirs(dirs)
}

func removeDirs(dirs []string) error {
	for _, dir := range dirs {
		if err := os.RemoveAll(dir); err != nil {
			return err
		}

		// Attempt to remove the parent directory as well (ignore any failure)
		os.Remove(filepath.Dir(dir))
	}

	return nil
}

// snapDataDirs returns the list of data directories for the given snap version
func snapDataDirs(snap *snap.Info) ([]string, error) {
	// collect the directories, homes first
	found, err := filepath.Glob(snap.DataHomeDir())
	if err != nil {
		return nil, err
	}
	// then system data
	found = append(found, snap.DataDir())

	return found, nil
}

// snapCommonDataDirs returns the list of data directories common between versions of the given snap
func snapCommonDataDirs(snap *snap.Info) ([]string, error) {
	// collect the directories, homes first
	found, err := filepath.Glob(snap.CommonDataHomeDir())
	if err != nil {
		return nil, err
	}
	// then system data
	found = append(found, snap.CommonDataDir())

	return found, nil
}

// Copy all data for oldSnap to newSnap
// (but never overwrite)
func copySnapData(oldSnap, newSnap *snap.Info) (err error) {
	oldDataDirs, err := snapDataDirs(oldSnap)
	if err != nil {
		return err
	}

	newSuffix := filepath.Base(newSnap.DataDir())
	for _, oldDir := range oldDataDirs {
		// replace the trailing "../$old-suffix" with the "../$new-suffix"
		newDir := filepath.Join(filepath.Dir(oldDir), newSuffix)
		if err := copySnapDataDirectory(oldDir, newDir); err != nil {
			return err
		}
	}

	return nil
}

// Lowlevel copy the snap data (but never override existing data)
func copySnapDataDirectory(oldPath, newPath string) (err error) {
	if _, err := os.Stat(oldPath); err == nil {
		if _, err := os.Stat(newPath); err != nil {
			if err := osutil.CopyFile(oldPath, newPath, osutil.CopyFlagPreserveAll|osutil.CopyFlagSync); err != nil {
				return fmt.Errorf("cannot copy %q to %q: %v", oldPath, newPath, err)
			}
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
