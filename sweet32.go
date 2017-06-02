package main

import (
    "os"
    "fmt"
    "net"
    "time"
    "crypto/tls"
)


func check(e error) {
    if e != nil {
        fmt.Println(e)
        os.Exit(1)
    }
}


func cipherstring(i uint16) string {
    switch {
    case i == 0x000a:
        return "TLS_RSA_WITH_3DES_EDE_CBC_SHA"
    case i == 0xc012:
        return "TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA"
    default:
        return ""
    }
}


func getConnection(server string, conf *tls.Config, timeout time.Duration) (*tls.Conn) {
    // Create a TCP connection.
    conn, err := net.DialTimeout("tcp", server, timeout)
    check(err)
    fmt.Printf("Successfully connected to: %s\n", conn.RemoteAddr())

    // Create TLS connection using our TCP connection and set a deadline
    // before attempting the handshake. This will ensure the handshake times
    // out.
    tlsconn := tls.Client(conn, conf)
    tlsconn.SetDeadline(time.Now().Add(timeout))

    err = tlsconn.Handshake()
    if err != nil {
        fmt.Println("Unable to complete TLS handshake.")
        os.Exit(0)
    }

    // Reset the deadline to zero.
    tlsconn.SetDeadline(time.Time{})

    // Document cipher suite
    state := tlsconn.ConnectionState()
    fmt.Printf("Using: %s\n", cipherstring(state.CipherSuite))

    return tlsconn
}


func main() {

    if len(os.Args) != 3 {
        fmt.Println("Usage go run sweet32.go server port")
        os.Exit(0)
    }

    host := os.Args[1]
    port := os.Args[2]
    server := fmt.Sprintf("%s:%s", host, port)
    timeout := 30 * time.Second


    // Build TLS Config
    conf := &tls.Config{
        InsecureSkipVerify: true,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
            tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
        },
    }

    // Make our connection
    conn := getConnection(server, conf, timeout)
    defer conn.Close()

    // Write data to the connection.
    for i := 1; i <= 10000; i++ {
        send := []byte(fmt.Sprintf("GET / HTTP/1.1\r\nHost: %s\r\n\r\n", server))
        _, err := conn.Write(send)
        if err != nil {
            fmt.Println("\n")
            fmt.Println(err)
            fmt.Printf("Cannot write to connection after %d requests.\n", i)
            break
        }

        resp := make([]byte, 512)
        _, err = conn.Read(resp)
        if err != nil {
            if err.Error() == "EOF" {
                fmt.Println("\n")
                fmt.Printf("Connection closed after %d requests. Server is not vulnerable.\n", i)
                break
            }
        }

        if i % 20 == 0 {
            fmt.Printf(".")
        }

        if i == 10000 {
            fmt.Println("\n")
            fmt.Println("The server accepted 10000 requests. Server is likely vulnerable.")
        }
    }
    fmt.Println("\n")
}
