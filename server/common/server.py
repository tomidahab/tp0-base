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
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        """try:
            # TODO: Modify the receive to avoid short-reads
            msg = client_sock.recv(1024).rstrip().decode('utf-8')
            addr = client_sock.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
            # TODO: Modify the send to avoid short-writes
            client_sock.send("{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            client_sock.close()"""
        try:
            addr = client_sock.getpeername()
            # logging.info(f'action: handle_client_connection | client_ip: {addr[0]} | result: in_progress')

            message_length = self.__recv_message_lenght(client_sock)

            bet = self.__recv_bet(client_sock, message_length)

            store_bets([bet])
            
            logging.info(f'action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}')

            # Send back the length as confirmation (2 bytes, big-endian)
            confirmation = message_length.to_bytes(2, "big")
            self.__send_exact(client_sock, confirmation)
        except Exception as e:
            logging.error(f'action: handle_client_connection | result: fail | error: {e}')
        finally:
            client_sock.close()

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