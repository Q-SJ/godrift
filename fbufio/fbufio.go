package fbufio

import (
	"bytes"
	"bufio"
	. "../log"
)

type FReader struct {
	*bufio.Reader
}

const sizePerPeek = 4096

func (b *FReader) ReadAfterAll(delims map[int][]byte) (index int, err error) {
	count := 0
	skip := 0
	for k, v := range delims {
		Logger.Debugf("%d-->%x", k, v)
		if skip < len(v) {
			skip = len(v)
		}
	}
	skip = sizePerPeek - (skip - 1)

	for {
		Logger.Debugf("Have peeked for %d bytes...", count*sizePerPeek)
		buffer, err := b.Peek(sizePerPeek)
		if err != nil {
			Logger.Warning(err)
			return -1, err
		}

		for i, delim := range delims {
			index := bytes.Index(buffer, delim)
			if index >= 0 {
				_, err = b.Discard(index)
				if err != nil {
					Logger.Warning(err)
					return -1, err
				}
				Logger.Infof("Find the specific bytes: %d", i)
				Logger.Debugf("The index is %d.", count*sizePerPeek+index)
				return i, nil
			}
		}

		_, err = b.Discard(skip)
		if err != nil {
			Logger.Warning(err)
			return -1, err
		}

		count++
	}
}

//ReadAfter skip the bytes before the first occurrence delim int the input
func (b *FReader) ReadAfter(delim []byte) (err error) {
	count := 0
	skip := sizePerPeek - (len(delim) - 1)
	for {
		Logger.Debugf("Have peeked for %d bytes...", count*sizePerPeek)
		buffer, err := b.Peek(sizePerPeek)
		if err != nil {
			Logger.Warning(err)
			return err
		}
		//Logger.Debug(delim)
		//Logger.Debug(buffer)
		index := bytes.Index(buffer, delim)
		if index >= 0 {
			_, err = b.Discard(index)
			if err != nil {
				Logger.Warning(err)
				return err
			}
			Logger.Info("Find the specific bytes!")
			Logger.Debugf("The index is %d.", count*sizePerPeek+index)
			return nil
		} else {
			_, err = b.Discard(skip)
			if err != nil {
				Logger.Warning(err)
				return nil
			}
		}

		count++
	}
	return nil
}

func NewReader(reader *bufio.Reader) (*FReader) {
	r := new(FReader)
	r.Reader = reader
	return r
}
