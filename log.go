package gostellar

import (
	"context"
	"fmt"
	"github.com/MilosSimic/pipes"
	sPb "github.com/c12s/scheme/stellar"
	"github.com/golang/protobuf/proto"
	"io/ioutil"
	"os"
	"path"
)

func writeBytes(file *os.File, bytes []byte) error {
	_, err := file.Write(bytes)
	if err != nil {
		return err
	}
	return nil
}

func readFileBytes(path string) ([]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func CheckCollectorDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func CollectTraces(ctx context.Context, path string) chan interface{} {
	files, err := ioutil.ReadDir(path)
	f := []interface{}{}
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, item := range files {
		f = append(f, item.Name())
	}

	pipe1 := pipes.Pipe(ctx, pipes.Source(ctx, f), func(data []interface{}) interface{} {
		for _, elem := range data {
			name := elem.(string)
			f := fmt.Sprintf("%s%s", path, name)
			b, err := readFileBytes(f)
			if err != nil {
				return err
			}
			return b
		}
		return nil
	})

	pipe2 := pipes.Pipe(ctx, pipe1, func(data []interface{}) interface{} {
		for _, elem := range data {
			switch value := elem.(type) {
			case []byte:
				s := &sPb.Span{}
				err = proto.Unmarshal(value, s)
				if err != nil {
					return err
				}
				return s
			case error:
				return value
			}
		}
		return nil
	})

	return pipe2
}

func ClearDir(dir string) error {
	names, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, entery := range names {
		os.RemoveAll(path.Join([]string{dir, entery.Name()}...))
	}
	return nil
}

func Log(data []byte, traceid, spanid string) error {
	file, err := os.Create(fmt.Sprintf("logs/%s_%s.log", traceid, spanid))
	defer file.Close()
	if err != nil {
		return err
	}

	err = writeBytes(file, data)
	if err != nil {
		return err
	}

	return nil
}
