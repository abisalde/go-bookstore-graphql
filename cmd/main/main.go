package main

import (
	"context"
	"log"
	"os"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/lru"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/abisalde/go-bookstore-graphql/config"
	"github.com/abisalde/go-bookstore-graphql/graph"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	_ "github.com/mattn/go-sqlite3"
	"github.com/vektah/gqlparser/v2/ast"
)

const defaultPort = "8080"

func LoggingMiddleware(ctx context.Context, next graphql.Resolver) (res interface{}, err error) {
	rc := graphql.GetFieldContext(ctx)
	log.Printf("GraphQL: %s.%s called", rc.Object, rc.Field.Name)
	return next(ctx)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	config.Connect()
	client := config.GetDB()

	bookRepo := graph.NewResolver(client)

	defer client.Close()

	app := fiber.New(fiber.Config{
		AppName:     "Bookstore",
		ProxyHeader: fiber.HeaderXForwardedFor,
	})

	srv := handler.New(graph.NewExecutableSchema(graph.Config{Resolvers: bookRepo}))

	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.GET{})
	srv.AddTransport(transport.POST{})

	srv.SetQueryCache(lru.New[*ast.QueryDocument](1000))

	srv.Use(extension.Introspection{})
	srv.Use(extension.FixedComplexityLimit(100))
	srv.Use(extension.AutomaticPersistedQuery{
		Cache: lru.New[string](100),
	})

	srv.AroundOperations(func(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
		log.Println("GraphQL operation received")
		return next(ctx)
	})
	// srv.AroundFields(LoggingMiddleware)
	app.Use(logger.New())

	// Use adaptor for GraphQL endpoint
	app.All("/graphql", func(c *fiber.Ctx) error {
		clientIP := c.IP()
		remoteAddr := c.Context().RemoteAddr().String()
		log.Printf("GraphQL request from IP: %s, RemoteAddr: %s", clientIP, remoteAddr)
		return adaptor.HTTPHandler(srv)(c)
	})

	// Use adaptor for Playground
	app.Get("/", adaptor.HTTPHandlerFunc(
		playground.ApolloSandboxHandler("GraphQL playground", "/graphql"),
	))

	log.Printf("🚀 Server ready at http://localhost:%s", port)
	log.Fatal(app.Listen(":" + port))
}
