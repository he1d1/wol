package main

import (
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gofiber/fiber/v3"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/sabhiram/go-wol/wol"
)

func main() {
	// Initialize a new Fiber app
	app := fiber.New()

	app.Get("/", func(c fiber.Ctx) error {
		return c.JSON(struct {
			Ping string `json:"/ping"`
			Wake string `json:"/wake"`
		}{
			Ping: "Ping desktop to see if its online",
			Wake: "Send a magic packet to desktop",
		})
	})

	app.Get("/ping", func(c fiber.Ctx) error {
		pinger, err := probing.NewPinger("desktop")
		if err != nil {
			panic(err)
		}
		pinger.Count = 3
		err = pinger.Run() // Blocks until finished.
		if err != nil {
			panic(err)
		}
		stats := pinger.Statistics()
		return c.JSON(stats)
	})

	app.Get("/wake", func(c fiber.Ctx) error {
		magicPacket, err := wol.New(os.Getenv("WOL-MAC"))
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		bytes, err := magicPacket.Marshal()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}

		conn, err := net.Dial("udp", "255.255.255.255:9")
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		defer conn.Close()

		n, err := conn.Write(bytes)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, err.Error())
		}
		if n != 102 {
			return fiber.NewError(fiber.StatusInternalServerError, fmt.Sprintf("Sent %d bytes instead of 102.", n))
		}

		return c.Redirect().Status(fiber.StatusSeeOther).To("/ping")
	})

	// Start the server on port 3000
	log.Fatal(app.Listen(":3000"))
}
