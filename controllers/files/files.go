package files

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/nitrous-io/nitrous-solo/controllers"
	"github.com/nitrous-io/nitrous-solo/pkg/atomicwrite"
	"github.com/nitrous-io/nitrous-solo/pkg/token"
)

func Upload(c *gin.Context) {
	ud := controllers.CurrentUserData(c)

	reader, err := c.Request.MultipartReader()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "invalid_request",
			"error_description": "the request should be encoded in multipart/form-data format",
		})
		return
	}

	var location string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			c.JSON(422, gin.H{
				"error": "invalid_params",
				"errors": map[string]interface{}{
					"files": "is required",
				},
			})
			return
		}

		if part.FormName() == "location" {
			out, err := ioutil.ReadAll(part)
			if err != nil {
				controllers.InternalServerError(c, err)
				return
			}

			location = string(out)
		} else if part.FormName() == "files[]" {
			if location == "" {
				c.JSON(422, gin.H{
					"error": "invalid_params",
					"errors": map[string]interface{}{
						"location": "is required",
					},
				})
				return
			}

			if err := atomicwrite.AtomicWrite(part, filepath.Join(location, part.FileName()), ud.User, true); err != nil {
				if err == atomicwrite.ErrPermissionDenied {
					c.JSON(http.StatusForbidden, gin.H{
						"error":             "invalid_request",
						"error_description": "the user does not have permission to write to the path",
					})
					return
				}

				if err == atomicwrite.ErrInvalidUser {
					c.JSON(http.StatusForbidden, gin.H{
						"error":             "invalid_request",
						"error_description": "the user could not be found",
					})
					return
				}

				controllers.InternalServerError(c, err)
				return
			}

			break
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func Download(c *gin.Context) {
	tok := c.Query("token")
	path := c.Query("path")

	if tok == "" {
		c.JSON(422, gin.H{
			"error": "invalid_params",
			"errors": map[string]string{
				"token": "is required",
			},
		})
		return
	}

	ud := controllers.GetUserDataByToken(tok, token.TempType)
	if ud == nil {
		c.JSON(422, gin.H{
			"error": "invalid_credentials",
			"errors": map[string]string{
				"token": "is invalid",
			},
		})
		return
	}

	if path == "" {
		c.JSON(422, gin.H{
			"error": "invalid_params",
			"errors": map[string]interface{}{
				"path": "is required",
			},
		})
		return
	}

	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(422, gin.H{
				"error": "invalid_params",
				"errors": map[string]interface{}{
					"path": "could not be found",
				},
			})
			return
		}
		controllers.InternalServerError(c, err)
		return
	}

	targetPath := path
	c.Writer.Header().Set("Expires", "0")
	c.Status(http.StatusOK)

	if fi.IsDir() {
		target, err := ioutil.TempFile("", "nitrous-download")
		if err != nil {
			controllers.InternalServerError(c, err)
			return
		}

		defer func() {
			target.Close()
			os.Remove(target.Name())
		}()

		if err := bundle(path, target); err != nil {
			controllers.InternalServerError(c, err)
			return
		}
		targetPath = target.Name()

		c.Writer.Header().Set("Content-Transfer-Encoding", "binary")
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(path)+".zip")

		c.File(targetPath)
	} else {
		c.Writer.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(path))
		c.File(targetPath)
	}
}

func bundle(basePath string, target *os.File) error {
	var (
		list  []os.FileInfo
		paths []string
	)

	walkFn := func(path string, fi os.FileInfo, err error) error {
		// if there is an error lstat-ing a file, just skip it
		if err != nil {
			return nil
		}

		if err != nil {
			// SkipDir error should fall through
			if err == filepath.SkipDir {
				return err
			}
			return nil
		}

		if !fi.Mode().IsRegular() {
			return nil
		}

		list = append(list, fi)
		paths = append(paths, path)

		return nil
	}

	if err := filepath.Walk(basePath, walkFn); err != nil {
		return err
	}

	zw := zip.NewWriter(target)
	defer func() {
		zw.Flush()
		zw.Close()
	}()

	for i, fi := range list {
		absPath := paths[i]

		hdr, err := zip.FileInfoHeader(fi)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(basePath, absPath)
		if err != nil {
			return err
		}
		hdr.Name = relPath

		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}

		f, err := os.Open(absPath)
		if err != nil {
			return err
		}
		defer f.Close()

		if _, err := io.Copy(w, f); err != nil {
			return err
		}
		f.Close()
	}

	return nil
}
