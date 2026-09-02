package safe_socket

import "io"

func SendAll(socket io.Writer, bytes []byte) error {
	for len(bytes) > 0 {
		written, err := socket.Write(bytes)
		if err != nil {
			return err
		}
		if written == 0 {
			return io.ErrShortWrite
		}
		bytes = bytes[written:]
	}
	return nil
}

func RecvAll(socket io.Reader, size int) ([]byte, error) {
	buff := make([]byte, size)
	read := 0
	for read < size {
		n, err := socket.Read(buff[read:])
		read += n
		if err != nil {
			return nil, err
		}
		if n == 0 {
			return nil, io.ErrNoProgress
		}
	}
	return buff, nil
}
