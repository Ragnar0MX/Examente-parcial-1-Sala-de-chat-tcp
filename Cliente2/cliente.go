package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

//Mensaje ...
type Mensaje struct {
	Tipo    int64
	Mensaje string
}

//Mensajeria ...
type Mensajeria struct {
	Nombre  string
	Tipo    int64
	Archivo string
}

//Paquete ...
type Paquete struct {
	Remitente string
	Nombre    string
	Archivo   []byte
}

func validarMensaje(c net.Conn, m Mensaje) bool {
	err := gob.NewEncoder(c).Encode(m)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return true
}

func recibirMensaje(c net.Conn) {
	for {
		var m Mensajeria
		err := gob.NewDecoder(c).Decode(&m)
		if err != nil {
			fmt.Println(err)
			return
		}
		switch m.Tipo {
		case 1:
			fmt.Println(m.Nombre + ".- " + m.Archivo)
		case 2:
			fmt.Println(m.Nombre + "/" + m.Archivo)
			recibirArchivo(m, c)
		}
	}
}

func recibirArchivo(m Mensajeria, c net.Conn) {
	var paquete Paquete
	paquete.Remitente = m.Nombre
	paquete.Nombre = m.Archivo
	archivo2, err := os.Create(paquete.Nombre)
	buff := make([]byte, 1024)
	defer archivo2.Close()
	if err != nil {
		fmt.Println(err)
		return
	} else {
		errorEnvio := false
		for {
			NDatos, err := c.Read(buff)
			if err != nil {
				fmt.Println(err)
				errorEnvio = true
				break
			} else {
				paquete.Archivo = append(paquete.Archivo, buff[:NDatos]...)
				archivo2.Write(buff[:NDatos])
				if NDatos != 1024 {
					break
				}
			}
		}
		if !errorEnvio {
			fmt.Println("Recibido")
			archivo2.Close()
		} else {
			fmt.Println("No recibido")
		}
	}
}

func mandarMensajeTexto(c net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	var aux2 Mensaje
	fmt.Println("Escribe mensaje: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println(err)
	} else {
		aux2.Mensaje = text
		aux2.Tipo = 1
		if !validarMensaje(c, aux2) {
			fmt.Println("Error al mandar el mensaje")
		}
	}
}

func desconectar(c net.Conn) {
	m := Mensaje{Tipo: -1, Mensaje: "Desconectao"}
	err := gob.NewEncoder(c).Encode(m)
	if err != nil {
		fmt.Println(err)
		return
	}
}

func enviarArchivo(c net.Conn) {
	reader := bufio.NewReader(os.Stdin)
	var aux2 Mensaje
	fmt.Print("Ingresa el nombre del archivo: ")
	text, _ := reader.ReadString('\n')
	dir := strings.TrimSpace(strings.TrimRight(text, "\n"))
	archivo, err := ioutil.ReadFile(dir)
	if err != nil {
		fmt.Println(err)
		return
	}
	aux2.Tipo = 2
	aux2.Mensaje = dir
	if !validarMensaje(c, aux2) {
		fmt.Println("Error al mandar el mensaje")
	} else {
		_, err := c.Write(archivo)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println("Archivo enviado!")
			c.Write(archivo)
		}
	}
}

func main() {
	var nombre string
	var mensaje Mensaje
	var opc int64
	menu := "\n\n***   Menu Cliente   ***\n************************\n\n0.- Salir\n1.- Mandar Mensaje\n2.- Mandar Archivo\nopcion: "
	c, err := net.Dial("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("Ingrese su nombre:")
	fmt.Scanln(&nombre)
	mensaje.Mensaje = nombre
	mensaje.Tipo = 0
	err = gob.NewEncoder(c).Encode(mensaje)
	if err != nil {
		fmt.Println(err)
		return
	}
	go recibirMensaje(c)
	fmt.Println(menu)
	for {
		fmt.Println("Ingrese opción:")
		fmt.Scanln(&opc)
		switch opc {
		case 0:
			desconectar(c)
			c.Close()
			return
		case 1:
			mandarMensajeTexto(c)
		case 2:
			enviarArchivo(c)
		default:
			fmt.Println("respuesta no valida")
		}
	}
}
