package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
)

//Gzip the data from in r into w
func Gzip(w io.Writer, r *bytes.Buffer) (int, error) {
	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		panic(err)
	}
	defer gw.Close()
	return gw.Write(r.Bytes())
}

//Gunzip the data from in r into w
func Gunzip(w io.Writer, r *bytes.Buffer) error {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gr.Close()
	data, err := ioutil.ReadAll(gr)
	if err != nil {
		return err
	}
	w.Write(data)
	return nil
}
