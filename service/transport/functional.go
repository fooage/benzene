package transport

import (
	"context"
	"io"
	"log"
	"os"
	"time"

	"github.com/fooage/messier/proto/pb"
	"github.com/fooage/messier/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AdjustStorage function periodically adjusts the position of the file based
// on the hash relationship between the nodes and remove it in this node.
func (p *Porter) AdjustStorage() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	trick := time.NewTicker(p.duration)
	defer trick.Stop()
	for {
		conn, err := grpc.Dial(p.info.NextAddress, opts...)
		if err != nil {
			log.Printf("the next node can't be dial: %v\n", err)
		}
		client := pb.NewPorterClient(conn)
		list, err := utils.LocalFolderList(p.storageDirectory)
		if err != nil {
			log.Fatalf("there's a directory read error: %v\n", err)
		}
		if p.info.CurrHash < p.info.NextHash {
			for _, hash := range list {
				if hash > p.info.NextHash {
					log.Printf("begin to transport file \"%s\"\n", hash)
					err := p.transportStorageFile(client, hash)
					if err != nil {
						log.Printf("transport file but failed: %v", err)
					} else {
						os.RemoveAll(p.storageDirectory + "/" + hash)
					}
				}
			}
		} else {
			// Make special judgments for tail nodes in the ring.
			for _, hash := range list {
				if hash > p.info.NextHash && hash < p.info.CurrHash {
					log.Printf("begin to transport file \"%s\"\n", hash)
					err := p.transportStorageFile(client, hash)
					if err != nil {
						log.Printf("try to transport file but failed: %v", err)
					} else {
						os.RemoveAll(p.storageDirectory + "/" + hash)
					}
				}
			}
		}
		// Continue to adjust the file position after a certain interval.
		conn.Close()
		<-trick.C
	}
}

// This function used to transport file in node's storage directory.
func (p *Porter) transportStorageFile(client pb.PorterClient, hash string) error {
	files, err := utils.LocalFileList(p.storageDirectory + "/" + hash)
	if err != nil {
		return err
	}
	// TODO: Further define a new way of selecting file.
	file, err := os.Open(p.storageDirectory + "/" + hash + "/" + files[0])
	if err != nil {
		return err
	}
	defer file.Close()
	meta := metadata.New(map[string]string{
		"hash":    hash,
		"name":    files[0],
		"method":  "storage",
		"address": p.info.CurrAddress})
	stream, err := client.MoveFile(metadata.NewOutgoingContext(context.Background(), meta))
	if err != nil {
		return err
	}
	for {
		// send file once by the chunk size
		chunk := make([]byte, p.chunkSize*1024)
		length, err := file.Read(chunk)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if length < len(chunk) {
			chunk = chunk[:length]
		}
		stream.Send(&pb.MoveRequest{Chunk: chunk})
	}
	_, err = stream.CloseAndRecv()
	if err != nil {
		return err
	}
	return nil
}
