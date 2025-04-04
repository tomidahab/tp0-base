import logging

class ProtocolHandler:
    def __init__(self, timeout=10):
        self.timeout = timeout

    def recv_exact(self, sock, num_bytes):
        buffer = bytearray()
        while len(buffer) < num_bytes:
            chunk = sock.recv(num_bytes - len(buffer))
            if not chunk:
                return None
            buffer.extend(chunk)
        return buffer

    def send_exact(self, sock, data):
        total_sent = 0
        while total_sent < len(data):
            sent = sock.send(data[total_sent:])
            if sent == 0:
                raise RuntimeError("Socket connection broken")
            total_sent += sent

    def recv_message_length(self, sock):
        length_bytes = self.recv_exact(sock, 2)
        if not length_bytes:
            raise ValueError("Failed to receive message length")
        return int.from_bytes(length_bytes, "big")

    def send_confirmation(self, sock, message_len):
        confirmation = message_len.to_bytes(2, "big")
        self.send_exact(sock, confirmation)

    def read_batch(self, sock):
        """
        Lee un batch de bets:
            1. Lee 2 bytes que le dan el tamaño del batch
            2. Lee esa cantidad de batchs
            3. Se fija si ese batch no tiene un END\n al final (en ese caso seria el ultimo batch)

        """
        message_length = self.recv_message_length(sock)
        batch_bytes = self.recv_exact(sock, message_length)
        if not batch_bytes:
            raise ValueError("Failed to receive complete batch message")
        batchMessage = batch_bytes.decode("utf-8")

        lines = batchMessage.strip().split("\n")
        lastBatch = False
        if len(lines) > 0 and lines[-1] == "END":
            lastBatch = True
            lines = lines[:-1]
            batchMessage = "\n".join(lines)

        self.send_confirmation(sock, len(lines))
        return batchMessage, lastBatch

    def send_winners(self, sock, documents):
        """
        Envía primero el número total de ganadores y luego, en un segundo mensaje, todos los documentos ganadores.
        (en el caso que no haya ganadores no manda el segundo mensaje)
        """
        n_winners = len(documents)
        try:
            self.send_exact(sock, n_winners.to_bytes(4, byteorder="big"))
            if n_winners != 0:
                msg = b''.join(int(document).to_bytes(4, byteorder="big") for document in documents)
                self.send_exact(sock, msg)
        except Exception as e:
            logging.error(f"Error sending winners: {e}")
