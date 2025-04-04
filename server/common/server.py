import socket
import logging
import signal
import os
import threading
from .protocol import ProtocolHandler
from .business import parse_batch, process_bets, find_winners

CLIENT_TOTAL = int(os.environ.get("CLIENT_TOTAL", 5))

class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._client_sockets = []
        self.running = True
        self.ended_clients = 0
        self.open_sockets = {}
        self._lock = threading.Lock()
        self.clients_done = threading.Condition(self._lock)
        self._threads = []
        self.protocol = ProtocolHandler()
        signal.signal(signal.SIGTERM, self.__shutdown_server)

    def run(self):
        accept_thread = threading.Thread(target=self.__accept_connections)
        accept_thread.start()
        with self.clients_done:
            while self.ended_clients < CLIENT_TOTAL:
                self.clients_done.wait()
        winners_dic = find_winners()
        self.__send_winners(winners_dic)
        self.__close_sockets()
        self.running = False
        self.__close_server()
        for t in self._threads:
            t.join()
        accept_thread.join()

    def __accept_connections(self):
        logging.info('action: accept_connections | result: in_progress')
        while self.running:
            try:
                client_sock, addr = self._server_socket.accept()
                logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
                self._client_sockets.append(client_sock)
                t = threading.Thread(target=self.__handle_client_connection, args=(client_sock,))
                t.start()
                self._threads.append(t)
            except OSError:
                break

    def __handle_client_connection(self, client_sock):
        allBets = []
        agency = None
        try:
            while True:
                batchMessage, lastBatch = self.protocol.read_batch(client_sock)
                betsInBatch = parse_batch(batchMessage)
                allBets.extend(betsInBatch)
                if lastBatch:
                    break
            with self._lock:
                process_bets(allBets)
            agency = allBets[0].agency
        except Exception as e:
            logging.error(f"action: handle_client_connection | result: fail | error: {e}")
        finally:
            with self.clients_done:
                self.open_sockets[agency] = client_sock
                exit(-1)
                self.ended_clients += 1
                self.clients_done.notify_all()
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(allBets)}")

    def __send_winners(self, winners_dic):
        for agency, documents in winners_dic.items():
            if agency in self.open_sockets:
                sock = self.open_sockets[agency]
                self.protocol.send_winners(sock, documents)
            else:
                logging.warning(f"action: send_winners | agency: {agency} | result: fail | reason: no_socket")

    def __close_sockets(self):
        for sock in self.open_sockets.values():
            sock.close()

    def __close_server(self):
        logging.info('action: shutdown_server | result: in_progress')
        self.running = False
        try:
            self._server_socket.close()
            logging.info('action: shutdown_server | result: success')
        except OSError as e:
            logging.error(f'action: shutdown_server | result: fail | error: {e}')

    def __shutdown_server(self, signum, frame):
        self.__close_server()
