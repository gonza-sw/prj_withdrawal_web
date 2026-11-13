package headers

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Transaccion representa la estructura del log JSON exigida en US-04.
type Transaccion struct {
	ID            int       `json:"id"`
	Tipo          string    `json:"tipo"` // "RETIRO" o "ABONO"
	Monto         float64   `json:"monto"`
	Estado        string    `json:"estado"` // "EXITO", "FALLO", "ERROR"
	SaldoRestante float64   `json:"saldo_restante"`
	Detalle       string    `json:"detalle"`
	Timestamp     time.Time `json:"timestamp"`
}

// Cuenta representa la entidad de negocio descrita en el Feature Gherkin.
type Cuenta struct {
	mu            sync.Mutex
	saldo         float64
	nombre        string
	transacciones []Transaccion
	contadorID    int
}

// NewCuenta es el constructor equivalente a Cuenta(nombre, saldoInicial).
func NewCuenta(nombre string, saldoInicial float64) *Cuenta {
	return &Cuenta{
		nombre:        nombre,
		saldo:         saldoInicial,
		transacciones: make([]Transaccion, 0),
		contadorID:    0,
	}
}

// ObtenerSaldo devuelve el saldo actual.
func (c *Cuenta) ObtenerSaldo() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.saldo
}

// Nombre devuelve el nombre del titular.
func (c *Cuenta) Nombre() string {
	return c.nombre
}

// String implementa fmt.Stringer para depuración.
func (c *Cuenta) String() string {
	return fmt.Sprintf("Cuenta{%s, saldo=%.2f}", c.nombre, c.ObtenerSaldo())
}

// registrarTransaccion crea y almacena un registro interno (debe llamarse con el mutex bloqueado).
func (c *Cuenta) registrarTransaccion(tipo string, monto float64, estado string, detalle string) Transaccion {
	c.contadorID++
	tx := Transaccion{
		ID:            c.contadorID,
		Tipo:          tipo,
		Monto:         monto,
		Estado:        estado,
		SaldoRestante: c.saldo,
		Detalle:       detalle,
		Timestamp:     time.Now(),
	}
	c.transacciones = append(c.transacciones, tx)
	return tx
}

// Retirar (US-01 + US-04) intenta retirar una cantidad y registra la transacción.
func (c *Cuenta) Retirar(cantidad float64) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cantidad < 0 {
		err := errors.New("no se pueden retirar cantidades negativas")
		c.registrarTransaccion("RETIRO", cantidad, "ERROR", err.Error())
		return false, err
	}
	if cantidad > c.saldo {
		c.registrarTransaccion("RETIRO", cantidad, "FALLO", "Fondos insuficientes para realizar el retiro")
		return false, nil
	}
	c.saldo -= cantidad
	c.registrarTransaccion("RETIRO", cantidad, "EXITO", "Retiro procesado correctamente")
	return true, nil
}

// Abonar (US-01b + US-04) intenta abonar una cantidad y registra la transacción.
func (c *Cuenta) Abonar(cantidad float64) (bool, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cantidad <= 0 {
		err := errors.New("no se pueden abonar cantidades negativas o cero")
		c.registrarTransaccion("ABONO", cantidad, "ERROR", err.Error())
		return false, err
	}
	c.saldo += cantidad
	c.registrarTransaccion("ABONO", cantidad, "EXITO", "Abono procesado correctamente")
	return true, nil
}

// ObtenerTransacciones devuelve una copia del slice de transacciones.
func (c *Cuenta) ObtenerTransacciones() []Transaccion {
	c.mu.Lock()
	defer c.mu.Unlock()
	copia := make([]Transaccion, len(c.transacciones))
	copy(copia, c.transacciones)
	return copia
}

// ObtenerHistorialJSON devuelve todas las transacciones en formato JSON indentado.
func (c *Cuenta) ObtenerHistorialJSON() (string, error) {
	txs := c.ObtenerTransacciones()
	bytes, err := json.MarshalIndent(txs, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ObtenerUltimaTransaccionJSON devuelve la última transacción en formato JSON.
func (c *Cuenta) ObtenerUltimaTransaccionJSON() (string, error) {
	txs := c.ObtenerTransacciones()
	if len(txs) == 0 {
		return "[]", nil
	}
	tx := txs[len(txs)-1]
	bytes, err := json.MarshalIndent(tx, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}