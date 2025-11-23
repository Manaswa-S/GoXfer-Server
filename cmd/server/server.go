package server

import (
	"encoding/base64"
	"fmt"
	"goxfer/server/cmd/db"
	"goxfer/server/internal/auth"
	redisauth "goxfer/server/internal/auth/redis"
	handler "goxfer/server/internal/handlers"
	"goxfer/server/internal/middleware"
	"goxfer/server/internal/service"
	localStore "goxfer/server/internal/storage/local"
	transfer "goxfer/server/internal/transfer"
	"log"
	"os"
	"path/filepath"

	"github.com/bytemare/opaque"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func InitHTTPServer(ds *db.DataStore) error {
	router := gin.Default()

	err := initRoutes(router, ds)
	if err != nil {
		return err
	}

	go func() {
		err := router.Run(":" + os.Getenv("PORT"))
		if err != nil {
			fmt.Println(err)
			return
		}
		// TODO: implement gracefull shutdown
	}()

	return nil
}

func initRoutes(router *gin.Engine, ds *db.DataStore) error {
	routerGrp := router.Group("")

	publicGrp := routerGrp.Group("/api/v1/public")
	privateGrp := routerGrp.Group("/api/v1/private")

	auth := newRedisAuth(ds.Redis)

	if err := setMiddlewares(privateGrp, auth); err != nil {
		return err
	}

	opaqueServer := newOpaqueServer()
	upload := transfer.NewUpload()
	builder := newBuilder()
	local, err := newLocalStorage()
	if err != nil {
		return err
	}

	srvc := service.NewService(ds.Queries, ds.Redis, opaqueServer, auth, local, upload, builder)
	hndlr := handler.NewHandler(srvc)
	hndlr.RegisterRoutes(publicGrp, privateGrp)

	return nil
}

func setMiddlewares(routerGrp *gin.RouterGroup, auth auth.Authenticator) error {

	authenticator := middleware.NewAuthenticator(auth)

	routerGrp.Use(authenticator.Authenticator())
	return nil
}

func newRedisAuth(redis *redis.Client) *redisauth.RedisAuth {
	auth := redisauth.NewRedisAuth(redis)
	return auth
}
func newOpaqueServer() *service.Opaque {

	serverID := []byte(loadFromSecrets("SERVER_ID"))
	if serverID == nil {
		log.Fatalln("server identity not found")
	}

	confBytes, err := base64.StdEncoding.DecodeString(loadFromSecrets("DEFAULT_CONF"))
	if err != nil {
		log.Fatalln(err)
	}
	conf, err := opaque.DeserializeConfiguration(confBytes)
	if err != nil {
		log.Fatalln(err)
	}

	secretOprfSeed, err := base64.StdEncoding.DecodeString(loadFromSecrets("OPRF_SEED"))
	if err != nil {
		log.Fatalln(err)
	}

	serverPrivateKey, err := base64.StdEncoding.DecodeString(loadFromSecrets("PRIVATE_KEY"))
	if err != nil {
		log.Fatalln(err)
	}

	serverPublicKey, err := base64.StdEncoding.DecodeString(loadFromSecrets("PUBLIC_KEY"))
	if err != nil {
		log.Fatalln(err)
	}

	if serverPrivateKey == nil || serverPublicKey == nil || secretOprfSeed == nil {
		log.Fatalf("Oh no! Something went wrong setting up the server secrets!")
	}

	server, err := conf.Server()
	if err != nil {
		log.Fatalln(err)
	}

	if err := server.SetKeyMaterial(serverID, serverPrivateKey, serverPublicKey, secretOprfSeed); err != nil {
		log.Fatalln(err)
	}

	opaqueSrvr := service.NewOpaque(server, serverID, serverPrivateKey, serverPublicKey, secretOprfSeed, confBytes)
	return opaqueSrvr
}
func newBuilder() *service.Builder {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}
	// TODO: Replace homeDir with default temp dir i.e. ("")
	tempDir, err := os.MkdirTemp(homeDir, "goxfer_defaulttemp_*")
	if err != nil {
		panic(err)
	}

	return service.NewBuilder(homeDir, tempDir)
}
func newLocalStorage() (*localStore.Local, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	baseDir := filepath.Join(homeDir, "goXfer-base-001")

	tempDir, err := os.MkdirTemp("", "goXfer-temp-*")
	if err != nil {
		return nil, err
	}
	return localStore.NewLocalStorage(
		baseDir,
		tempDir,
	)
}

func loadFromSecrets(name string) string {

	value, exists := os.LookupEnv(name)
	if !exists {
		return ""
	}

	return value
}
