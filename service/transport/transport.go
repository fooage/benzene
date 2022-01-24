package transport

import (
	"errors"
	"io"
	"log"
	"os"
	"time"

	"github.com/fooage/messier/proto/pb"
	"github.com/fooage/messier/utils"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

// Porter will be responsible for the transfer of file backups between nodes.
type Porter struct {
	info             *pb.Information
	duration         time.Duration
	chunkSize        int64
	storageDirectory string // node's file storage directory
	pb.UnimplementedPorterServer
}

func NewPorter(info *pb.Information) *Porter {
	storageDirectory := viper.GetString("server.transport.storage_directory")
	exist, err := utils.CheckPathExists(storageDirectory)
	if err != nil {
		log.Fatalf("check path has an error: %v\n", err)
	}
	if !exist {
		os.Mkdir(storageDirectory, os.ModePerm)
	}
	return &Porter{
		info:             info,
		duration:         time.Second * viper.GetDuration("server.transport.duration"),
		chunkSize:        viper.GetInt64("server.transport.chunk_size") * 1024,
		storageDirectory: storageDirectory,
	}
}

// MoveFile is a function which do the basic file trans. The file hash is
// treated as a unique identifier of the storage path and file.
func (p *Porter) MoveFile(stream pb.Porter_MoveFileServer) error {
	meta, ok := metadata.FromIncomingContext(stream.Context())
	if ok {
		log.Printf("file \"%s\" transform from %s", meta["name"][0], meta["address"][0])
	} else {
		return errors.New("file meta data incomplete could not be storage")
	}
	// The file's hash and name will be used as the source information for the file.
	name := meta["name"][0]
	hash := meta["hash"][0]
	method := meta["method"][0]
	var path string
	if method == "storage" {
		path = p.storageDirectory + "/" + hash
	}
	exist, err := utils.CheckPathExists(path)
	if err != nil {
		return err
	}
	if !exist {
		os.Mkdir(path, os.ModePerm)
		var fileBuffer []byte
		// The same hash file will be overwritten to forbiden some special situation.
		file, _ := os.OpenFile(path+"/"+name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		defer file.Close()
		for {
			content, err := stream.Recv()
			if err == io.EOF {
				end, _ := file.Seek(0, os.SEEK_END)
				_, err = file.WriteAt(fileBuffer, end)
				if err != nil {
					os.RemoveAll(path)
					return err
				}
				log.Printf("file which hash: \"%s\" transport success\n", hash)
				return stream.SendAndClose(&pb.MoveReply{Status: pb.MoveReply_SUCCESS})
			}
			if err != nil {
				os.RemoveAll(path)
				return err
			}
			fileBuffer = append(fileBuffer, content.Chunk...)
		}
	} else {
		log.Printf("file \"%s\" had been backuped in here\n", hash)
	}
	return nil
}
