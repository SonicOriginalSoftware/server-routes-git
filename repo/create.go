//revive:disable:package-comments

package repo

import (
	"io/fs"

	billy "github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
)

// Create a repo in a given filesystem
func Create(rootFS billy.Filesystem, repoPath string) (err error) {
	err = rootFS.MkdirAll(repoPath, fs.ModeDir)
	if err != nil {
		return
	}
	repoFS, err := rootFS.Chroot(repoPath)
	if err != nil {
		return
	}

	dotGit := dotgit.New(repoFS)
	if err = dotGit.Initialize(); err != nil {
		return
	}
	if _, err = dotGit.ConfigWriter(); err != nil {
		return
	}
	if err = dotGit.Close(); err != nil {
		return
	}
	return
}
