import sys
import yaml

def generar_docker_compose(nombre_archivo, cantidad_clientes):
    compose_dict = {
        'name': 'tp0',
        'services': {
            'server': {
                'container_name': 'server',
                'image': 'server:latest',
                'entrypoint': 'python /main.py', #Uso python en vez de python3
                'environment': [
                    'PYTHONUNBUFFERED=1',
                    'LOGGING_LEVEL=DEBUG'
                ],
                'networks': ['testing_net']
            }
        },
        'networks': {
            'testing_net': {
                'ipam': {
                    'driver': 'default',
                    'config': [{
                        'subnet': '172.25.125.0/24'
                    }]
                }
            }
        }
    }

    for i in range(1, cantidad_clientes + 1):
        cliente_nombre = f'client{i}'
        compose_dict['services'][cliente_nombre] = {
            'container_name': cliente_nombre,
            'image': 'client:latest',
            'entrypoint': '/client',
            'environment': [
                f'CLI_ID={i}',
                'CLI_LOG_LEVEL=DEBUG'
            ],
            'networks': ['testing_net'],
            'depends_on': ['server']
        }

    with open(nombre_archivo, 'w') as file:
        yaml.dump(compose_dict, file, default_flow_style=False)

if len(sys.argv) != 3:
    print("Error en args, ej: python3 mi-generador.py <archivo_salida> <cantidad_clientes>")
    sys.exit(1)

nombre_archivo = sys.argv[1]
cantidad_clientes = int(sys.argv[2])

generar_docker_compose(nombre_archivo, cantidad_clientes)
