import socket
import logging
import signal
import os
from common.utils import Bet, load_bets, store_bets, has_won

CLIENT_TOTAL = int(os.environ.get("CLIENT_TOTAL", 5))
TIMEOUT = int(os.environ.get("TIMEOUT", 10))

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_sockets = []
        self.running = True # to stop the server loop during shutdown
        self.ended_clients = 0
        self.open_sockets = {}

        signal.signal(signal.SIGTERM, self.__shutdown_server)

    def run(self):
        """
        Server accepts new connection and manages them one by one.
        Once all clients have been processed, it finds the winners, and
        shuts down itself.
        """

        while self.running:
            try:
                client_sock = self.__accept_new_connection()
                if client_sock == None:
                    self.__close_server()
                    return
                self._client_sockets.append(client_sock)
                self.__handle_client_connection(client_sock)

                if self.ended_clients == CLIENT_TOTAL:
                    winners_dic = self.__find_winners()
                    self.__send_winners(winners_dic)
                    self.__close_sockets()
            except:
                if not self.running:
                    return
            

    def __close_sockets(self):
        for socket in self.open_sockets:
            socket.close()


    def __send_winners(self, winners):
        """
        Envía primero el número total de ganadores y luego, en un segundo mensaje, todos los documentos ganadores.
        (en el caso que no haya ganadores no manda el segundo mensaje)
        """
        for agency, documents in winners.items():
            n_winners = len(documents) 

            if agency in self.open_sockets:
                sock = self.open_sockets[agency]
                try:
                    self.__send_exact(sock, n_winners.to_bytes(4, byteorder="big"))
                    if n_winners != 0:
                        msg = b''.join(int(document).to_bytes(4, byteorder="big") for document in documents)
                        self.__send_exact(sock, msg)

                except Exception as e:
                    logging.error(f"action: send_winners | agency: {agency} | result: fail | error: {e}")
            else:
                logging.warning(f"action: send_winners | agency: {agency} | result: fail | reason: no_socket")

            

    def __find_winners(self):
        logging.info("action: sorteo | result: success")

        winners_dic = {}

        bets = load_bets()

        for bet in bets:
            if has_won(bet):
                agency = bet.agency
                if agency in winners_dic:
                    winners_dic[agency].append(bet.document)
                else:
                    winners_dic[agency] = [bet.document]
        
        return winners_dic
        

    def __handle_client_connection(self, client_sock):
        """
        Maneja conexion del cliente
        1. Lee el batch
        2. Transforma cada line en un Bet
        3. Se fija si ese batch no era el ultimo, en caso que 
        no vuelve a 1, sino termina y guarda todos los bets.

        """
        allBets = []
        try:
            while True:
                batchMessage, lastBatch = self.__read_batch(client_sock)
                betsInBatch = self.__parse_batch(batchMessage)
                allBets.extend(betsInBatch)
                if lastBatch:
                    break

            store_bets(allBets)

            agency = allBets[0].agency
        except Exception as e:
            logging.error(f"action: handle_client_connection | result: fail | error: {e}")
        finally:
            #client_sock.close()
            self.open_sockets[agency] = client_sock

        # Generar log consolidado al final
        logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(allBets)}")


    def __read_batch(self, client_sock):
        """
        Lee un batch de bets:
            1. Lee 2 bytes que le dan el tamaño del batch
            2. Lee esa cantidad de batchs
            3. Se fija si ese batch no tiene un END\n al final (en ese caso seria el ultimo batch)

        """
        message_length = self.__recv_message_lenght(client_sock)
        
        batch_bytes = self.__recv_exact(client_sock, message_length)
        if not batch_bytes:
            raise ValueError("Failed to receive complete batch message")
        batchMessage = batch_bytes.decode("utf-8")

        lines = batchMessage.strip().split("\n")
        lastBatch = False
        if len(lines) > 0 and lines[-1] == "END":
            lastBatch = True
            lines = lines[:-1]  # Remover la línea 'END'
            batchMessage = "\n".join(lines)
            self.ended_clients += 1

        self.__send_confirmation(client_sock, len(lines))
        
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
        Accept new connections.
        Blocks until a connection to a client is made or timeout is reached.
        Then prints and returns the new socket.
        """
        logging.info('action: accept_connections | result: in_progress')
        self._server_socket.settimeout(TIMEOUT)
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            self._server_socket.settimeout(None)
            return c
        except socket.timeout:
            return None


    def __close_server(self):
        logging.info('action: shutdown_server | result: in_progress')
        self.running = False
        try:
            self._server_socket.close()

            logging.info('action: shutdown_server | result: success')
        except OSError as e:
            logging.error(f'action: shutdown_server | result: fail | error: {e}')
            exit(-1)


    def __shutdown_server(self, signum, frame):
        self.__close_server()
