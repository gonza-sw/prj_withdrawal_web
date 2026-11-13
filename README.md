# Banca Digital — Aplicación Web en Go (`prj_withdrawal`)

Aplicación web local desarrollada en Go para gestionar una cuenta bancaria mediante operaciones de abono, retiro y consulta de movimientos. El proyecto adapta la lógica de negocio y validaciones de la versión de consola hacia una interfaz web usando únicamente la biblioteca estándar de Go (`net/http`, `embed`, `encoding/json` y `sync`).

## Requisitos

- **Go 1.22 o superior** (desarrollado con Go 1.27).
- Navegador web moderno.
- No requiere base de datos, Node.js ni paquetes externos.

## Cómo iniciar la aplicación

1. Abre una terminal dentro de la carpeta raíz `prj_withdrawal` (al nivel de `main.go`).
2. Ejecuta el servidor web:
   ```bash
   go run main.go
   ```
3. Ingresa desde el navegador a:
   ```text
   http://localhost:8080
   ```

- **Cambiar el puerto:** Si el `8080` está ocupado, indica otro con `-port`:
  ```bash
  go run main.go -port=8081
  ```
- **Ejecutar validaciones en consola (CLI):** Para correr los escenarios de prueba en terminal sin levantar la web:
  ```bash
  go run main.go -cli
  ```
- **Detener el servidor:** Presiona `Ctrl + C` en la terminal. El estado de la cuenta se mantiene en memoria y vuelve a los valores por defecto al reiniciar.

## Uso de la interfaz

1. **Panel superior y saldo:** Muestra el titular activo (por defecto `Alice` con `234.56 €`).
2. **Operaciones de cuenta:** Ingresa un monto y selecciona **Abonar** o **Retirar**. El sistema valida el saldo disponible y muestra una notificación con el resultado.
3. **Abrir / Reiniciar cuenta:** Permite configurar un nuevo titular y saldo inicial en cualquier momento limpiando el historial en memoria.
4. **Historial de movimientos:** Muestra cada intento de operación con su hora, detalle, monto, saldo resultante y estado (`Exitoso`, `Rechazado` o `Error`).

## Pruebas rápidas recomendadas

Iniciando con la cuenta por defecto (**Alice**, saldo inicial **234.56 €**):

| Operación |  Monto   | Estado esperado | Saldo posterior |
| :-------- | :------: | :-------------: | :-------------: |
| Retirar   | `200.00` |     Exitoso     |    `34.56 €`    |
| Retirar   | `100.00` |    Rechazado    |    `34.56 €`    |
| Abonar    | `150.00` |     Exitoso     |   `184.56 €`    |
| Abonar    | `-10.00` |      Error      |   `184.56 €`    |

## Registro y exportación JSON

Cada operación se registra internamente con los campos `id`, `tipo`, `monto`, `estado`, `saldo_restante`, `detalle` y `timestamp`. Desde la cabecera de la tabla puedes pulsar **Exportar JSON** para descargar el archivo `historial_transacciones.json`, o consultar el estado completo en formato JSON directamente en `http://localhost:8080/api/estado`.

## Estructura del proyecto

- `main.go`: Servidor HTTP, endpoints REST (`/api/estado`, `/api/operacion`, `/api/reiniciar`), HTML embebido y modo `-cli`.
- `index.html`: Interfaz web (HTML, CSS y JavaScript).
- `headers/class.go`: Entidad `Cuenta` con control de concurrencia (`sync.Mutex`), operaciones `Retirar` y `Abonar`, y serialización JSON.
- `headers/test_bdd.go`: Pruebas automatizadas de escenarios en consola.
- `go.mod`: Definición del módulo Go.
