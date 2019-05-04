package main

import (
	"bitlogistics/compression"
	"bitlogistics/crypto"
	"bytes"
	"log"
	"runtime"
)

const fmtstr = "stage=%s => len(\n%q\n) == %d;\n"

const str = `Everything which is in any way beautiful is beautiful in itself, and terminates in itself, not having praise as part of itself. 
Neither worse then, nor better, is a thing made by being praised. 

I affirm this also of the things which are called beautiful by the vulgar; for example, material things and works of art. 

That which is really beautiful has no need of anything; 
not more than law, not more than truth, not more than benevolence or modesty.

Which of these things is beautiful because it is praised, or spoiled by being blamed?

Is such a thing as an emerald made worse than it was, if it is not praised? 
Or gold, ivory, purple, a lyre, a little knife, a flower, a shrub?`

var (
	gzippedbuf bytes.Buffer
	dzippedbuf bytes.Buffer

	encryptedgzippedbuf *bytes.Buffer
	decryptedgzippedbuf *bytes.Buffer
)

func main() {

	runtime.GOMAXPROCS(1)

	log.Printf(fmtstr, "unencrypted", str, len(str))

	in := compression.Gzip(&gzippedbuf, bytes.NewBuffer([]byte(str)))
	in.Observe()
	<-in.Done

	log.Printf(fmtstr, "gzipped", gzippedbuf.Bytes(), gzippedbuf.Len())

	encryptedgzippedbuf = crypto.Encrypt(gzippedbuf.Bytes())
	log.Printf(fmtstr, "gzipped encrypted", encryptedgzippedbuf.Bytes(), encryptedgzippedbuf.Len())

	decryptedgzippedbuf = crypto.Decrypt(encryptedgzippedbuf.Bytes())
	log.Printf(fmtstr, "gzipped decrypted", encryptedgzippedbuf.Bytes(), encryptedgzippedbuf.Len())

	compression.Gunzip(&dzippedbuf, decryptedgzippedbuf)
	log.Printf(fmtstr, "dzipped decrypted", dzippedbuf.Bytes(), dzippedbuf.Len())

}
