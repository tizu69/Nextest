package api

import "net/http"

// peekRW is a http.ResponseWriter that copies the response into a buffer
// so that it can be read back.
type peekRW struct {
	http.ResponseWriter
	buffer []byte
}

func newPeekRW(w http.ResponseWriter) *peekRW {
	return &peekRW{ResponseWriter: w, buffer: make([]byte, 0)}
}

func (rw *peekRW) Write(b []byte) (int, error) {
	rw.buffer = append(rw.buffer, b...)
	return rw.ResponseWriter.Write(b)
}

func (rw *peekRW) Peek() []byte {
	return rw.buffer
}
