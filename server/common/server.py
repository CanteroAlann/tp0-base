import socket
import logging
import signal
import struct

from common.protocol import ProtocolProcessedError, encode_response_message, receive_message
from common.utils import Bet, store_bets, load_bets, has_won

class Server:
    def __init__(self, port, listen_backlog):
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.running = True
        self.finished_agencies = set()
        self.TOTAL_AGENCIES = 5

        signal.signal(signal.SIGTERM, self.__handle_sigterm)

    def __handle_sigterm(self, signum, frame):
        logging.info('action: shutdown | result: in_progress')
        self.running = False

    def run(self):
        while self.running:
            try:
                client_sock = self.__accept_new_connection()
                self.__handle_client_connection(client_sock)
            except OSError:
                if not self.running:
                    break
        
        logging.info('action: shutdown | result: success')
        self._server_socket.close()

    def __handle_client_connection(self, client_sock):
        try:
            reading_bets = True
            while reading_bets:
                msg_type, data = receive_message(client_sock)

                if msg_type == 1:
                    payload_length, user_data = data
                    bets = [Bet(agency=d.agencia, first_name=d.nombre, last_name=d.apellido, 
                                document=d.documento, birthdate=d.nacimiento.isoformat(), number=d.numero) 
                            for d in user_data]
                    store_bets(bets)
                    
                    logging.info(f'action: apuesta_recibida | result: success | cantidad: {payload_length}')    
                    client_sock.sendall(f'Bets received: {payload_length} \n'.encode('utf-8'))
                    continue

                elif msg_type == 2:
                    agency = data
                    self.finished_agencies.add(agency)
                    client_sock.sendall(b'ACK\n')
                    reading_bets = False
                    
                    if len(self.finished_agencies) == self.TOTAL_AGENCIES:
                        logging.info('action: sorteo | result: success')

                elif msg_type == 3:
                    agency = data
                    
                    if len(self.finished_agencies) < self.TOTAL_AGENCIES:
                        logging.debug(f'action: query_winners | result: fail | error: Not all agencies have finished yet (finished: {len(self.finished_agencies)}/{self.TOTAL_AGENCIES})')
                        client_sock.sendall(b'\x00')

                    else:
                        winners_dni = []
                        for bet in load_bets():
                            if bet.agency == agency and has_won(bet):
                                winners_dni.append(bet.document)
                        
                        
                        response = bytearray([0x01])
                        response.extend(struct.pack('>I', len(winners_dni)))
                        for dni in winners_dni:
                            response.extend(struct.pack('>I', dni))
                        
                        client_sock.sendall(response)
                    reading_bets = False

        except Exception as e:
            logging.error(f'action: handle_message | result: fail | error: {e}')

        finally:
            client_sock.close()

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
