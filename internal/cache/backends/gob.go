package backends

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"github.com/sinclairtarget/git-who/internal/cache"
	"github.com/sinclairtarget/git-who/internal/git"
)

// Stores commits on disk at a particular filepath.
//
// Commits are stored in Gob format. The file stored on disk is a series of
// Gob-encoded arrays, each prefixed with a four-byte value indicating the
// number of bytes in the next array. This framing creates redundancy (since
// the Gob type metadata is repeated for each array) but allows us to append to
// the file on disk instead of replacing the whole file when we want to cache
// new commits.
type GobBackend struct {
	Path string
}

func (b GobBackend) Name() string {
	return "gob"
}

func (b GobBackend) Get(revs []string) (cache.Result, error) {
	result := cache.EmptyResult()

	lookingFor := map[string]bool{}
	for _, rev := range revs {
		lookingFor[rev] = true
	}

	f, err := os.Open(b.Path)
	if errors.Is(err, fs.ErrNotExist) {
		return result, nil
	} else if err != nil {
		return result, err
	}

	// In theory we shouldn't get any duplicates into the cache if we're
	// careful about what we write to it. But let's make sure by detecting dups
	// and throwing an error if we see one.
	seen := map[string]bool{}

	it := func(yield func(git.Commit, error) bool) {
		defer f.Close() // Don't care about error closing when reading

		for {
			var commit git.Commit

			// -- Find length of next gob in bytes --
			prefix := make([]byte, 4)
			_, err = f.Read(prefix)
			if err == io.EOF {
				return
			} else if err != nil {
				yield(commit, err)
				return
			}

			var size uint32
			err = binary.Read(
				bytes.NewReader(prefix),
				binary.LittleEndian,
				&size,
			)
			if err != nil {
				yield(commit, err)
				return
			}

			// -- Decode next gob --
			var commits []git.Commit

			data := make([]byte, size)
			_, err = f.Read(data)

			dec := gob.NewDecoder(bytes.NewReader(data))
			err = dec.Decode(&commits)
			if err != nil {
				yield(commit, err)
				return
			}

			// -- Yield matching commits --
			for _, c := range commits {
				hit, _ := lookingFor[c.Hash]
				if hit {
					if isDup, _ := seen[c.Hash]; isDup {
						yield(c, fmt.Errorf(
							"duplicate commit in cache: %s",
							c.Hash,
						))
						return
					}

					seen[c.Hash] = true
					if !yield(c, nil) {
						return
					}
				}
			}
		}
	}

	return cache.Result{Commits: it}, nil
}

func (b GobBackend) Add(commits []git.Commit) (err error) {
	f, err := os.OpenFile(
		b.Path,
		os.O_WRONLY|os.O_APPEND|os.O_CREATE,
		0644,
	)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := f.Close()
		if err == nil {
			err = closeErr
		}
	}()

	var data bytes.Buffer

	enc := gob.NewEncoder(&data)
	err = enc.Encode(&commits)
	if err != nil {
		return err
	}

	if data.Len() > 0xFFFF_FFFF {
		return errors.New(
			"cannot add more than 4,294,976,296 bytes to cache at once", // lol
		)
	}

	err = binary.Write(f, binary.LittleEndian, uint32(data.Len()))
	if err != nil {
		return err
	}

	_, err = f.Write(data.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func (b GobBackend) Clear() error {
	return os.Remove(b.Path)
}
