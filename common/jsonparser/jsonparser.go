package jsonparser

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

type Parser struct {
	buffSize  int
	chunkSize int
}

func New(buff, chunk int) Parser{
	return Parser{
		buffSize:  buff,
		chunkSize: chunk,
	}
}

func(p Parser) ParseFile(file io.Reader) ( <-chan map[string]interface{}, <-chan error, error) {
	errChan := make(chan error)
	pr, pw := io.Pipe()
	go func() {
		// close the writer, so the reader knows there's no more data
		defer func() {
			if err := pw.Close(); err != nil {
				fmt.Printf("error closing pipe writer. err: %v", err)
			}
		}()

		buf := make([]byte, int64(p.buffSize))
		defer func() {
			buf = nil
		}()
		for {
			chunk, err := file.Read(buf)
			if err != nil && err != io.EOF {
				errChan <- err
				fmt.Printf("error reading file. err: %v", err)
				break
			}
			if chunk == 0 {
				break
			}

			if _, err := pw.Write(buf[:chunk]); err != nil{
				return
			}
		}
	}()

	dec := json.NewDecoder(pr)
	// read open bracket
	if _, err := dec.Token(); err != nil {
		fmt.Printf("error reading open json bracket. err: %v", err)
		return nil, nil, err
	}

	readChan := make(chan map[string]interface{})
	go func() {
		chunk := make(map[string]interface{}, p.chunkSize)
		var parseErr error
		// while the map contains keys
		for dec.More() {
			var m map[string]interface{}

			// read the key
			key, err := dec.Token()
			if err != nil {
				parseErr = err
				fmt.Printf("error reading struct key. err: %v", err)
				errChan <- err
				break
			}

			strKey, ok := key.(string)
			if !ok {
				err = errors.New("error reading struct key. key is not a string")
				parseErr = err
				errChan <- err
				break
			}

			// read the data
			if err = dec.Decode(&m); err != nil {
				parseErr = err
				fmt.Printf("error reading struct key value. err: %v", err)
				errChan <- err
				break
			}

			chunk[strKey] = m
			if len(chunk) == p.chunkSize {
				readChan <- chunk
				chunk = make(map[string]interface{}, p.chunkSize)
			}
		}

		if parseErr != nil {
			return
		}

		close(readChan)
	}()

	return readChan, errChan, nil
}
