from dataclasses import dataclass
from datetime import date
import struct
import logging


LENGTH_PREFIX_FORMAT = '>I'
HEADER_FORMAT = '>HH'
AGENCY_FORMAT = '>H'
DATE_FORMAT = '>HBB'
NUMBERS_FORMAT = '>II'


@dataclass
class Header:
    nombre_length: int
    apellido_length: int


@dataclass
class UserData:
    agencia: int
    nombre: str
    apellido: str
    nacimiento: date
    documento: int
    numero: int


class ProtocolError(ValueError):
    pass

class ProtocolProcessedError(ProtocolError):
    """Excepción que incluye la cantidad de mensajes procesados antes del error."""
    def __init__(self, message, processed_count):
        super().__init__(message)
        self.processed_count = processed_count

def _read_exact(sock, size):
    data = bytearray()
    while len(data) < size:
        chunk = sock.recv(size - len(data))
        if not chunk:
            raise ConnectionError('socket closed while reading payload')
        data.extend(chunk)
    return bytes(data)


def receive_user_data(sock):
    raw_len = _read_exact(sock, struct.calcsize(LENGTH_PREFIX_FORMAT))
    batch_length = struct.unpack(LENGTH_PREFIX_FORMAT, raw_len)[0]
    data = []
    batch_processed = 0

    try:
        for _ in range(batch_length):
            msg_len = _read_exact(sock, struct.calcsize(LENGTH_PREFIX_FORMAT))
            payload_length = struct.unpack(LENGTH_PREFIX_FORMAT, msg_len)[0]
            payload = _read_exact(sock, payload_length)
            user_data = decode_user_data(payload)
            data.append(user_data)
            batch_processed += 1
        return batch_length,data

    except ProtocolError as e:
        raise ProtocolProcessedError(str(e), batch_processed)


def decode_user_data(payload):
    header_size = struct.calcsize(HEADER_FORMAT)
    if len(payload) < header_size:
        raise ProtocolError('payload too short for header')

    nombre_length, apellido_length = struct.unpack(HEADER_FORMAT, payload[:header_size])
    header = Header(nombre_length=nombre_length, apellido_length=apellido_length)

    offset = header_size

    agency_size = struct.calcsize(AGENCY_FORMAT)
    if len(payload) < offset + agency_size:
        raise ProtocolError('payload too short for agencia')

    agencia = struct.unpack(AGENCY_FORMAT, payload[offset:offset + agency_size])[0]
    offset += agency_size

    nombre_end = offset + header.nombre_length
    apellido_end = nombre_end + header.apellido_length

    if len(payload) < apellido_end:
        raise ProtocolError('payload too short for nombre/apellido')

    nombre = payload[offset:nombre_end].decode('utf-8')
    apellido = payload[nombre_end:apellido_end].decode('utf-8')

    offset = apellido_end

    date_size = struct.calcsize(DATE_FORMAT)
    numbers_size = struct.calcsize(NUMBERS_FORMAT)
    expected_size = header_size + agency_size + header.nombre_length + header.apellido_length + date_size + numbers_size
    if len(payload) != expected_size:
        raise ProtocolError('payload size mismatch with protocol definition')

    year, month, day = struct.unpack(DATE_FORMAT, payload[offset:offset + date_size])
    nacimiento = date(year, month, day)
    offset += date_size

    documento, numero = struct.unpack(NUMBERS_FORMAT, payload[offset:offset + numbers_size])

    return UserData(
        agencia=agencia,
        nombre=nombre,
        apellido=apellido,
        nacimiento=nacimiento,
        documento=documento,
        numero=numero,
    )


def encode_response_message(user_data):
    return (
        f'dni={user_data.documento} '
        f'numero={user_data.numero} '
        f'\n'
    ).encode('utf-8')
