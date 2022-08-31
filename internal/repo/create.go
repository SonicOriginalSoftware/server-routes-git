//revive:disable:package-comments

package repo

import (
	billy "github.com/go-git/go-billy/v5"
	go_git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/storage/filesystem/dotgit"
	"github.com/go-git/go-git/v5/storage/memory"
)

// Create a repo in a given filesystem
func Create(fsys billy.Filesystem) (repoFS billy.Filesystem, repo *go_git.Repository, err error) {
	dotGit := dotgit.New(fsys)
	if err = dotGit.Initialize(); err != nil {
		return
	}
	if _, err = dotGit.ConfigWriter(); err != nil {
		return
	}
	if err = dotGit.Close(); err != nil {
		return
	}
	repoFS = dotGit.Fs()
	repo, err = go_git.Init(memory.NewStorage(), nil)
	return
}
