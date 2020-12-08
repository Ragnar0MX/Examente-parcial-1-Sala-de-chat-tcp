package main

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
)

//Paquete ...
type Paquete struct {
	Remitente string
	Nombre    string
	Archivo   []byte
}

//Mensaje ...
type Mensaje struct {
	Tipo    int64
	Mensaje string
}

//Usuario ...
type Usuario struct {
	nombre string
	c      net.Conn
}

//Mensajeria ...
type Mensajeria struct {
	Nombre  string
	Tipo    int64
	Archivo string
}

var cliente []net.Conn
var usuarios []Usuario
var mensajeria []Mensajeria

func crearNUsuario(c net.Conn, nombre string) {
	var aux Usuario
	aux.c = c
	aux.nombre = nombre
	usuarios = append(usuarios, aux)
	cliente = append(cliente, c)
}

func eliminarUsuario(c net.Conn) {
	nombre := devolverNombre(c)
	var aux []Usuario
	for _, u := range usuarios {
		if u.c != c {
			aux = append(aux, u)
		}
	}
	m := Mensajeria{Tipo: -1, Nombre: nombre, Archivo: "Desconección"}
	fmt.Println(m.Nombre + ":" + m.Archivo)
	agregarCacheMensajeria(m)
	mandarClinte(c, m)
	usuarios = aux
}

func devolverNombre(c net.Conn) string {
	var nombre string
	for _, u := range usuarios {
		if c == u.c {
			nombre = u.nombre
		}
	}
	return nombre
}
func agregarCacheMensajeria(m Mensajeria) {
	mensajeria = append(mensajeria, m)
}

func server() {
	s, err := net.Listen("tcp", ":9999")
	if err != nil {
		fmt.Println(err)
		return
	}
	for {
		c, err := s.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go handledClient(c)
	}
}

func handledClient(c net.Conn) {
	for {
		var mensaje Mensaje
		err := gob.NewDecoder(c).Decode(&mensaje)
		if err != nil {
			fmt.Println(err)
		} else {
			switch mensaje.Tipo {
			case -1:
				eliminarUsuario(c)
				return
			case 0:
				//Nuevo usuario
				crearNUsuario(c, mensaje.Mensaje)
				m := Mensajeria{Tipo: mensaje.Tipo, Nombre: devolverNombre(c), Archivo: "Conexión"}
				fmt.Println(m.Nombre + "/" + m.Archivo)
				agregarCacheMensajeria(m)
				mandarClinte(c, m)
			case 1:
				m := Mensajeria{Tipo: mensaje.Tipo, Nombre: devolverNombre(c), Archivo: mensaje.Mensaje}
				agregarCacheMensajeria(m)
				fmt.Println(m.Nombre, ": ", m.Archivo)
				mandarClinte(c, m)
			case 2:
				m := Mensajeria{Tipo: mensaje.Tipo, Nombre: devolverNombre(c), Archivo: mensaje.Mensaje}
				agregarCacheMensajeria(m)
				fmt.Println(m.Nombre, ": ", m.Archivo)
				var paquete Paquete
				paquete.Remitente = m.Nombre
				paquete.Nombre = m.Archivo
				archivo2, err := os.Create(paquete.Nombre)
				defer archivo2.Close()
				if err != nil {
					fmt.Println(err)
					continue
				}
				errorEnvio := false
				buff := make([]byte, 1024)
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
					mandarClienteArchivo(paquete, m, c)
				}
			}
		}
	}
}

func mandarClienteArchivo(archivo2 Paquete, m Mensajeria, c net.Conn) {
	for _, n := range usuarios {
		if n.c != c {
			err := gob.NewEncoder(n.c).Encode(m)
			if err != nil {
				fmt.Println(err)
			} else {
				_, err := n.c.Write(archivo2.Archivo)
				if err != nil {
					fmt.Println(err)
				} else {
					fmt.Println("Archivo enviado!")
					n.c.Write(archivo2.Archivo)
				}
			}
		}
	}
}

func mandarClinte(c net.Conn, m Mensajeria) {
	for _, n := range usuarios {
		if n.c != c {
			err := gob.NewEncoder(n.c).Encode(m)
			if err != nil {
				fmt.Println(err)
			}
			return
		}
	}
}

func imprimirMensajes() {
	for _, m := range mensajeria {
		fmt.Println(m.Nombre + ":" + m.Archivo)
	}
}
func crearBackup() {
	file, err := os.Create("Respaldo.txt")
	if err != nil {
		fmt.Println("No se pudo crear")
	}
	defer file.Close()
	for _, m := range mensajeria {
		file.WriteString("Remitente:" + m.Nombre + "\tMensaje o Archivo:" + m.Archivo)
	}
}
func main() {
	go server()
	menu := "Menu\n0.-Salir\n1.-Mostrar mensajes\n2.-Crear respaldo mensaje"
	var input int64
	fmt.Println("ServidorArrancado")
	for {
		fmt.Println(menu)
		fmt.Println("Opcion:")
		fmt.Scanln(&input)
		switch input {
		case 1:
			imprimirMensajes()
		case 2:
			crearBackup()
		case 0:
			return
		default:
			fmt.Println("Ingrese una opcón valida")
		}
	}
}
