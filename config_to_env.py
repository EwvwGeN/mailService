import yaml
from sys import argv

def read_config_from_file(file_path):
    with open(file_path, 'r') as config_file:
        config = yaml.safe_load(config_file)
    return config

def write_to_env_file(config):
    with open('.env', 'w') as env_file:
        for key, value in config.items():
            if isinstance(value, dict):
                write_nested_keys(env_file, key.upper(), value)
            else:
                env_file.write(f'{key.upper()}={value}\n')

def write_nested_keys(env_file, prefix, config):
    for key, value in config.items():
        if isinstance(value, dict):
            write_nested_keys(env_file, f'{prefix}.{key.upper()}', value)
        else:
            env_file.write(f'{prefix}.{key.upper()}={value}\n')

script, path = argv

config = read_config_from_file(path)
write_to_env_file(config)