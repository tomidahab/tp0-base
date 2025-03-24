import socket
import logging
import signal
from common.utils import Bet, load_bets, store_bets


class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_sockets = []
        self.running = True # to stop the server loop during shutdown

        signal.signal(signal.SIGTERM, self.__shutdown_server)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self.running:
            try:
                client_sock = self.__accept_new_connection()
                self._client_sockets.append(client_sock)
                self.__handle_client_connection(client_sock)
            except:
                if not self.running:
                    return
            

    def __handle_client_connection(self, client_sock):
        """
        Procesa la conexión de un cliente leyendo batches de apuestas:
        1. Lee el tamaño del batch.
        2. Lee el contenido (dividido por líneas) y, si se encuentra 'END',
            significa que es el último batch.
        3. Parsea cada línea en un objeto Bet y los acumula.
        4. Si no se recibió 'END', se repite la lectura.
        5. Al finalizar, se almacenan todas las apuestas y se envía una confirmación.
        """
        allBets = []  # Acumula todas las apuestas recibidas
        try:
            # Bucle de lectura de batches
            while True:
                batchMessage, lastBatch = self.__read_batch(client_sock)
                betsInBatch = self.__parse_batch(batchMessage)
                allBets.extend(betsInBatch)
                if lastBatch:
                    break

            store_bets(allBets)
        except Exception as e:
            logging.error(f"action: handle_client_connection | result: fail | error: {e}")
        finally:
            client_sock.close()

        # Generar log consolidado al final
        logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(allBets)}")


    def __read_batch(self, client_sock):
        """
        Lee un batch de apuestas:
          - Lee el header de 2 bytes para obtener la longitud del batch.
          - Lee el mensaje completo y lo decodifica.
          - Si la última línea es 'END', se indica que es el último batch y se remueve esa línea.
        Devuelve: (batchMessage, lastBatch)
        """
        #logging.info("DELETE por recibir len")
        length_bytes = self.__recv_exact(client_sock, 2)
        if not length_bytes:
            raise ValueError("Failed to receive message length")
        message_length = int.from_bytes(length_bytes, "big")
        #logging.info("DELETE len = " + str(message_length))
        
        batch_bytes = self.__recv_exact(client_sock, message_length)
        if not batch_bytes:
            raise ValueError("Failed to receive complete batch message")
        batchMessage = batch_bytes.decode("utf-8")

        #logging.info("recibi msg" + str(batchMessage))

        lines = batchMessage.strip().split("\n")
        #logging.info(str(lines))
        lastBatch = False
        if len(lines) > 0 and lines[-1] == "END":
            lastBatch = True
            lines = lines[:-1]  # Remover la línea 'END'
            batchMessage = "\n".join(lines)

        confirmation = len(lines).to_bytes(2, "big")
        self.__send_exact(client_sock, confirmation)
        
        return batchMessage, lastBatch

    def __parse_batch(self, batchMessage):
        """
        Parsea el contenido de un batch (cadena de texto) en una lista de objetos Bet.
        Se espera que cada línea tenga 6 campos separados por comas:
            agency, first_name, last_name, document, birthdate, number
        """
        bets = []
        lines = batchMessage.strip().split("\n")
        for line in lines:
            trimmed = line.strip()
            if trimmed == "":
                continue
            parts = trimmed.split(",")
            if len(parts) != 6:
                raise ValueError(f"Invalid bet format: {line}")
            betObj = Bet(parts[0], parts[1], parts[2], parts[3], parts[4], parts[5])
            bets.append(betObj)
        return bets


    def __recv_message_lenght(self, client_sock):
        length_bytes = self.__recv_exact(client_sock, 2)
        if not length_bytes:
            raise ValueError("Failed to receive message length")

        return int.from_bytes(length_bytes, "big")

    def __recv_bet(self, client_sock, message_lenght):
        message = self.__recv_exact(client_sock, message_lenght).decode('utf-8')
        #logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {message}')

        agency, first_name, last_name, document, birthdate, number = message.split(",")
        return Bet(agency, first_name, last_name, document, birthdate, number)

    def __send_confirmation(self, client_sock, message_len):
        confirmation = message_len.to_bytes(2, "big")
        self.__send_exact(client_sock, confirmation)

    def __recv_exact(self, sock, num_bytes):
        """
        Receive an exact number of bytes from the socket.
        """
        buffer = bytearray()
        while len(buffer) < num_bytes:
            chunk = sock.recv(num_bytes - len(buffer))
            if not chunk:
                return None
            buffer.extend(chunk)
        return buffer
    
    def __send_exact(self, sock, data):
        """
        Send all bytes of the data to the socket.
        """
        total_sent = 0
        while total_sent < len(data):
            sent = sock.send(data[total_sent:])
            if sent == 0:
                raise RuntimeError("Socket connection broken")
            total_sent += sent

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def __shutdown_server(self, signum, frame):
        logging.info('action: shutdown_server | result: in_progress')
        self.running = False
        try:
            self._server_socket.close()
            for client_socket in self._client_sockets:
                client_socket.close()

            logging.info('action: shutdown_server | result: success')
        except OSError as e:
            logging.error(f'action: shutdown_server | result: fail | error: {e}')
            exit(-1)