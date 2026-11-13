package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"prj_withdrawal/headers"
)

//go:embed index.html
var htmlPage string

var cuentaActiva *headers.Cuenta

type EstadoResponse struct {
	Nombre        string                `json:"nombre"`
	Saldo         float64               `json:"saldo"`
	Mensaje       string                `json:"mensaje,omitempty"`
	TipoMensaje   string                `json:"tipo_mensaje,omitempty"` // "EXITO", "FALLO", "ERROR"
	Transacciones []headers.Transaccion `json:"transacciones"`
}

type OperacionRequest struct {
	Tipo  string  `json:"tipo"`  // "RETIRO" o "ABONO"
	Monto float64 `json:"monto"`
}

type ReiniciarRequest struct {
	Nombre       string  `json:"nombre"`
	SaldoInicial float64 `json:"saldo_inicial"`
}

func responderJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlPage))
}

func handleEstado(w http.ResponseWriter, r *http.Request) {
	resp := EstadoResponse{
		Nombre:        cuentaActiva.Nombre(),
		Saldo:         cuentaActiva.ObtenerSaldo(),
		Transacciones: cuentaActiva.ObtenerTransacciones(),
	}
	responderJSON(w, http.StatusOK, resp)
}

func handleOperacion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req OperacionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "JSON inválido", http.StatusBadRequest)
		return
	}

	var msg, tipoMsg string

	switch req.Tipo {
	case "RETIRO":
		exito, err := cuentaActiva.Retirar(req.Monto)
		if err != nil {
			msg = "Error de validación: " + err.Error()
			tipoMsg = "ERROR"
		} else if !exito {
			msg = "Operación rechazada: Fondos insuficientes"
			tipoMsg = "FALLO"
		} else {
			msg = fmt.Sprintf("Retiro de %.2f € procesado con éxito", req.Monto)
			tipoMsg = "EXITO"
		}
	case "ABONO":
		exito, err := cuentaActiva.Abonar(req.Monto)
		if err != nil {
			msg = "Error de validación: " + err.Error()
			tipoMsg = "ERROR"
		} else if exito {
			msg = fmt.Sprintf("Abono de %.2f € procesado con éxito", req.Monto)
			tipoMsg = "EXITO"
		}
	default:
		http.Error(w, "Tipo de operación no soportado", http.StatusBadRequest)
		return
	}

	resp := EstadoResponse{
		Nombre:        cuentaActiva.Nombre(),
		Saldo:         cuentaActiva.ObtenerSaldo(),
		Mensaje:       msg,
		TipoMensaje:   tipoMsg,
		Transacciones: cuentaActiva.ObtenerTransacciones(),
	}
	responderJSON(w, http.StatusOK, resp)
}

func handleReiniciar(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	var req ReiniciarRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Nombre == "" {
		http.Error(w, "Datos inválidos", http.StatusBadRequest)
		return
	}

	cuentaActiva = headers.NewCuenta(req.Nombre, req.SaldoInicial)
	resp := EstadoResponse{
		Nombre:        cuentaActiva.Nombre(),
		Saldo:         cuentaActiva.ObtenerSaldo(),
		Mensaje:       fmt.Sprintf("Cuenta reiniciada para %s con saldo %.2f €", req.Nombre, req.SaldoInicial),
		TipoMensaje:   "EXITO",
		Transacciones: cuentaActiva.ObtenerTransacciones(),
	}
	responderJSON(w, http.StatusOK, resp)
}

func main() {
	modoCLI := flag.Bool("cli", false, "Ejecutar pruebas BDD/TDD en modo consola")
	puerto := flag.String("port", "8080", "Puerto del servidor web")
	flag.Parse()

	if *modoCLI {
		fmt.Println("=== EJECUTANDO PRUEBAS BDD + US-04 LOGS JSON ===")
		headers.TestEscenario1Exito("Alice", 234.56, 200.00)
		headers.TestEscenario2Fallo("Hamza", 198.76, 200.00)
		headers.TestEscenario3Error("Sofia", 150.00, -50.00)
		headers.TestEscenarioAbonoExito("Carlos", 100.00, 50.00)
		headers.TestEscenarioAbonoError("Elena", 100.00, -20.00)
		headers.TestIntegridadJSONCompleto()
		os.Exit(0)
	}

	// Inicializamos la cuenta por defecto (Escenario base Gherkin: Alice con 234.56 €)
	cuentaActiva = headers.NewCuenta("Alice", 234.56)

	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/api/estado", handleEstado)
	http.HandleFunc("/api/operacion", handleOperacion)
	http.HandleFunc("/api/reiniciar", handleReiniciar)

	fmt.Printf("Servidor Web prj_withdrawal iniciado en: http://localhost:%s\n", *puerto)
	fmt.Println("Presiona Ctrl+C para detener el servidor.")
	log.Fatal(http.ListenAndServe(":"+*puerto, nil))
}