package compression

import (
	"compress/gzip"
	"io"
)

const (
	buffersize = 1
	chansize   = 1

	rs = "r"
	ws = "w"
	be = ""

	eof = "[EOF]"
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
func (j GzJob) Observe(w io.Writer) {
	var (
		read, written int
		err           error
	)
	go func() {
		for {
			select {
			case read = <-j.R:
				w.Write([]byte(rs + string(rune(read+48)))) //+ be))
			case written = <-j.W:
				w.Write([]byte(ws + string(rune(written+48)))) //+ be))
			case err = <-j.E:
				if err != io.EOF {
					w.Write([]byte(err.Error()))
					panic(err)
				}
				w.Write([]byte(eof))
				return
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
				break
			}
			a <- read
			written, e = gw.Write(buf[:read])
			if e != nil {
				d <- e
			}
			b <- written
		}
		gw.Flush()
		gw.Close()
		c <- true
		d <- io.EOF
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
				break
			}
			a <- read
			written, e = w.Write(buf[:read])
			if e != nil {
				d <- e
			}
			b <- written
		}
		c <- true
		d <- io.EOF
		gr.Close()
		close(a)
		close(b)
		close(c)
		close(d)
		return
	}(rq, wq, signal, errorQueue)

	return GzJob{rchunks(rq), wchunks(wq), completionSignal(signal), errorQueue}

}
