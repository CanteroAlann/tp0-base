#!/bin/bash

if [ "$#" -ne 2 ]; then
    echo "Uso: $0 <nombre_archivo_salida> <cantidad_clientes>"
    exit 1
fi

OUTPUT_FILE=$1
CLIENT_COUNT=$2

echo "Generando archivo: $OUTPUT_FILE"
echo "Cantidad de clientes: $CLIENT_COUNT"

python3 mi-generador.py "$OUTPUT_FILE" "$CLIENT_COUNT"

