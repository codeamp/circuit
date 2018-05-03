// Code generated by go-bindata.
// sources:
// plugins/codeamp/schema.graphql
// plugins/codeamp/static/index.html
// DO NOT EDIT!

package assets

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _pluginsCodeampSchemaGraphql = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xcc\x58\x51\x73\x1b\x27\x10\x7e\xd7\xaf\x40\x93\x17\x65\xc6\xbf\xe0\x1e\x13\xa7\x6d\xda\xa6\x75\xad\xe6\x29\x93\x07\x7c\xc2\x12\xf5\xdd\xa1\x00\xe7\xd6\xd3\xe9\x7f\xef\x00\x0b\xb7\xbb\x77\x48\x96\x93\xce\xf4\x45\x3a\x16\x58\xf6\xfb\xd8\x5d\x16\x5c\x2b\x3b\x69\xc5\xef\xba\x57\x2b\xf8\xfe\x71\xfb\xeb\x2f\xab\x95\x6b\x0f\xaa\x97\xe2\xef\x95\x10\x5f\x46\x65\x9f\x1a\xf1\x5b\xf8\x5b\x09\xd1\x8f\x5e\x7a\x6d\x86\x46\x7c\x80\xaf\xd5\x3f\xab\xd5\x2b\xe8\xf7\x4f\x47\x95\x3e\xe3\xdc\x57\xe2\xa3\x53\x76\x25\xc4\xe8\x94\xdd\xe8\x5d\x23\xde\x5f\xbf\x6e\xb2\x30\xf5\x3a\xe8\x76\x9b\xd7\x8d\xf8\x14\x24\x9f\xd7\xb1\xf3\xc6\x9a\x3f\x54\xeb\x57\x42\x1c\xd3\x17\x28\xb8\x12\xae\x1b\xf7\x8d\xd8\x7a\xab\x87\xfd\x95\x18\x64\xaf\xa6\x96\x1a\x1e\xb5\x35\x43\xaf\x06\xff\xfe\x3a\x8b\x5f\x37\x48\x5b\xd1\xec\x26\xd5\x6e\x03\x1f\x5b\x25\x6d\x7b\x28\xc3\x53\xf3\xfd\x70\x1c\x7d\xb0\x0e\xa4\x60\xe0\x77\x4a\xfa\xd1\xaa\xa0\xe6\x1e\x3e\x23\x06\x90\xc3\xa8\xad\xb2\x8f\xba\x8d\xa3\x1c\x7c\xc6\x51\x20\xa7\xa3\x84\x3b\xaa\x16\x0d\xdd\x86\x26\x1e\x1e\x04\x30\xe5\x56\x75\x4a\xba\xa8\xd8\xc2\x67\x1c\x09\x72\x18\xf5\x6e\xe2\x23\x8c\x44\xf4\xa4\xd1\xa8\x7f\x3e\x43\x3c\x4a\xab\xe5\x5d\x07\xd6\xb7\x56\xf9\x6c\x4d\xf8\xa6\x1b\xf5\xee\x2f\xaf\x06\xa7\xcd\x10\x41\x84\xb5\xb2\xc0\x6d\x6a\xbb\xf2\xa9\x4c\xaa\xe8\x42\x7b\x34\xc9\x36\x68\x2f\xb8\x02\x40\x8f\x16\x9f\xf8\x61\x1a\x6e\x99\x34\x9b\xa0\x6c\xaf\x5d\x59\x7c\x6a\x85\x49\x21\x3c\xd6\xc9\xe3\x8b\xff\x47\xa7\xcf\x2d\xf0\xfb\xb7\x56\x49\xaf\xb2\xe9\x2b\x21\xda\x28\x00\xa3\xb3\xb7\x15\x3f\xcb\x1e\x86\xbd\xf4\xe3\x71\x47\x55\x8c\x51\x70\x89\x0a\xb0\x02\xe0\x17\x2b\x00\xf8\x06\xe4\x4d\x26\x2d\xab\xb8\x2d\xe3\x5f\x89\xad\x37\x47\xa4\xc0\x79\x73\xcc\xd3\x53\x38\xae\xd9\x04\x58\x13\x1c\xb8\xac\x09\xfe\xbb\x01\x79\x93\x3d\x3e\xaf\xb9\x2d\xe3\x0b\xf2\x49\x94\x90\x5f\xa2\xe2\x5a\x75\x8a\x58\xb1\x8b\x82\x4b\x54\x50\x20\xd9\xa9\x09\x9a\x10\x8d\x1b\x14\xaa\x45\x41\x68\x30\x9d\xdb\x34\xbf\xa0\x63\x7a\x09\xc4\x17\xe9\xa5\x90\xb3\x5e\x82\xfb\x45\x7a\x81\x07\x14\xc2\x85\x06\x94\x2a\x70\x88\x37\x38\x87\x64\xb5\xef\xc8\xfc\x42\x03\x55\x9b\x58\xf8\x1a\xb5\xc0\x02\x55\x9b\x48\xf8\x1a\xb5\x73\x12\x4a\x6e\x44\x4e\x11\x92\xe2\x26\xe5\xc9\xc0\x63\xf8\x9f\x58\x0d\xad\x0a\x72\xac\x2b\x3b\xc2\xf3\x74\xcd\xe1\x16\x5d\x02\xed\xfe\xf3\x94\x65\x90\x3c\x91\xc3\x66\x67\xf1\xa6\x0c\x68\x44\x11\x16\xe2\xb2\x80\x40\xe5\x1a\x61\x9f\x5f\xa0\x31\x03\xe6\x1a\x61\x8b\x5f\xa0\x91\xa3\xe6\xc9\x7a\xd2\xc9\x0f\xa2\x66\x76\x5c\xb1\x24\x7c\x9a\x0c\x9e\xd2\xbf\xd9\x42\x88\x23\x90\x25\x76\xfe\x23\x40\xa1\x7a\xc3\xe7\x64\xc1\x15\x8a\x39\x74\x9a\x6e\x46\xda\x4e\x85\x20\x12\x94\x3a\x2b\xd5\x07\x6b\x38\x90\xe9\x41\x48\x4a\x98\x19\x85\xb8\xbc\x39\xce\x65\x13\x10\x24\x2c\xcb\x22\x61\x5e\xfb\x8d\x31\x0f\xbd\xb4\x0f\xe8\x18\xbe\x03\xd1\x0d\xa9\x4a\xc3\x31\xf8\xc6\x98\x4e\xc9\x01\xea\x83\x58\xe5\xc6\xda\x20\x7c\xc5\xba\x00\x86\x86\xda\xa8\x97\xba\xcb\x85\xd0\x9a\xd6\x19\x8c\x80\xe4\x8c\xbb\x26\x96\xe9\xa0\x1b\x4a\xcc\xa4\x1e\x1a\x6c\x05\x5e\x1f\x04\x59\xaf\x9c\x93\x7b\x85\xd7\x0d\x7b\x82\xdb\x07\xe9\x0e\xc4\x2e\x69\xd5\xe0\x7f\x60\x52\xab\xee\x71\x73\xc9\xc4\x7c\x94\x46\x13\x73\x79\x7b\xde\xc4\xd6\xf4\xbd\x1c\x76\x58\x3b\x2e\xf0\xd7\xb4\x36\x26\xa7\x55\x9a\x3e\x86\x8c\x8e\xec\x37\x36\x6c\xfb\xa7\x50\xb8\x45\x3a\x6b\x99\x3f\xf4\x05\x5b\xcf\xe1\xc2\x27\x43\xc4\x86\xeb\x65\x8a\x8f\x1b\xfe\xa0\x9e\x70\x53\xbb\x6b\x75\x2f\xc7\xce\x23\xcf\x09\x08\x3a\x63\x2f\x30\x62\x3a\x3b\x80\xe9\x90\xce\x99\x21\x6c\xe1\x47\xd9\x8d\x94\xd1\xd6\x50\xe0\x4b\x3b\x93\x3c\x25\xf8\xf2\x12\x55\x8f\xca\x16\xe7\x9d\x2e\x07\xa7\xc8\x66\xb8\x02\x1f\x5b\x38\xa1\x68\x20\xe1\xbb\x11\x71\xa7\xb0\xe9\x67\x28\x6f\x8f\xe3\xad\xfa\x32\x2a\xe7\x99\xf4\x67\xdd\x6b\x22\xeb\x55\x6f\xec\xd3\xc2\xe0\xd4\x31\x1b\xef\x43\xc0\x0e\xb1\xe6\xff\xde\xca\x56\xdd\x28\xab\xcd\xee\xdc\xc6\xe5\x4a\x39\xc2\xc8\x77\x95\xf3\x51\x41\xb9\x97\xd6\xeb\x7b\xd9\x06\xbf\x4e\xf7\x11\x21\x0e\x4a\xee\x20\x0f\x34\x39\x21\x44\x2b\xa5\xee\x96\xe4\xce\x4b\xaf\x68\x44\xb3\x2b\x52\xed\x82\x14\x67\x7e\x98\x27\x92\x0b\xb6\x3a\x32\x91\xef\x29\x91\x09\x68\x9c\xd9\x4c\x7c\xef\x4f\x36\x1f\x8d\xd3\xde\xd8\x27\x9a\x1e\xa0\xcc\x29\x92\xbd\xf6\x1f\x6d\xc7\x24\x37\xd6\x78\xd3\x1a\x22\xb6\x4e\xde\x58\xfd\x28\xbd\xfa\x89\x46\x4c\xe8\x18\xef\x3a\xdd\x32\x79\xb9\xf7\xbb\x83\xf9\xf3\x5a\x1d\x3b\xf3\x14\x60\x82\xff\xf2\xc7\x80\x76\xb4\x21\x99\xde\xb2\x6b\x17\xba\xc0\xb3\xeb\x7b\x7e\x30\x60\xcf\x05\x70\x13\x67\xa1\x86\x77\x6e\xf1\x72\xbc\xd7\xfe\x8d\x95\x43\x4b\x32\x79\x6b\x06\xaf\x87\xd1\x8c\x2e\x99\x4f\x72\x91\x22\x47\xe7\xfc\x7c\xcc\x47\x21\xc2\x5c\xcd\x58\xb4\x60\x4b\x99\xb3\xc8\xce\x04\xb1\xe9\x8f\x66\x50\x34\xb1\xf3\x04\xc4\x72\xdc\x49\x77\x34\xc3\xbd\xde\x4f\xc1\x73\xc2\x3d\xa7\x82\x07\xfb\x69\xcd\xf0\xa5\xd0\x5d\x2a\x44\x2b\x41\x3c\xb3\x6b\x74\xde\xf4\x6f\x99\x74\x16\xba\xdf\x2a\x22\x67\x4f\x27\x24\x4b\xd5\x30\xf3\x47\x04\x86\x99\x33\x86\xdd\x7a\xab\xf7\x03\xa4\x26\x1e\xbe\x95\x3e\x0e\x9d\x3b\x41\x8d\x8a\x05\xb2\xef\xf5\xa0\xdd\x01\x9f\x3f\x27\xdc\x40\xe8\x50\x23\xae\xe2\x2f\x79\x70\x29\x64\xa4\xb5\xea\xa9\x65\x31\x07\xd1\x70\x5c\x0e\x28\xba\x91\xd3\xeb\xd9\xa9\xd8\x25\xb6\xa7\x37\xcc\x25\x04\xe8\x75\x33\xe2\x98\xe7\xd3\xab\x5a\x90\x13\x87\xc1\xaa\xf1\x53\xd2\x9c\x1c\x74\x4c\x4d\x48\x50\xe4\x50\xe1\x22\xee\xb8\x77\xc6\xb6\xea\x56\xdd\x8d\xba\x9b\x19\x95\xab\x05\x6c\x14\x7e\xe5\x99\x1b\xb5\xb8\xf6\x65\x75\x28\x9f\x5b\x29\x42\xc1\x8e\x1b\x63\x93\xf7\xac\x3f\x2f\xf8\x70\x05\x75\x04\xf7\xd6\x0c\x5e\xea\x21\xdc\xb7\x8c\xf5\x4b\x18\x8b\xee\x88\x33\x8c\x62\x34\x33\xcf\x9c\x55\x58\x4b\x4a\xcb\xf3\xd0\x9c\xbc\xff\x4f\xc1\x35\x2b\x8c\x31\x10\xfe\xc4\x73\x1e\xc8\x4b\x8a\xf5\x6a\x6d\xce\x48\x2d\xef\x2f\x73\x33\xce\x95\xea\xb3\x8c\x57\x29\xdd\x49\xa2\xa8\x06\x52\xad\xe4\x5e\x7c\xd2\xa7\x84\x92\x17\x82\x8b\xe9\x5c\x3c\xd2\xab\x66\xf2\x83\x91\x92\xb0\x6c\xf1\x42\xbe\x3b\x69\xf3\xd5\x62\x26\xb8\xc2\xa7\x19\x93\x3f\xef\xb8\x3e\x15\xce\xf4\xfd\x83\x98\xbc\xf4\x34\x12\x2d\x0e\x57\x01\x96\x3b\xc9\xd3\x01\x9d\x08\x69\xe6\xd2\xf5\xa6\xe5\x66\x0e\xb8\xb7\x72\xa8\xfa\x4b\x25\xfa\xe6\xdd\xd3\x02\xd5\x4d\x7f\xe6\x42\x89\x9a\xea\x42\x13\x73\x4b\x69\x9e\x52\x57\x31\x33\xd1\xf7\x6f\x00\x00\x00\xff\xff\x17\xb1\x66\x75\x2c\x1d\x00\x00")

