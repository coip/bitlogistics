package main

import (
	"bitlogistics/compression"
	"bitlogistics/crypto"
	"bytes"
	"fmt"
	"time"
	// _ "net/http/pprof"
	"os"
	// "runtime"
)

const fmtstr = "[pipeline-stage-completed=(%s) (in %s) => len(\n%q\n) == %d;\n]"

const str = `Everything which is in any way beautiful is beautiful in itself, and terminates in itself, not having praise as part of itself. Neither worse then, nor better, is a thing made by being praised. I affirm this also of the things which are called beautiful by the vulgar; for example, material things and works of art. That which is really beautiful has no need of anything; not more than law, not more than truth, not more than benevolence or modesty. Which of these things is beautiful because it is praised, or spoiled by being blamed? Is such a thing as an emerald made worse than it was, if it is not praised? Or gold, ivory, purple, a lyre, a little knife, a flower, a shrub?`

var (
	gzippedbuf bytes.Buffer
	dzippedbuf bytes.Buffer

	encryptedgzippedbuf *bytes.Buffer
	decryptedgzippedbuf *bytes.Buffer
)

func main() {
	start := time.Now()
	// runtime.GOMAXPROCS(4)
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()
	fmt.Printf(fmtstr, "unencrypted", time.Since(start), str, len(str))

	in := compression.Gzip(&gzippedbuf, bytes.NewBuffer([]byte(str)))
	in.Observe(os.Stdout)
	<-in.Done

	fmt.Printf(fmtstr, "gzipped", time.Since(start), gzippedbuf.Bytes(), gzippedbuf.Len())

	encryptedgzippedbuf = crypto.Encrypt(gzippedbuf.Bytes())
	fmt.Printf(fmtstr, "gzipped encrypted", time.Since(start), encryptedgzippedbuf.Bytes(), encryptedgzippedbuf.Len())

	decryptedgzippedbuf = crypto.Decrypt(encryptedgzippedbuf.Bytes())
	fmt.Printf(fmtstr, "gzipped decrypted", time.Since(start), encryptedgzippedbuf.Bytes(), encryptedgzippedbuf.Len())

	out := compression.Gunzip(&dzippedbuf, decryptedgzippedbuf) //bytes.NewReader(gzippedbuf.Bytes())) //
	out.Observe(os.Stdout)
	<-out.Done
	fmt.Printf(fmtstr, "dzipped decrypted", time.Since(start), dzippedbuf.Bytes(), dzippedbuf.Len())

	// <-time.After(35 * time.Second)
}
