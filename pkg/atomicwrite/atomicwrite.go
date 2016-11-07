package atomicwrite

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

var (
	ErrSymlinkTooDeep   = errors.New("symlink_too_deep")
	ErrPathIsDirectory  = errors.New("path_is_directory")
	ErrPermissionDenied = errors.New("permission_denied")
	ErrFileExist        = errors.New("file_exists_at_path")
	ErrInvalidUser      = errors.New("invalid_user")

	// Unexpected error
	ErrInvalidOwner = errors.New("failed_to_get_file_owner")

	TempDir = ".nitrous-solo-temp"
)

func AtomicWrite(r io.Reader, path string, runAs string, perm os.FileMode, overwrite bool) error {
	targetPath := path

	if runAs == "" {
		return ErrInvalidUser
	}

	u, err := user.Lookup(runAs)
	if err != nil {
		return ErrInvalidUser
	}

	fi, err := os.Lstat(targetPath)
	if err != nil && !os.IsNotExist(err) {
		if os.IsPermission(err) {
			return ErrPermissionDenied
		}
		return err
	}

	if fi != nil {
		if fi.Mode()&os.ModeSymlink > 0 {
			// file is symlink, try to follow
			fi, err = os.Stat(targetPath)
			if err != nil && !os.IsNotExist(err) {
				if strings.Contains(strings.ToLower(err.Error()), "too many levels of symbolic links") {
					return ErrSymlinkTooDeep
				}
				return err
			}

			// was possible to follow symlink, evaluate symlink and get target path
			p, err := filepath.EvalSymlinks(targetPath)
			if err != nil {
				return err
			}
			targetPath = p
		}

		if fi.IsDir() {
			return ErrPathIsDirectory
		}
	}

	if fi != nil {
		// if file exists, make sure it's writable
		ff, err := os.OpenFile(targetPath, os.O_RDWR, fi.Mode().Perm())
		if err != nil {
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return err
		}
		ff.Close()

		if !overwrite {
			return ErrFileExist
		}
	}

	tempDir := filepath.Join(u.HomeDir, TempDir)
	if err := os.MkdirAll(tempDir, 0700); err != nil {
		return errors.New(fmt.Sprintf(`could not create directory "%s"`, tempDir))
	}

	f, err := ioutil.TempFile(tempDir, "nitrous-temp")
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = io.Copy(f, r); err != nil {
		return err
	}

	if fi != nil {
		// if file already exists at path
		// copy permission
		if err := f.Chmod(fi.Mode().Perm()); err != nil {
			return err
		}

		// copy owner
		s, ok := fi.Sys().(*syscall.Stat_t)
		if !ok {
			return ErrInvalidOwner
		}

		if err := f.Chown(int(s.Uid), int(s.Gid)); err != nil {
			return err
		}
	} else {
		// it is a new file
		uid, err := strconv.Atoi(u.Uid)
		if err != nil {
			return err
		}

		gid, err := strconv.Atoi(u.Gid)
		if err != nil {
			return err
		}

		if err := f.Chown(uid, gid); err != nil {
			return err
		}

		if perm == 0 {
			if err := f.Chmod(0644); err != nil {
				return err
			}
		}
	}

	f.Close()

	// move tmp file over to the intended file path
	if err := os.Rename(f.Name(), targetPath); err != nil {
		os.Remove(f.Name())
		return err
	}

	return nil
}
