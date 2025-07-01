package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"strconv"

	"github.com/asaskevich/EventBus"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/wisepythagoras/pos-system/core"
	"github.com/wisepythagoras/pos-system/crypto"
)

type AssetManifest struct {
	Files       map[string]string `json:"files"`
	EntryPoints []string          `json:"entrypoints"`
}

// parseConfig parses the configuration either from the same folder, or
// from an explicit path.
func parseConfig(customConfig *string) (*core.Config, error) {
	if customConfig == nil || len(*customConfig) == 0 {
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
	} else {
		viper.SetConfigFile(*customConfig)
	}

	var config core.Config

	// Try to read the configuration file.
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	// Default config.
	viper.SetDefault("server.port", 8088)

	// Parse the configuration into the config object.
	err := viper.Unmarshal(&config)

	if err != nil {
		return nil, err
	}

	return &config, nil
}

func authHandler(isAdmin bool, configAuthToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		userCookie := session.Get("user")
		xAuthToken := c.GetHeader("x-auth-token")

		if (userCookie == nil || userCookie == "") && xAuthToken != configAuthToken {
			c.AbortWithStatus(http.StatusForbidden)
		} else {
			if isAdmin && xAuthToken != configAuthToken {
				user := &core.UserStruct{}
				json.Unmarshal([]byte(userCookie.(string)), &user)

				// Prevent anyone who is not logged in to view this page.
				if user == nil || !user.IsAdmin {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}

			c.Next()
		}
	}
}

