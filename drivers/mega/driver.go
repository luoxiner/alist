package mega

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/alist-org/alist/v3/internal/driver"
	"github.com/alist-org/alist/v3/internal/errs"
	"github.com/alist-org/alist/v3/internal/model"
	"github.com/alist-org/alist/v3/pkg/chanio"
	log "github.com/sirupsen/logrus"
	"github.com/t3rm1n4l/go-mega"
)

type Mega struct {
	model.Storage
	Addition
	c *mega.Mega
}

func (d *Mega) Config() driver.Config {
	return config
}

func (d *Mega) GetAddition() driver.Additional {
	return &d.Addition
}

func (d *Mega) Init(ctx context.Context) error {
	d.c = mega.New()
	return d.c.Login(d.Email, d.Password)
}

func (d *Mega) Drop(ctx context.Context) error {
	return nil
}

func (d *Mega) List(ctx context.Context, dir model.Obj, args model.ListArgs) ([]model.Obj, error) {
	if node, ok := dir.(*MegaNode); ok {
		nodes, err := d.c.FS.GetChildren(node.Node)
		if err != nil {
			return nil, err
		}
		res := make([]model.Obj, 0)
		for i := range nodes {
			n := nodes[i]
			if n.GetType() == mega.FILE || n.GetType() == mega.FOLDER {
				res = append(res, &MegaNode{n})
			}
		}
		return res, nil
	}
	log.Errorf("can't convert: %+v", dir)
	return nil, fmt.Errorf("unable to convert dir to mega node")
}

func (d *Mega) Get(ctx context.Context, path string) (model.Obj, error) {
	if path == "/" {
		n := d.c.FS.GetRoot()
		log.Debugf("mega root: %+v", *n)
		return &MegaNode{n}, nil
	}
	return nil, errs.NotSupport
}

func (d *Mega) Link(ctx context.Context, file model.Obj, args model.LinkArgs) (*model.Link, error) {
	if node, ok := file.(*MegaNode); ok {
		//link, err := d.c.Link(node.Node, true)
		//if err != nil {
		//	return nil, err
		//}
		//return &model.Link{URL: link}, nil
		down, err := d.c.NewDownload(node.Node)
		if err != nil {
			return nil, err
		}
		//u := down.GetResourceUrl()
		//u = strings.Replace(u, "http", "https", 1)
		//return &model.Link{URL: u}, nil
		c := chanio.New()
		go func() {
			defer func() {
				_ = recover()
			}()
			log.Debugf("chunk size: %d", down.Chunks())
			for id := 0; id < down.Chunks(); id++ {
				chunk, err := down.DownloadChunk(id)
				if err != nil {
					log.Errorf("mega down: %+v", err)
					return
				}
				log.Debugf("id: %d,len: %d", id, len(chunk))
				//_, _, err = down.ChunkLocation(id)
				//if err != nil {
				//	log.Errorf("mega down: %+v", err)
				//	return
				//}
				//_, err = c.Write(chunk)
				_, err = c.Write(chunk)
			}
			err := c.Close()
			if err != nil {
				log.Errorf("mega down: %+v", err)
			}
		}()
		return &model.Link{Data: c}, nil
	}
	return nil, fmt.Errorf("unable to convert dir to mega node")
}

func (d *Mega) MakeDir(ctx context.Context, parentDir model.Obj, dirName string) error {
	if parentNode, ok := parentDir.(*MegaNode); ok {
		_, err := d.c.CreateDir(dirName, parentNode.Node)
		return err
	}
	return fmt.Errorf("unable to convert dir to mega node")
}

func (d *Mega) Move(ctx context.Context, srcObj, dstDir model.Obj) error {
	if srcNode, ok := srcObj.(*MegaNode); ok {
		if dstNode, ok := dstDir.(*MegaNode); ok {
			return d.c.Move(srcNode.Node, dstNode.Node)
		}
	}
	return fmt.Errorf("unable to convert dir to mega node")
}

func (d *Mega) Rename(ctx context.Context, srcObj model.Obj, newName string) error {
	if srcNode, ok := srcObj.(*MegaNode); ok {
		return d.c.Rename(srcNode.Node, newName)
	}
	return fmt.Errorf("unable to convert dir to mega node")
}

func (d *Mega) Copy(ctx context.Context, srcObj, dstDir model.Obj) error {
	return errs.NotImplement
}

func (d *Mega) Remove(ctx context.Context, obj model.Obj) error {
	if node, ok := obj.(*MegaNode); ok {
		return d.c.Delete(node.Node, false)
	}
	return fmt.Errorf("unable to convert dir to mega node")
}

func (d *Mega) Put(ctx context.Context, dstDir model.Obj, stream model.FileStreamer, up driver.UpdateProgress) error {
	if dstNode, ok := dstDir.(*MegaNode); ok {
		u, err := d.c.NewUpload(dstNode.Node, stream.GetName(), stream.GetSize())
		if err != nil {
			return err
		}

		for id := 0; id < u.Chunks(); id++ {
			_, chkSize, err := u.ChunkLocation(id)
			if err != nil {
				return err
			}
			chunk := make([]byte, chkSize)
			n, err := io.ReadFull(stream, chunk)
			if err != nil && err != io.EOF {
				return err
			}
			if n != len(chunk) {
				return errors.New("chunk too short")
			}

			err = u.UploadChunk(id, chunk)
			if err != nil {
				return err
			}
			up(id * 100 / u.Chunks())
		}

		_, err = u.Finish()
		return err
	}
	return fmt.Errorf("unable to convert dir to mega node")
}

//func (d *Mega) Other(ctx context.Context, args model.OtherArgs) (interface{}, error) {
//	return nil, errs.NotSupport
//}

var _ driver.Driver = (*Mega)(nil)
