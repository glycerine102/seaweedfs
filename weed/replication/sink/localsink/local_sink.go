package localsink

import (
	"github.com/chrislusf/seaweedfs/weed/filer"
	"github.com/chrislusf/seaweedfs/weed/glog"
	"github.com/chrislusf/seaweedfs/weed/pb/filer_pb"
	"github.com/chrislusf/seaweedfs/weed/replication/repl_util"
	"github.com/chrislusf/seaweedfs/weed/replication/sink"
	"github.com/chrislusf/seaweedfs/weed/replication/source"
	"github.com/chrislusf/seaweedfs/weed/util"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type LocalSink struct {
	dir         string
	filerSource *source.FilerSource
}

func init() {
	sink.Sinks = append(sink.Sinks, &LocalSink{})
}

func (localsink *LocalSink) SetSourceFiler(s *source.FilerSource) {
	localsink.filerSource = s
}

func (localsink *LocalSink) GetName() string {
	return "local"
}

func (localsink *LocalSink) isMultiPartEntry(key string) bool {
	return strings.HasSuffix(key, ".part") && strings.Contains(key, "/.uploads/")
}

func (localsink *LocalSink) initialize(dir string) error {
	localsink.dir = dir
	return nil
}

func (localsink *LocalSink) Initialize(configuration util.Configuration, prefix string) error {
	dir := configuration.GetString(prefix + "directory")
	glog.V(4).Infof("sink.local.directory: %v", dir)
	return localsink.initialize(dir)
}

func (localsink *LocalSink) GetSinkToDirectory() string {
	return localsink.dir
}

func (localsink *LocalSink) DeleteEntry(key string, isDirectory, deleteIncludeChunks bool, signatures []int32) error {
	if localsink.isMultiPartEntry(key) {
		return nil
	}
	glog.V(4).Infof("Delete Entry key: %s", key)
	if err := os.Remove(key); err != nil {
		return err
	}
	return nil
}

func (localsink *LocalSink) CreateEntry(key string, entry *filer_pb.Entry, signatures []int32) error {
	if entry.IsDirectory || localsink.isMultiPartEntry(key) {
		return nil
	}
	glog.V(4).Infof("Create Entry key: %s", key)

	totalSize := filer.FileSize(entry)
	chunkViews := filer.ViewFromChunks(localsink.filerSource.LookupFileId, entry.Chunks, 0, int64(totalSize))

	dir := filepath.Dir(key)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		glog.V(4).Infof("Create Direcotry key: %s", dir)
		if err = os.MkdirAll(dir, 0); err != nil {
			return err
		}
	}

	writeFunc := func(data []byte) error {
		writeErr := ioutil.WriteFile(key, data, 0)
		return writeErr
	}

	if err := repl_util.CopyFromChunkViews(chunkViews, localsink.filerSource, writeFunc); err != nil {
		return err
	}

	return nil
}

func (localsink *LocalSink) UpdateEntry(key string, oldEntry *filer_pb.Entry, newParentPath string, newEntry *filer_pb.Entry, deleteIncludeChunks bool, signatures []int32) (foundExistingEntry bool, err error) {
	if localsink.isMultiPartEntry(key) {
		return true, nil
	}
	glog.V(4).Infof("Update Entry key: %s", key)
	// do delete and create
	return false, nil
}