func pluginsCodeampSchemaGraphqlBytes() ([]byte, error) {
	return bindataRead(
		_pluginsCodeampSchemaGraphql,
		"plugins/codeamp/schema.graphql",
	)
}

func pluginsCodeampSchemaGraphql() (*asset, error) {
	bytes, err := pluginsCodeampSchemaGraphqlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "plugins/codeamp/schema.graphql", size: 7468, mode: os.FileMode(436), modTime: time.Unix(1525379419, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _pluginsCodeampStaticIndexHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xb4\x56\xeb\x6f\xdb\x36\x10\xff\xde\xbf\xe2\xe6\x6c\x90\x5d\xd8\x94\xd3\xf5\x01\xa8\x76\x86\xb6\x49\xbb\x16\x69\xd3\xc6\x19\x8a\x7d\x2b\x4d\x9e\x2c\xa6\x14\xa9\x1e\x29\x3b\x6a\x91\xff\x7d\xa0\x64\xc9\x8a\x91\x0c\xd9\xb0\x41\x1f\x4c\xde\xe3\x77\x0f\xde\xc3\xb3\x9f\x8e\xcf\x5e\x5d\xfc\xf9\xf1\x04\x32\x9f\xeb\xa3\x07\xb3\xe6\x07\x60\x96\x21\x97\xe1\x00\x30\x73\xbe\xd2\xd8\x9c\x01\x96\x56\x56\xf0\x63\x7b\x01\xc8\x50\xad\x32\x9f\xc0\xe1\x74\xfa\xcb\xf3\x8e\x9a\x73\x5a\x29\x93\xc0\x74\x47\xda\x28\xe9\xb3\x7d\x39\xbb\x46\x4a\xb5\xdd\x24\x90\x29\x29\xd1\xb4\x9c\xeb\xed\xef\xc1\x8a\x78\x91\xa9\x6f\xfa\x76\x8b\xeb\x6c\x5f\x81\x5d\x6e\xfc\xc4\xdb\xaf\x68\x7a\x1a\x4b\x2e\xbe\xae\xc8\x96\x46\x26\xa0\x95\x41\x4e\x93\x15\x71\xa9\xd0\xf8\xe1\x41\xfa\x2c\x7c\x63\x38\xc0\x47\xe1\x1b\xed\x9c\x5b\x5a\x92\x48\x93\xa5\xf5\xde\xe6\x09\x1c\x16\x57\xe0\xac\x56\x12\x0e\xe4\x34\x7c\x3b\xc9\xd4\x1a\x3f\x49\x79\xae\x74\x95\x80\xab\x9c\xc7\x7c\x0c\x13\x5e\x14\x1a\x27\xed\x35\x5a\x70\x03\xaf\x89\x1b\xa1\x9c\xb0\xd1\x18\x22\xb6\x78\xfd\x61\x71\xac\x5c\xa1\x79\x35\x39\xc7\x55\xa9\x39\x05\xfa\x02\x57\x16\xe1\x8f\xb7\xd1\x18\xea\x63\x47\xfa\xfc\x31\xb0\x7f\x47\xbd\x46\xaf\x04\x87\x0f\x58\x62\x34\x86\xac\x25\x8c\x21\x3a\x2d\x85\x92\x1c\xde\x10\x37\x32\xf0\x38\x29\xae\xc7\xe0\xb8\x71\x13\x87\xa4\xd2\x9d\xd3\x05\x97\x52\x99\x55\x02\xcf\x8a\x2b\x38\x7c\x5c\x5c\xc1\xd3\xe2\x6a\x2f\x26\xa7\xbe\x63\x52\x33\x6f\x26\x7a\x16\xf7\x6a\x62\xa6\x95\xf9\x0a\x84\x7a\x3e\xa8\xa9\x2e\x43\xf4\x03\xc8\x08\xd3\xf9\x20\xf3\xbe\x70\x49\x1c\x0b\x69\x2e\x1d\x13\xda\x96\x32\xd5\x9c\x90\x09\x9b\xc7\xfc\x92\x5f\xc5\x5a\x2d\x5d\xdc\xbe\x73\x3c\x65\x87\x53\xf6\xa8\xbb\x33\xe1\xdc\x00\xe2\xdb\x0a\x31\x7e\x08\x67\x6b\x24\x52\x12\x1d\x3c\x8c\xdb\x02\x68\x35\x27\xc2\x1a\xcf\x95\x41\x02\xb6\x0e\x69\x58\x6a\x9c\xa0\x54\xde\xd2\x2d\xc5\xf4\xf4\xe9\xdf\x87\xe8\x04\xa9\xc2\x83\x23\x71\xef\x90\x52\xf4\x22\x8b\x1f\xb1\x29\xfb\xb5\x39\xb3\x5c\x19\x76\xe9\x06\x47\xb3\xb8\x81\xfb\xf7\xd8\x84\x5c\xf8\xf8\xf0\x09\x7b\xc2\x1e\x37\x97\xff\x15\x7c\x22\x6d\xfe\x1f\x1a\xb8\xf3\xb1\xf7\xe1\x67\x71\x3b\x85\x66\x61\xec\x6c\x2d\x4a\xb5\x06\xa1\xb9\x73\xf3\x41\xd7\xed\x83\xa3\x77\x9f\x2f\xe0\xa2\x6e\xfc\x99\x32\x45\xe9\x41\xc9\x3e\x1f\x0a\xcd\x05\x66\x56\x4b\xa4\xf9\x60\x27\xbc\xb2\xe8\x20\x43\xc2\x60\x59\xaa\x75\xcf\x46\x00\x68\x5d\x1b\x1c\x9d\x5a\x1e\xda\x85\x31\xd6\x97\xeb\xa7\x62\xcd\x09\x1c\x72\x12\x19\xcc\x61\xa3\x8c\xb4\x1b\xa6\xad\xe0\x5e\x59\xc3\x1a\xc6\xf3\x4e\xb0\xe0\xc4\x73\xf4\x48\x0e\xe6\xf0\xe3\xba\x61\x48\x2b\xca\x1c\x8d\x67\x2b\xf4\x27\x1a\xc3\xf1\x65\xf5\x56\x0e\xa3\x2e\x8e\x68\xc4\xd6\x5c\x97\x08\x73\x08\xd0\x7a\xe1\x2d\xf1\x15\x06\x85\xb7\x1e\xf3\x61\xd4\x3a\x9c\x5c\x6e\xfc\x45\xa3\xf1\xfc\x41\x0d\x9e\x96\x46\x04\x57\xa0\x16\xf9\x74\xfa\x3a\xd4\x24\xd2\x70\x7b\xfd\x18\x1c\x72\xa3\xae\x37\x84\x35\xce\x43\x8b\x02\xf3\x7f\xe0\x5c\xdb\x48\x1a\x3d\x84\x07\xdc\x06\xd9\x35\x5d\xf4\x42\x08\x2c\x7c\x94\x40\x14\x66\xa4\x6a\x52\x14\x5f\x3a\x6b\xa2\xf1\x4e\xea\x95\x35\x1e\x8d\x9f\x5c\x54\x05\xde\x2a\xdb\xf6\x6b\x6b\x4f\xa5\x30\x6c\x1d\x1e\xf5\xec\xdd\xc8\x94\xbb\x3b\x53\xe3\x2e\xdc\xde\x16\xb8\x2d\x80\xfb\x86\x70\x9f\x20\x6e\x48\xbf\x28\x7d\x66\x49\x7d\xaf\xf9\x51\xb2\x4b\xff\x6f\xf0\xe5\x25\x72\x42\x82\x9f\x7f\xb4\xc4\xeb\x2f\x90\x80\x29\xb5\xee\x10\xae\xf7\xd7\x21\xa1\x2f\xc9\x40\x3d\x7e\x86\x51\xfc\xad\x44\xaa\xa2\x71\x2f\x92\x1c\x7d\x66\x65\x02\x51\x61\x9d\xef\xf9\xb2\x8d\x7a\xdc\x5b\x86\xb2\x4a\xe0\xdd\xe2\xec\x03\x73\x9e\x94\x59\xa9\xb4\xda\x2b\x9d\x9d\xb0\x20\x94\x68\xbc\xe2\xda\x25\x10\x29\x23\x74\x19\xb6\x51\xeb\xdc\x88\xf9\x0c\xcd\xb0\x2b\xc8\x21\xa1\x2b\xac\x71\xd8\x7f\xb4\xad\xeb\x2d\x8b\x79\xbc\xf2\xc3\xee\x5d\xee\xc6\x78\x69\x65\xd5\xc7\xf1\x54\xdd\x78\xb9\x2d\x6e\x1d\x49\xc1\xc9\xe1\x4d\xcd\xdd\xc3\x5f\x83\xe0\x5e\x64\x30\x44\x22\x4b\xa3\xdb\x40\xfa\x9a\x3d\xc5\xce\xc7\x86\xd6\xdc\xe3\x18\xce\xd1\x48\x24\x98\xbd\xa9\x2b\xef\xd3\x29\xc4\x47\xa0\x8c\xb7\xe0\x33\xac\x13\xcc\x5a\xc9\x05\x62\x4d\x3c\x3f\x79\x71\xfc\xfe\x04\x94\xa9\x6f\xde\x16\xa0\x71\x8d\x1a\x6c\x0a\x3e\x53\x0e\x72\x2b\x4b\x1d\x18\xa0\x91\x93\x81\xdc\x12\x02\x5f\xda\xd2\xb7\x48\x99\xdd\x40\x65\x4b\x10\xdc\x80\x28\x9d\xb7\xb9\xfa\x8e\xd0\x79\xb0\xac\xa0\x20\xbb\x56\x61\xb4\x81\x54\x69\x8a\x84\xc6\x43\xdd\xc6\x0e\x2c\xb5\x30\xe1\xbf\x42\xc8\x33\xd7\x20\x32\xa5\x25\x60\x33\x01\x5c\xe3\xf2\x79\xd8\x12\xc7\x67\xef\x19\xd5\x21\x0e\xb7\x19\xa8\xc9\x4c\x10\x72\x8f\xdb\x91\x31\x6c\x4d\xf7\xab\x30\x6d\x46\x51\xb2\x37\x9a\x76\x05\xd3\x9e\xee\x9c\x41\x6d\x33\x47\xa3\x5a\x72\x9b\xf9\x9b\xbb\xa4\x59\x21\xb3\xb8\xf9\x8b\xfb\x57\x00\x00\x00\xff\xff\x45\xab\x31\x54\xfa\x0a\x00\x00")

func pluginsCodeampStaticIndexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_pluginsCodeampStaticIndexHtml,
		"plugins/codeamp/static/index.html",
	)
}

func pluginsCodeampStaticIndexHtml() (*asset, error) {
	bytes, err := pluginsCodeampStaticIndexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "plugins/codeamp/static/index.html", size: 2810, mode: os.FileMode(436), modTime: time.Unix(1524543030, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"plugins/codeamp/schema.graphql": pluginsCodeampSchemaGraphql,
	"plugins/codeamp/static/index.html": pluginsCodeampStaticIndexHtml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}
var _bintree = &bintree{nil, map[string]*bintree{
	"plugins": &bintree{nil, map[string]*bintree{
		"codeamp": &bintree{nil, map[string]*bintree{
			"schema.graphql": &bintree{pluginsCodeampSchemaGraphql, map[string]*bintree{}},
			"static": &bintree{nil, map[string]*bintree{
				"index.html": &bintree{pluginsCodeampStaticIndexHtml, map[string]*bintree{}},
			}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}

