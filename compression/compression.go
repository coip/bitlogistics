package compression

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"log"
)

const buffersize = 8

type (
	//Rchunks transports the size of bytes a Read() operation returns, from the consuming goroutine to the caller asynchronously.
	rchunks <-chan int
	//Wchunks transports the size of bytes a Write() operation returns, from the consuming goroutine to the caller asynchronously.
	wchunks <-chan int
	//CompletionSignal receives true on complete. a false indicates an underlying irrecoverable failure has been detected for the processing.
	completionSignal <-chan bool
)

//GzJob is returned from Gzip() to enable non-blocking GzJob management.
type GzJob struct {
	R    rchunks
	W    wchunks
	Done completionSignal
	E    <-chan error
}

//Observe enables observation of a GzJob once it has began.
func (j GzJob) Observe() {
	var (
		read, written int
		err           error
		success       bool
	)
	go func() {
		for {
			select {
			case read = <-j.R:
				log.Printf("read %d bytes", read)
			case written = <-j.W:
				log.Printf("written %d bytes", written)
			case err = <-j.E:
				if err != io.EOF {
					log.Fatalf("error detected while observing compression: %+v", err)
					return
				}
				log.Println("EOF reached!")
				return
			case success = <-j.Done:
				log.Printf("success: %t", success)
			}
		}
	}()
}

//Gzip the data stored in src into dst
func Gzip(w io.Writer, r io.Reader) GzJob {
	var (
		read, written int
		e             error

		buf = make([]byte, buffersize)

		rq     = make(chan int, 1000)
		wq     = make(chan int, 1000)
		signal = make(chan bool)

		errorQueue = make(chan error, 1)
	)

	gw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
	if err != nil {
		panic(err)
	}

	go func(a chan int, b chan int, c chan bool, d chan error) {
		for {
			read, e = r.Read(buf)
			if e == io.EOF {
				c <- true
				d <- io.EOF
				gw.Close()
				close(a)
				close(b)
				close(c)
				close(d)
				return
			}
			a <- read
			written, e = gw.Write(buf)
			if e != nil {
				d <- e
			}
			b <- written
		}
	}(rq, wq, signal, errorQueue)

	return GzJob{rchunks(rq), wchunks(wq), completionSignal(signal), errorQueue}
}

//Gunzip the data stored in src into dst
func Gunzip(w io.Writer, r *bytes.Buffer) error {
	// Write gzipped data from src to dest
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
