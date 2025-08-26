# NBOX Tools - Hasher de Contraseñas

Esta es una herramienta de línea de comandos (CLI) simple y segura para generar hashes de contraseñas utilizando el algoritmo **bcrypt**.

Está diseñada para ayudar a los desarrolladores a crear de forma segura las credenciales que se utilizarán en archivos de configuración o gestores de secretos, como AWS Secrets Manager.

## Características

-   **Entrada Segura:** Lee la contraseña de forma interactiva sin mostrarla en la pantalla para proteger contra "shoulder surfing".
-   **Costo Configurable:** Permite ajustar el costo computacional de bcrypt para balancear seguridad y rendimiento.
-   **Robusta:** Maneja errores comunes y valida los parámetros de entrada.
-   **Multiplataforma:** Compila y funciona en Linux, macOS y Windows.

## Instalación

Para compilar la herramienta, navega a la raíz de tu proyecto y ejecuta el siguiente comando. Esto creará un ejecutable llamado `hasher`.

```shell
go build -o hasher ./cmd/hasher/main.go
```

## Uso

Una vez compilado, puedes ejecutar el programa directamente desde tu terminal.

#### Generar un hash con el costo por defecto (10)

```shell
./hasher
```
El programa te pedirá que ingreses la contraseña.

#### Generar un hash con un costo específico

Puedes usar el flag `-cost` para especificar un costo diferente. Un costo mayor es más seguro pero más lento.

```shell
./hasher -cost=12
```

#### Obtener Ayuda

Para ver las opciones disponibles y la descripción del comando, usa el flag `-h`.

```shell
./hasher -h
```

La salida será:
```
▗▖  ▗▖▗▄▄▖  ▗▄▖ ▗▖  ▗▖    ▗▄▄▄▖▗▄▖  ▗▄▖ ▗▖    ▗▄▄▖
▐▛▚▖▐▌▐▌ ▐▌▐▌ ▐▌ ▝▚▞☗       █ ▐▌ ▐▌▐▌ ▐▌▐▌   ▐▌
▐▌ ▝▜▌▐▛▀▚▖▐▌ ▐▌  ▐▌        █ ▐▌ ▐▌▐▌ ▐▌▐▌    ▝▀▚▖
▐▌  ▐▌▐▙▄▞☗▝▚▄▞☗▗▞☗▝▚▖      █ ▝▚▄▞☗▝▚▄▞☗▐▙▄▄▖▗▄▄▞☗

Herramienta para generar hashes de contraseñas usando bcrypt.

Uso: hasher [opciones]

Opciones:
  -cost int
        Costo de Bcrypt (entre 4 y 31). (default 10)
```

## Seguridad

-   La herramienta utiliza `golang.org/x/term` para leer la contraseña sin eco en la terminal.
-   Utiliza `golang.org/x/crypto/bcrypt`, la implementación estándar y segura para hashing de contraseñas en Go.