func main() {
	customConfig := flag.String("config", "", "The path to a custom config file")
	flag.Parse()

	db, err := core.ConnectDB()

	if err != nil {
		panic(err)
	}

	config, err := parseConfig(customConfig)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	bus := EventBus.New()

	// Instanciate all of the route handlers here.
	productHandlers := &core.ProductHandlers{
		DB:  db,
		Bus: bus,
	}
	orderHandlers := &core.OrderHandlers{
		DB:     db,
		Bus:    bus,
		Config: config,
	}
	userHandlers := &core.UserHandlers{
		DB:     db,
		Config: config,
	}
	stationHandlers := &core.StationHandlers{
		DB:     db,
		Config: config,
	}

	manifsetFileName := "build/asset-manifest.json"
	var info fs.FileInfo

	// TODO: Maybe add a command line param for this.
	if info, err = os.Stat(manifsetFileName); os.IsNotExist(err) {
		fmt.Println("You need to build the UI code.")
		os.Exit(1)
	}

	assetManifest := new(AssetManifest)
	assetManifestBt, err := os.ReadFile(manifsetFileName)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	} else {
		err = json.Unmarshal(assetManifestBt, assetManifest)

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	// Start listeningfor messages and send them to the clients, if there are any.
	productHandlers.StartWSHandler()

	adminAuthToken := config.Admin.Token

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	// Apply the sessions middleware.
	store := cookie.NewStore([]byte(config.Secret))
	router.Use(sessions.Sessions("mysession", store))

	router.Any("/", func(c *gin.Context) {
		session := sessions.Default(c)
		userCookie := session.Get("user")

		if userCookie == nil || userCookie == "" {
			c.Redirect(http.StatusMovedPermanently, "/login")
			return
		}

		// c.HTML(http.StatusOK, "index.html", gin.H{
		// 	"title": "POS",
		// 	"admin": false,
		// })

		showNewLanding := c.Query("_nlp") == "1"
		params := gin.H{}
		params["showNewLanding"] = showNewLanding
		params["title"] = "POS"
		params["entrypoint"] = assetManifest.EntryPoints[0]
		params["admin"] = false

		formatted := strconv.FormatInt(info.ModTime().UnixMilli(), 16)
		hash, _ := crypto.GetSHA3512Hash([]byte(formatted))
		params["ts"] = crypto.ByteArrayToHex(hash)

		c.HTML(http.StatusOK, "index.html", params)
	})

	router.GET("/admin", authHandler(true, adminAuthToken), func(c *gin.Context) {
		session := sessions.Default(c)
		userCookie := session.Get("user")

		if userCookie == nil || userCookie == "" {
			c.Redirect(http.StatusMovedPermanently, "/login")
			return
		}

		showNewLanding := c.Query("_nlp") == "1"
		params := gin.H{}
		params["showNewLanding"] = showNewLanding
		params["title"] = "POS Admin"
		params["entrypoint"] = assetManifest.EntryPoints[0]
		params["admin"] = true

		formatted := strconv.FormatInt(info.ModTime().UnixMilli(), 16)
		hash, _ := crypto.GetSHA3512Hash([]byte(formatted))
		params["ts"] = crypto.ByteArrayToHex(hash)

		c.HTML(http.StatusOK, "index.html", params)
	})

	// Set the static/public path.
	router.Use(static.Serve("/", static.LocalFile("./public", false)))
	router.Use(static.Serve("/static", static.LocalFile("./build/static", false)))

	router.POST("/login", userHandlers.Login)
	router.GET("/login", userHandlers.LoginPage)
	router.GET("/logout", userHandlers.Logout)

	router.GET("/api/user", authHandler(false, adminAuthToken), userHandlers.GetLoggedInUser)
	router.POST("/api/user", authHandler(true, adminAuthToken), userHandlers.Create)
	router.DELETE("/api/user/:userId", authHandler(true, adminAuthToken), userHandlers.Delete)
	router.GET("/api/users", authHandler(true, adminAuthToken), userHandlers.List)

	router.GET("/api/orders/earnings", authHandler(true, adminAuthToken), orderHandlers.GetTotalEarnings)
	router.GET("/api/orders/earnings/:day", authHandler(true, adminAuthToken), orderHandlers.EarningsPerDay)
	router.GET("/api/orders/totals/export", authHandler(true, adminAuthToken), orderHandlers.ExportTotalEarnings)
	router.GET("/api/orders", authHandler(false, adminAuthToken), orderHandlers.GetOrders)
	router.GET("/api/orders/past_year", authHandler(true, adminAuthToken), orderHandlers.OrdersPastYear)
	router.POST("/api/order", authHandler(false, adminAuthToken), orderHandlers.CreateOrder)
	router.GET("/api/order/:orderId", authHandler(false, adminAuthToken), orderHandlers.FetchOrder)
	router.GET("/api/order/:orderId/receipt/:printerId", authHandler(false, adminAuthToken), orderHandlers.PrintReceipt)
	router.GET("/api/order/:orderId/pub", orderHandlers.PublicOrder)
	router.GET("/api/order/:orderId/qrcode", authHandler(false, adminAuthToken), orderHandlers.OrderQRCode)
	router.PUT("/api/order/:orderId/product/:productId/:state", authHandler(false, adminAuthToken), orderHandlers.UpdateFulfilled)
	router.DELETE("/api/order/:orderId", authHandler(true, adminAuthToken), orderHandlers.ToggleOrder)

	router.POST("/api/product", authHandler(true, adminAuthToken), productHandlers.CreateProduct)
	router.GET("/api/products", authHandler(false, adminAuthToken), productHandlers.ListProducts)
	router.PUT("/api/product/:productId", authHandler(false, adminAuthToken), productHandlers.UpdateProduct)
	router.DELETE("/api/product/:productId", authHandler(false, adminAuthToken), productHandlers.ToggleDiscontinued)
	router.PUT("/api/product/type", authHandler(true, adminAuthToken), productHandlers.CreateProductType)
	router.GET("/api/product/types", productHandlers.ListProductTypes)

	router.POST("/api/station", authHandler(true, adminAuthToken), stationHandlers.CreateStation)
	router.POST("/api/station/:stationId/:productId", authHandler(true, adminAuthToken), stationHandlers.AddProductToStation)
	router.DELETE("/api/station/:stationId/:productId", authHandler(true, adminAuthToken), stationHandlers.RemoveProductFromStation)
	router.GET("/api/station/:stationId", authHandler(false, adminAuthToken), stationHandlers.Station)
	router.DELETE("/api/station/:stationId", authHandler(true, adminAuthToken), stationHandlers.Delete)
	router.GET("/api/stations", authHandler(false, adminAuthToken), stationHandlers.Stations)

	router.GET("/api/printers", func(c *gin.Context) {
		c.JSON(http.StatusOK, &core.ApiResponse{
			Data:    config.Printers,
			Success: true,
		})
	})

	// The websocket and streaming endpoints.
	router.GET("/api/products/ws", authHandler(false, adminAuthToken), productHandlers.ProductUpdateWS)
	router.GET("/api/products/stream", authHandler(false, adminAuthToken), productHandlers.ProductUpdateStream)
	router.GET("/api/orders/stream", authHandler(false, adminAuthToken), orderHandlers.OrderStream)

	router.Run(":" + strconv.Itoa(config.Server.Port))
}
