# tp0

## Alan Cantero - Padron 99676

## Requerimientos para correr los tests

```txt
iniconfig==2.3.0
packaging==26.0
pluggy==1.6.0
Pygments==2.19.2
pytest==9.0.2
pytest-timeout==2.4.0
PyYAML==6.0.3
```

## Protocolo de mensajes

### tipos de mensajes

Para el trabajo se realizaron 4 tipos de mensajes:

-Bets       Usados para enviar los batchs de apuestas.

-Finish     Para indicar que se terminaron de enviar todos los batchs.

-Query      Para consultar los ganadores del sorteo.

### diagrama de paquetes

Estos diagramas muestran como se envian los distintos tipos de mensajes a traves de la red.

#### 1) Payload de UserData (serializado)

| Campo | Tamano | Tipo | Descripcion |
|---|---|---|---|
| NomL | 2 bytes | uint16 BE | Longitud de `Nombre` |
| ApeL | 2 bytes | uint16 BE | Longitud de `Apellido` |
| Agencia | 2 bytes | uint16 BE | Identificador de agencia |
| Nombre | variable | bytes | Cadena de largo NomL |
| Apellido | variable | bytes | Cadena de largo ApeL |
| Nacimiento | 8 bytes | int64 BE | Timestamp de fecha de nacimiento |
| Documento | 4 bytes | uint32 BE | DNI |
| Numero | 4 bytes | uint32 BE | Numero apostado |

#### 2) Message individual

| Campo | Tamano | Tipo | Descripcion |
|---|---|---|---|
| PayloadSize | 4 bytes | uint32 BE | Cantidad de bytes del payload |
| Payload | variable | bytes | Estructura completa de UserData |

#### 3) BatchMessage (SendBets)

| Campo | Tamano | Tipo | Descripcion |
|---|---|---|---|
| Tipo | 1 byte | byte | `0x01` (Bets) |
| Amount | 4 bytes | uint32 BE | Cantidad de `Message` enviados |
| Message 1..N | variable | bytes | Secuencia de mensajes individuales |

#### 4) FinishMessage

| Campo | Tamano | Tipo | Descripcion |
|---|---|---|---|
| Tipo | 1 byte | byte | `0x02` (Finish) |
| Agencia | 2 bytes | uint16 BE | Agencia que finalizo el envio |

#### 5) QueryMessage

| Campo | Tamano | Tipo | Descripcion |
|---|---|---|---|
| Tipo | 1 byte | byte | `0x03` (Query) |
| Agencia | 2 bytes | uint16 BE | Agencia que consulta ganadores |

### protocolo

En esta subseccion se documentara el protocolo de comunicacion, incluyendo formato de mensajes, codigos de estado y reglas de intercambio.
