package headers

import (
	"fmt"
	"math"
)

// imprimirSaldo imprime el saldo formateado a 2 decimales.
func imprimirSaldo(cuenta *Cuenta) {
	fmt.Printf("Saldo de %.2f €\n", cuenta.ObtenerSaldo())
}

// imprimirLogJSON muestra el registro JSON de la transacción actual.
func imprimirLogJSON(cuenta *Cuenta) {
	jsonLog, err := cuenta.ObtenerUltimaTransaccionJSON()
	if err == nil {
		fmt.Println("Log de Transacción (JSON US-04):")
		fmt.Println(jsonLog)
	}
}

// ------------------------------------------------------------------
// Validación de Escenarios Gherkin Dinámicos - RETIRO (US-01 / US-04)
// ------------------------------------------------------------------

func TestEscenario1Exito(nombre string, saldoInicial float64, cantidadRetirar float64) {
	fmt.Printf("--- Escenario 1: Successful withdrawal within balance (%s) ---\n", nombre)

	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)
	fmt.Printf("When: %s intenta retirar %.2f\n", nombre, cantidadRetirar)

	exito, err := cuenta.Retirar(cantidadRetirar)

	if err != nil {
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n\n", err)
		return
	}

	if exito {
		fmt.Println("Resultado: ÉXITO. La retirada fue procesada.")
		imprimirSaldo(cuenta)
		saldoEsperado := saldoInicial - cantidadRetirar
		if math.Abs(cuenta.ObtenerSaldo()-saldoEsperado) < 0.001 {
			fmt.Printf("Validación: El saldo restante es correcto (%.2f).\n", saldoEsperado)
		} else {
			fmt.Println("ERROR: El saldo restante no coincide.")
		}
	} else {
		fmt.Println("Resultado: FALLO. La retirada fue rechazada (incorrecto).")
	}
	imprimirLogJSON(cuenta)
	fmt.Println()
}

func TestEscenario2Fallo(nombre string, saldoInicial float64, cantidadRetirar float64) {
	fmt.Printf("--- Escenario 2: Declined withdrawal in excess of balance (%s) ---\n", nombre)

	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)
	fmt.Printf("When: %s intenta retirar %.2f\n", nombre, cantidadRetirar)

	exito, err := cuenta.Retirar(cantidadRetirar)

	if err != nil {
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n\n", err)
		return
	}

	if !exito {
		fmt.Println("Resultado: ÉXITO. La retirada fue rechazada correctamente.")
		imprimirSaldo(cuenta)
		if math.Abs(cuenta.ObtenerSaldo()-saldoInicial) < 0.001 {
			fmt.Printf("Validación: El saldo no ha cambiado (%.2f).\n", saldoInicial)
		} else {
			fmt.Println("ERROR: El saldo se modificó incorrectamente.")
		}
	} else {
		fmt.Println("Resultado: FALLO. La retirada fue procesada (debería haber sido rechazada).")
	}
	imprimirLogJSON(cuenta)
	fmt.Println()
}

func TestEscenario3Error(nombre string, saldoInicial float64, cantidadRetirar float64) {
	fmt.Printf("--- Escenario 3: Validation Error (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)

	_, err := cuenta.Retirar(cantidadRetirar)
	if err != nil {
		fmt.Printf("Resultado: ÉXITO. El sistema atrapó el error correctamente: %v\n", err)
		imprimirSaldo(cuenta)
	} else {
		fmt.Println("Resultado: FALLO. Se procesó un retiro que debía dar error.")
	}
	imprimirLogJSON(cuenta)
	fmt.Println()
}

// ------------------------------------------------------------------
// Validación de Escenarios Gherkin Dinámicos - ABONO (US-01b / US-04)
// ------------------------------------------------------------------

func TestEscenarioAbonoExito(nombre string, saldoInicial float64, cantidadAbonar float64) {
	fmt.Printf("--- Escenario Abono 1: Successful deposit (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)
	fmt.Printf("Given: %s tiene %.2f en su cuenta\n", nombre, saldoInicial)
	fmt.Printf("When: %s abona %.2f en su cuenta\n", nombre, cantidadAbonar)

	exito, err := cuenta.Abonar(cantidadAbonar)
	if err != nil {
		fmt.Printf("Resultado: FALLO. Error inesperado: %v\n\n", err)
		return
	}

	if exito {
		fmt.Println("Resultado: ÉXITO. El abono fue procesado.")
		imprimirSaldo(cuenta)
		saldoEsperado := saldoInicial + cantidadAbonar
		if math.Abs(cuenta.ObtenerSaldo()-saldoEsperado) < 0.001 {
			fmt.Printf("Validación: El nuevo saldo es correcto (%.2f).\n", saldoEsperado)
		} else {
			fmt.Println("ERROR: El saldo resultante no coincide.")
		}
	}
	imprimirLogJSON(cuenta)
	fmt.Println()
}

func TestEscenarioAbonoError(nombre string, saldoInicial float64, cantidadAbonar float64) {
	fmt.Printf("--- Escenario Abono 2: Invalid deposit amount (%s) ---\n", nombre)
	cuenta := NewCuenta(nombre, saldoInicial)

	_, err := cuenta.Abonar(cantidadAbonar)
	if err != nil {
		fmt.Printf("Resultado: ÉXITO. El sistema atrapó el error de abono correctamente: %v\n", err)
		imprimirSaldo(cuenta)
	} else {
		fmt.Println("Resultado: FALLO. Se procesó un abono inválido.")
	}
	imprimirLogJSON(cuenta)
	fmt.Println()
}

// TestIntegridadJSONCompleto: Muestra el historial completo acumulado de transacciones en JSON.
func TestIntegridadJSONCompleto() {
	fmt.Println("=== DEMOSTRACIÓN US-04: INTEGRIDAD DE TRANSACCIONES Y HISTORIAL JSON ===")
	cuenta := NewCuenta("Juan Perez", 500.00)

	fmt.Println("Ejecutando operaciones secuenciales sobre la cuenta de Juan Perez (Saldo inicial: 500.00 €)...")
	cuenta.Abonar(200.00)   // EXITO (Abono)
	cuenta.Retirar(150.00)  // EXITO (Retiro)
	cuenta.Retirar(1000.00) // FALLO (Fondos insuficientes)
	cuenta.Retirar(-50.00)  // ERROR (Monto negativo)
	cuenta.Abonar(0.00)     // ERROR (Abono cero)

	historial, err := cuenta.ObtenerHistorialJSON()
	if err == nil {
		fmt.Println("\nHistorial Completo de Transacciones (JSON Log):")
		fmt.Println(historial)
	}
}