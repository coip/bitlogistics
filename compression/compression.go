package compression

import (
	"compress/gzip"
	"fmt"
	"io"
	"log"
)

const (
	buffersize = 4
	chansize   = 4
)

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

		rq     = make(chan int, chansize)
		wq     = make(chan int, chansize)
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
				d <- io.EOF
				c <- true
				break
			}
			a <- read
			written, e = gw.Write(buf[:read])
			if e != nil {
				log.Println("gzip::gw.Write()")
				d <- e
			}
			fmt.Printf("wrote %s\n", buf[:read])
			b <- written
		}
		gw.Flush()
		gw.Close()
		close(a)
		close(b)
		close(c)
		close(d)
		return
	}(rq, wq, signal, errorQueue)

	return GzJob{rchunks(rq), wchunks(wq), completionSignal(signal), errorQueue}
}

//Gunzip the data stored in src into dst
func Gunzip(w io.Writer, r io.Reader) GzJob {
	var (
		read, written int
		e             error

		buf = make([]byte, buffersize)

		rq     = make(chan int, chansize)
		wq     = make(chan int, chansize)
		signal = make(chan bool)

		errorQueue = make(chan error, 1)
	)
	// Write gzipped data from src to dest
	gr, err := gzip.NewReader(r)
	if err != nil {
		panic(err)
	}
	go func(a chan int, b chan int, c chan bool, d chan error) {
		for {
			read, e = gr.Read(buf)
			if e == io.EOF {
				d <- io.EOF
				c <- true
				break
			}
			a <- read
			written, e = w.Write(buf[:read])
			if e != nil {
				d <- e
			}
			fmt.Printf("wrote %s\n", buf[:read])
			b <- written
		}
		gr.Close()
		close(a)
		close(b)
		close(c)
		close(d)
		return
	}(rq, wq, signal, errorQueue)

	return GzJob{rchunks(rq), wchunks(wq), completionSignal(signal), errorQueue}
	// defer gr.Close()
	// data, err := ioutil.ReadAll(gr)
	// if err != nil {
	// 	return err
	// }
	// w.Write(data)
	// return nil
}
