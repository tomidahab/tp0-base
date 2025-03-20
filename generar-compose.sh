#!/bin/bash

if [ $# -ne 2 ]; then
    echo "Error en los argumentos. Uso: ./generar_docker_compose.sh <archivo_salida> <n_clients>"
    exit 1
fi

nombre=$1
n_clients=$2

cat <<EOL > $nombre
version: '1.0'
name: tp0

networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24

services:
  server:
    container_name: server
    entrypoint: python /main.py
    environment:
      - PYTHONUNBUFFERED=1
    image: server:latest
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini
    ports:
      - "12345:12345"

EOL

for i in $(seq 1 $n_clients); do
    cat <<EOL >> $nombre
  client$i:
    container_name: client$i
    depends_on:
      - server
    entrypoint: /client
    environment:
      - CLI_ID=$i
    image: client:latest
    networks:
      - testing_net
    volumes:
      - ./client/config.yaml:/config.yaml

EOL
done

echo "Archivo $nombre generado exitosamente."
