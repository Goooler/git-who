package git_test

import (
	"testing"

	"github.com/sinclairtarget/git-who/internal/git"
	"github.com/sinclairtarget/git-who/internal/iterutils"
)

func TestCommitsFileRename(t *testing.T) {
	path := "file-rename"

	commitsSeq, closer, err := git.Commits([]string{"HEAD"}, []string{path})
	if err != nil {
		t.Fatalf("error getting commits: %v", err)
	}

	defer func() {
		err := closer()
		if err != nil {
			t.Errorf("encountered error cleaning up: %v", err)
		}
	}()

	commits, err := iterutils.Collect(commitsSeq)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(commits) != 3 {
		t.Fatalf("expected 3 commits but found %d", len(commits))
	}

	commit := commits[1]
	if commit.Hash != "879e94bbbcbbec348ba1df332dd46e7314c62df1" {
		t.Errorf(
			"expected commit to have hash %s but got %s",
			"879e94bbbcbbec348ba1df332dd46e7314c62df1",
			commit.Hash,
		)
	}

	if len(commit.FileDiffs) != 1 {
		t.Errorf(
			"len of commit file diffs should be 1, but got %d",
			len(commit.FileDiffs),
		)
	}

	diff := commit.FileDiffs[0]
	if diff.Path != "file-rename/foo.go" {
		t.Errorf(
			"expected diff path to be %s but got \"%s\"",
			"file-rename/foo.go",
			diff.Path,
		)
	}

	if diff.MoveDest != "file-rename/bim.go" {
		t.Errorf(
			"expected diff move dest to be %s but got \"%s\"",
			"file-rename/bim.go",
			diff.MoveDest,
		)
	}
}

// Test moving a file into a new directory (to make sure we handle { => foo})
func TestCommitsFileRenameNewDir(t *testing.T) {
	path := "rename-new-dir"
	commitsSeq, closer, err := git.Commits([]string{"HEAD"}, []string{path})
	if err != nil {
		t.Fatalf("error getting commits: %v", err)
	}

	defer func() {
		err := closer()
		if err != nil {
			t.Errorf("encountered error cleaning up: %v", err)
		}
	}()

	commits, err := iterutils.Collect(commitsSeq)
	if err != nil {
		t.Fatalf(err.Error())
	}

	if len(commits) != 2 {
		t.Fatalf("expected 2 commits but found %d", len(commits))
	}

	commit := commits[1]
	if commit.Hash != "13b6f4f70c682ab06da9ef433cdb4fcbf65d78c3" {
		t.Errorf(
			"expected commit to have hash %s but got %s",
			"13b6f4f70c682ab06da9ef433cdb4fcbf65d78c3",
			commit.Hash,
		)
	}

	if len(commit.FileDiffs) != 1 {
		t.Errorf(
			"len of commit file diffs should be 1, but got %d",
			len(commit.FileDiffs),
		)
	}

	diff := commit.FileDiffs[0]
	if diff.Path != "rename-new-dir/hello.txt" {
		t.Errorf(
			"expected diff path to be %s but got \"%s\"",
			"rename-new-dir/hello.txt",
			diff.Path,
		)
	}

	if diff.MoveDest != "rename-new-dir/foo/hello.txt" {
		t.Errorf(
			"expected diff move dest to be %s but got \"%s\"",
			"rename-new-dir/foo/hello.txt",
			diff.MoveDest,
		)
	}
}
